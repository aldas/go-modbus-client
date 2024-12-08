package modbus

import (
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"sort"
	"time"
)

// split groups (by host:port+UnitID, "optimized" max amount of fields for max quantity) fields into packets
func split(fields []Field, functionCode uint8, protocol ProtocolType) ([]BuilderRequest, error) {
	connectionGroup, err := groupForSingleConnection(fields, functionCode, protocol)
	if err != nil {
		return nil, err
	}
	batches := batchToRequests(connectionGroup)

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
		result = append(result, BuilderRequest{
			Request: req,

			ServerAddress:   b.ServerAddress,
			UnitID:          b.UnitID,
			StartAddress:    b.StartAddress,
			Protocol:        b.Protocol,
			RequestInterval: time.Duration(b.RequestInterval),

			Fields: b.fields,
		})
	}
	if len(result) == 0 {
		return nil, errors.New("splitting resulted 0 requests")
	}
	return result, nil
}

// groupForSingleConnection groups fields into groups what can be requested potentially by same request
// (same server + function + unit ID + protocol + interval)
func groupForSingleConnection(fields []Field, functionCode uint8, protocol ProtocolType) ([]builderSlotGroup, error) {
	onlyCoils := functionCode == packet.FunctionReadCoils || functionCode == packet.FunctionReadDiscreteInputs

	groups := map[groupID]builderSlotGroup{}
	for _, f := range fields {
		// adjust field fc and protocol to any cases
		isCoil := f.Type == FieldTypeCoil
		if (!isCoil && functionCode != 0 && f.FunctionCode == 0) || (isCoil && onlyCoils) {
			f.FunctionCode = functionCode
		}
		if protocol != protocolAny && f.Protocol == protocolAny {
			f.Protocol = protocol
		}

		if err := f.Validate(); err != nil {
			return nil, err
		}

		if !onlyCoils && functionCode != 0 && functionCode != f.FunctionCode {
			// when functionCode is provided and field does not have it set - consider field included
			continue
		}
		if protocol != protocolAny && f.Protocol != protocolAny && protocol != f.Protocol {
			// when protocol is provided and field does not have it set - consider field included
			continue
		}

		if onlyCoils && !isCoil {
			continue
		} else if !onlyCoils && isCoil {
			continue
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

func batchToRequests(connectionGroup []builderSlotGroup) []requestBatch {
	// Coils are always grouped to separate requests (fc1/fc2) from fields suitable for registers (fc3/fc4)
	//
	// NB: is batching/grouping algorithm is very naive. It just sorts fields by register and creates N number
	// of requests of them by limiting quantity to MaxRegistersInReadResponse. It does not try to optimise long caps
	// between fields
	// assumes that UnitID is same for all fields within group

	var result = make([]requestBatch, 0)
	for _, slotGroup := range connectionGroup {
		if len(slotGroup.slots) == 0 {
			continue
		}
		addressLimit := packet.MaxRegistersInReadResponse
		if slotGroup.isForCoils {
			addressLimit = packet.MaxCoilsInReadResponse
		}
		sort.Sort(slotsSorter(slotGroup.slots))

		firstAddress := slotGroup.slots[0].address
		batch := requestBatch{
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
			if addressDiff > addressLimit {
				result = append(result, batch)

				batch = requestBatch{
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

			batch.fields = append(batch.fields, slot.fields...)
		}
		result = append(result, batch)
	}
	return result
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

type slotsSorter builderSlots

func (a slotsSorter) Len() int      { return len(a) }
func (a slotsSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a slotsSorter) Less(i, j int) bool {
	return a[i].address < a[j].address
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

type requestBatch struct {
	ServerAddress string
	FunctionCode  uint8
	UnitID        uint8
	Protocol      ProtocolType

	StartAddress uint16
	Quantity     uint16

	IsForCoils      bool
	RequestInterval Duration

	fields Fields
}
