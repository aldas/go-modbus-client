package modbus

import (
	"fmt"
	"github.com/aldas/go-modbus-client/packet"
	"sort"
)

// split groups (by host:port+unitID+ "optimized" max amount of fields for max quantity) fields into packets
func split(fields []*Field, funcType string) ([]RegisterRequest, error) {
	byAddressUnitID := groupByAddressUnitID(fields)
	batches := batchGrouped(byAddressUnitID)

	result := make([]RegisterRequest, len(batches))
	for i, b := range batches {
		var req packet.Request
		var err error
		switch funcType {
		case "fc3_tcp":
			req, err = packet.NewReadHoldingRegistersRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
		case "fc3_rtu":
			req, err = packet.NewReadHoldingRegistersRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
		case "fc4_tcp":
			req, err = packet.NewReadInputRegistersRequestTCP(b.UnitID, b.StartAddress, b.Quantity)
		case "fc4_rtu":
			req, err = packet.NewReadInputRegistersRequestRTU(b.UnitID, b.StartAddress, b.Quantity)
		}
		if err != nil {
			return nil, err
		}
		result[i] = RegisterRequest{
			startAddress: b.StartAddress,
			Request:      req,
			fields:       b.fields,
		}
	}
	return result, nil
}

func groupByAddressUnitID(fields []*Field) map[string]map[uint16]registerSlot {
	result := map[string]map[uint16]registerSlot{}
	for _, f := range fields {
		// create groups by modbus address + unitID ... and on second level by register address
		gID := fmt.Sprintf("%v_%v", f.address, f.unitID)
		group, ok := result[gID]
		if !ok {
			group = map[uint16]registerSlot{}
			result[gID] = group
		}

		registerSize := f.registerSize()
		slot, ok := group[f.registerAddress]
		if !ok {
			slot = registerSlot{
				registerAddress: f.registerAddress,
				size:            registerSize,
				fields:          Fields{},
			}
		}
		if registerSize > slot.size {
			slot.size = registerSize
		}
		slot.fields = append(slot.fields, f)
		group[f.registerAddress] = slot
	}
	return result
}

func batchGrouped(grouped map[string]map[uint16]registerSlot) []requestBatch {
	// NB: is batching/grouping algorithm is very naive. It just sorts fields by register and creates N number
	// of requests of them by limiting quantity to MaxRegistersInReadResponse. It does not try to optimise long caps
	// between fields

	var result = make([]requestBatch, 0)
	for _, group := range grouped {
		groupByAddress := slotsSorter{}
		for _, slot := range group {
			groupByAddress = append(groupByAddress, slot)
		}
		sort.Sort(groupByAddress)

		address := groupByAddress[0].fields[0].address
		unitID := groupByAddress[0].fields[0].unitID

		batch := requestBatch{}
		isFirstSeen := false
		var firstAddress uint16
		for _, slot := range groupByAddress {
			registerAddress := slot.registerAddress
			if !isFirstSeen {
				firstAddress = registerAddress
				isFirstSeen = true

				batch.StartAddress = firstAddress
				batch.Address = address
				batch.UnitID = unitID
			}

			addressDiff := (registerAddress + slot.size) - firstAddress
			if addressDiff >= packet.MaxRegistersInReadResponse {
				result = append(result, batch)

				batch = requestBatch{
					Address:      address,
					UnitID:       unitID,
					StartAddress: registerAddress,
				}
				firstAddress = registerAddress
				addressDiff = slot.size
			}
			batch.Quantity = addressDiff

			batch.fields = append(batch.fields, slot.fields...)
		}
		result = append(result, batch)
	}
	return result
}

type registerSlot struct {
	registerAddress uint16
	size            uint16
	fields          Fields
}

type slotsSorter []registerSlot

func (a slotsSorter) Len() int      { return len(a) }
func (a slotsSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a slotsSorter) Less(i, j int) bool {
	return a[i].registerAddress < a[j].registerAddress
}

type requestBatch struct {
	Address      string
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16

	fields Fields
}
