package modbus

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aldas/go-modbus-client/packet"
)

// Splitter groups Fields into Modbus requests
type Splitter interface {
	Split(fields Fields, functionCode uint8, protocol ProtocolType) ([]BuilderRequest, error)
}

// DefaultSplitter implements splitter logic that groups Fields by host:port+UnitID and splits them to BuilderRequest
// "optimized" max amount of Fields for max number of fields as Modbus request size can contain
type DefaultSplitter struct {
	// ShouldSplitFunc is called for field addresses. If it returns true, the current address will be split into a
	// new batch (request) and not added to the current batch.
	ShouldSplitFunc func(batch SplitBatch, address uint16, size uint16) bool
}

// Split groups (by host:port+UnitID, "optimized" max amount of Fields for max quantity) Fields into packets
func (s DefaultSplitter) Split(fields Fields, functionCode uint8, protocol ProtocolType) ([]BuilderRequest, error) {
	connectionGroup, err := groupForSingleConnection(fields, functionCode, protocol)
	if err != nil {
		return nil, err
	}
	batches, err := splitGroupToBatches(connectionGroup, s.ShouldSplitFunc)
	if err != nil {
		return nil, err
	}

	result := make([]BuilderRequest, 0, len(batches))
	for _, b := range batches {
		if b.Protocol == protocolAny || b.FunctionCode == 0 {
			continue
		}
		var req packet.Request
		var err error
		switch b.FunctionCode {
		case packet.FunctionReadCoils: // fc1
			switch b.Protocol {
			case ProtocolTCP:
				req, err = packet.NewReadCoilsRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
			case ProtocolRTU:
				req, err = packet.NewReadCoilsRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
			}

		case packet.FunctionReadDiscreteInputs: // fc2
			switch b.Protocol {
			case ProtocolTCP:
				req, err = packet.NewReadDiscreteInputsRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
			case ProtocolRTU:
				req, err = packet.NewReadDiscreteInputsRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
			}

		case packet.FunctionReadHoldingRegisters: // fc3
			switch b.Protocol {
			case ProtocolTCP:
				req, err = packet.NewReadHoldingRegistersRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
			case ProtocolRTU:
				req, err = packet.NewReadHoldingRegistersRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
			}

		case packet.FunctionReadInputRegisters: // fc4
			switch b.Protocol {
			case ProtocolTCP:
				req, err = packet.NewReadInputRegistersRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
			case ProtocolRTU:
				req, err = packet.NewReadInputRegistersRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
			}
		}
		if err != nil {
			return nil, err
		}
		var quantity uint16
		switch r := req.(type) {
		case *packet.ReadCoilsRequestTCP: // fc1
			quantity = r.Quantity
		case *packet.ReadCoilsRequestRTU: // fc1
			quantity = r.Quantity
		case *packet.ReadDiscreteInputsRequestTCP: // fc2
			quantity = r.Quantity
		case *packet.ReadDiscreteInputsRequestRTU: // fc2
			quantity = r.Quantity
		case *packet.ReadHoldingRegistersRequestTCP: // fc3
			quantity = r.Quantity
		case *packet.ReadHoldingRegistersRequestRTU: // fc3
			quantity = r.Quantity
		case *packet.ReadInputRegistersRequestTCP: // fc4
			quantity = r.Quantity
		case *packet.ReadInputRegistersRequestRTU: // fc4
			quantity = r.Quantity
		}

		result = append(result, BuilderRequest{
			Request: req,

			ServerAddress: b.ServerAddress,
			UnitID:        b.UnitID,
			StartAddress:  b.StartAddress,
			Quantity:      quantity,

			Protocol:        b.Protocol,
			RequestInterval: time.Duration(b.RequestInterval),

			Fields: b.Fields,
		})
	}
	if len(result) == 0 {
		return nil, errors.New("splitting resulted 0 requests")
	}
	return result, nil
}

// groupForSingleConnection groups Fields into groups what can be requested potentially by same request
// (same server + function + unit ID + protocol + interval)
func groupForSingleConnection(fields []Field, functionCode uint8, protocol ProtocolType) ([]builderSlotGroup, error) {
	//onlyCoils := functionCode == packet.FunctionReadCoils || functionCode == packet.FunctionReadDiscreteInputs

	groups := map[groupID]builderSlotGroup{}
	for _, f := range fields {
		// adjust field fc and protocol to any cases
		isCoil := f.Type == FieldTypeCoil
		if f.FunctionCode == 0 && functionCode != 0 {
			f.FunctionCode = functionCode
		}

		if protocol != protocolAny && f.Protocol == protocolAny {
			f.Protocol = protocol
		}

		if functionCode != 0 { // we only want want specific FCs
			if functionCode != f.FunctionCode {
				// when functionCode is provided and field does not have it - ignore it
				continue
			}
		}
		if protocol != protocolAny && f.Protocol != protocolAny && protocol != f.Protocol {
			// when protocol is provided and field does not have it set - consider field included
			continue
		}

		if err := f.Validate(); err != nil {
			return nil, err
		}

		// create groups by modbus server ServerAddress + fc + unitID + isCoil
		gID := groupID{
			serverAddress: f.ServerAddress,
			functionCode:  f.FunctionCode,
			unitID:        f.UnitID,
			protocol:      f.Protocol,
			interval:      f.RequestInterval,
		}
		group, ok := groups[gID]
		if !ok {
			group = builderSlotGroup{
				group:      gID,
				isForCoils: isCoil,
				slots:      make([]builderSlot, 0),
			}
			groups[gID] = group
		}

		group.AddField(f)
		groups[gID] = group
	}
	// TODO: as of Go 1.23 this could be shortened to `slices.Collect(maps.Values(groups))`
	result := make([]builderSlotGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, g)
	}
	return result, nil
}

func splitGroupToBatches(slotGroups []builderSlotGroup, shouldSplitFunc func(batch SplitBatch, address uint16, size uint16) bool) ([]SplitBatch, error) {
	// Coils are always grouped to separate requests (fc1/fc2) from fields suitable for registers (fc3/fc4)
	//
	// NB: this batching/grouping algorithm is very naive. It just sorts fields by register and creates N number
	// of requests of them by limiting the quantity to MaxRegistersInReadResponse. It does not try to optimize long caps
	// between fields.
	// assumes that UnitID is the same for all fields within a group

	var result = make([]SplitBatch, 0)
	for _, slotGroup := range slotGroups {
		if len(slotGroup.slots) == 0 {
			continue
		}
		config, err := addressToSplitterConfig(slotGroup.group.serverAddress)
		if err != nil {
			return nil, err
		}
		addressLimit := packet.MaxRegistersInReadResponse
		if slotGroup.isForCoils {
			addressLimit = packet.MaxCoilsInReadResponse
		}
		if config.MaxQuantityPerRequest > 0 && config.MaxQuantityPerRequest < addressLimit {
			addressLimit = config.MaxQuantityPerRequest
		}

		sort.Slice(slotGroup.slots, func(i, j int) bool {
			return slotGroup.slots[i].address < slotGroup.slots[j].address
		})

		firstAddress := slotGroup.slots[0].address
		batch := SplitBatch{
			ServerAddress:   slotGroup.group.serverAddress,
			FunctionCode:    slotGroup.group.functionCode,
			UnitID:          slotGroup.group.unitID,
			Protocol:        slotGroup.group.protocol,
			RequestInterval: slotGroup.group.interval,

			StartAddress: firstAddress,
			Quantity:     slotGroup.slots[0].size,
		}
		for _, slot := range slotGroup.slots {
			slotAddress := slot.address
			slotEndAddress := slotAddress + slot.size
			addressDiff := slotEndAddress - firstAddress
			isBiggerThanAddressLimit := addressDiff > addressLimit
			isInvalidRangeOverlap, err := config.InvalidRange.Overlaps(firstAddress, slotAddress, slotEndAddress-1)
			if err != nil {
				return nil, err
			}
			if isBiggerThanAddressLimit || isInvalidRangeOverlap || (shouldSplitFunc != nil && shouldSplitFunc(batch, slotAddress, slot.size)) {
				result = append(result, batch)

				batch = SplitBatch{
					ServerAddress:   slotGroup.group.serverAddress,
					FunctionCode:    slotGroup.group.functionCode,
					UnitID:          slotGroup.group.unitID,
					Protocol:        slotGroup.group.protocol,
					RequestInterval: slotGroup.group.interval,

					StartAddress: slotAddress,
					Quantity:     slot.size,
				}
				firstAddress = slotAddress
				addressDiff = slot.size
			}
			if batch.Quantity < addressDiff {
				batch.Quantity = addressDiff
			}

			batch.Fields = append(batch.Fields, slot.fields...)
		}
		result = append(result, batch)
	}
	return result, nil
}

type builderSlot struct {
	address uint16
	size    uint16
	fields  Fields
}

type builderSlots []builderSlot

func (bs *builderSlots) IndexOf(address uint16) int {
	for i, b := range *bs {
		if b.address == address {
			return i
		}
	}
	return -1
}

type groupID struct {
	serverAddress string
	functionCode  uint8
	unitID        uint8
	protocol      ProtocolType
	interval      Duration
}

type builderSlotGroup struct {
	group groupID

	isForCoils bool

	slots builderSlots
}

func (g *builderSlotGroup) AddField(f Field) {
	registerSize := f.registerSize()
	i := g.slots.IndexOf(f.Address)
	if i == -1 {
		g.slots = append(g.slots, builderSlot{
			address: f.Address,
			size:    registerSize,
			fields:  Fields{f},
		})
		return
	}

	slot := g.slots[i]

	slot.fields = append(slot.fields, f)
	if registerSize > slot.size {
		slot.size = registerSize
	}
	g.slots[i] = slot
}

// SplitBatch represents a group of fields that are requested to the same Modbus request
type SplitBatch struct {
	ServerAddress string
	FunctionCode  uint8
	UnitID        uint8
	Protocol      ProtocolType

	StartAddress uint16
	Quantity     uint16

	IsForCoils      bool
	RequestInterval Duration

	Fields Fields
}

// invalidAddress is (register/coil) range in between modbus server should not be requested
type invalidAddress struct {
	from uint16
	to   uint16
}

type invalidRange []invalidAddress

func (r *invalidRange) Overlaps(requestStart uint16, slotStart uint16, slotEnd uint16) (bool, error) {
	if r == nil || len(*r) == 0 {
		return false, nil
	}
	for _, ir := range *r {
		if ir.from <= slotEnd && slotStart <= ir.to {
			return true, fmt.Errorf("field overlaps invalid address range, addr: %d, range: %d-%d", slotStart, ir.from, ir.to)
		}
		if ir.from <= slotEnd && requestStart <= ir.to {
			return true, nil
		}
	}
	return false, nil
}

type splitterConfig struct {
	MaxQuantityPerRequest uint16
	InvalidRange          invalidRange
}

func addressToSplitterConfig(address string) (splitterConfig, error) {
	if !strings.ContainsRune(address, '?') {
		return splitterConfig{}, nil
	}
	url, err := url.Parse(address)
	if err != nil {
		return splitterConfig{}, fmt.Errorf("failed to parse server address for invalid range: %q, err: %w", address, err)
	}
	invalid, err := addressToInvalidRange(url)
	if err != nil {
		return splitterConfig{}, err
	}

	maxQuantityPerRequest := uint16(0)
	if raw := url.Query().Get("max_quantity_per_request"); raw != "" {
		tmpQ, err := strconv.ParseUint(raw, 10, 16)
		if err != nil {
			return splitterConfig{}, fmt.Errorf("failed to parse max_quantity_per_request, err: %w", err)
		}
		maxQuantityPerRequest = uint16(tmpQ)
	}

	return splitterConfig{
		MaxQuantityPerRequest: maxQuantityPerRequest,
		InvalidRange:          invalid,
	}, nil
}

func addressToInvalidRange(url *url.URL) (invalidRange, error) {
	result := make(invalidRange, 0)
	raw := url.Query()["invalid_addr"]
	for _, addrParam := range raw {
		for _, addr := range strings.Split(addrParam, ",") {
			before, after, found := strings.Cut(addr, "-")
			if !found {
				before = addr
			}
			from, err := strconv.ParseUint(before, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("failed to parse invalid range: %q, err: %w", addr, err)
			}
			tmp := invalidAddress{
				from: uint16(from),
				to:   uint16(from),
			}
			if found {
				to, err := strconv.ParseUint(after, 10, 16)
				if err != nil {
					return nil, fmt.Errorf("failed to parse invalid range: %q, err: %w", addr, err)
				}
				tmp.to = uint16(to)
			}
			result = append(result, tmp)
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}
