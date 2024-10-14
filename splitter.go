package modbus

import (
	"fmt"
	"github.com/aldas/go-modbus-client/packet"
	"sort"
)

type splitToFuncType uint8

const (
	splitToFC1TCP splitToFuncType = iota
	splitToFC1RTU
	splitToFC2TCP
	splitToFC2RTU
	splitToFC3TCP
	splitToFC3RTU
	splitToFC4TCP
	splitToFC4RTU
)

// split groups (by host:port+UnitID, "optimized" max amount of fields for max quantity) fields into packets
func split(fields []Field, funcType splitToFuncType) ([]BuilderRequest, error) {
	onlyCoils := funcType == splitToFC1TCP || funcType == splitToFC1RTU || funcType == splitToFC2TCP || funcType == splitToFC2RTU
	connectionGroup, err := groupForSingleConnection(fields, onlyCoils)
	if err != nil {
		return nil, err
	}
	batches := batchToRequests(connectionGroup)

	result := make([]BuilderRequest, 0, len(batches))
	for _, b := range batches {
		var req packet.Request
		var err error
		switch funcType {
		case splitToFC1TCP:
			req, err = packet.NewReadCoilsRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
		case splitToFC1RTU:
			req, err = packet.NewReadCoilsRequestRTU(b.UnitID, b.StartAddress, b.Quantity)

		case splitToFC2TCP:
			req, err = packet.NewReadDiscreteInputsRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
		case splitToFC2RTU:
			req, err = packet.NewReadDiscreteInputsRequestRTU(b.UnitID, b.StartAddress, b.Quantity)

		case splitToFC3TCP:
			req, err = packet.NewReadHoldingRegistersRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
		case splitToFC3RTU:
			req, err = packet.NewReadHoldingRegistersRequestRTU(b.UnitID, b.StartAddress, b.Quantity)

		case splitToFC4TCP:
			req, err = packet.NewReadInputRegistersRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
		case splitToFC4RTU:
			req, err = packet.NewReadInputRegistersRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, BuilderRequest{
			Request: req,

			ServerAddress: b.Address,
			UnitID:        b.UnitID,
			StartAddress:  b.StartAddress,
			Fields:        b.fields,
		})
	}
	return result, nil
}

// groupForSingleConnection groups fields into groups what can be requested potentially by same request (same server + unit ID + function)
func groupForSingleConnection(fields []Field, onlyCoils bool) ([]builderSlotGroup, error) {
	groups := map[string]builderSlotGroup{}
	for _, f := range fields {
		if err := f.Validate(); err != nil {
			return nil, err
		}
		// create groups by modbus server Address + unitID + isCoil
		isCoil := f.Type == FieldTypeCoil
		if onlyCoils && !isCoil {
			continue
		} else if !onlyCoils && isCoil {
			continue
		}

		gID := fmt.Sprintf("%v_%v_%v", f.ServerAddress, f.UnitID, isCoil)
		group, ok := groups[gID]
		if !ok {
			group = builderSlotGroup{
				serverAddress: f.ServerAddress,
				unitID:        f.UnitID,
				isForCoils:    isCoil,
				slots:         make([]builderSlot, 0),
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
		address := slotGroup.serverAddress
		unitID := slotGroup.unitID
		addressLimit := packet.MaxRegistersInReadResponse
		if slotGroup.isForCoils {
			addressLimit = packet.MaxCoilsInReadResponse
		}
		sort.Sort(slotsSorter(slotGroup.slots))

		batch := requestBatch{}
		isFirstSeen := false
		var firstAddress uint16
		for _, slot := range slotGroup.slots {
			slotAddress := slot.address
			if !isFirstSeen {
				firstAddress = slotAddress
				isFirstSeen = true

				batch.StartAddress = firstAddress
				batch.Address = address
				batch.UnitID = unitID
			}

			slotEndAddress := slotAddress + slot.size
			addressDiff := slotEndAddress - firstAddress
			if addressDiff > addressLimit {
				result = append(result, batch)

				batch = requestBatch{
					Address:      address,
					UnitID:       unitID,
					StartAddress: slotAddress,
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

type builderSlotGroup struct {
	serverAddress string
	unitID        uint8
	isForCoils    bool

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
	Address      string
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16

	IsForCoils bool

	fields Fields
}
