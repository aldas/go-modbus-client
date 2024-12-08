package modbus

import (
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplit_validationError(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: "", UnitID: 0, // ServerAddress is empty
			Address: 1, Type: FieldTypeInt8,
		},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.EqualError(t, err, "field server address can not be empty")
	assert.Nil(t, batched)
}

func TestSplit_single(t *testing.T) {
	given := []Field{
		{ServerAddress: ":502", UnitID: 0, Address: 1, Type: FieldTypeInt8},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	pReq, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	pReq.TransactionID = 123
	expect := BuilderRequest{
		ServerAddress: ":502",
		StartAddress:  1,
		Protocol:      ProtocolTCP,
		Request:       pReq,
		Fields: []Field{
			{ServerAddress: ":502", UnitID: 0, Address: 1, Type: FieldTypeInt8, FunctionCode: 3, Protocol: ProtocolTCP},
		},
	}
	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0])
}

func TestSplit_quantity(t *testing.T) {
	var testCases = []struct {
		name               string
		givenFields        []Field
		expectStartAddress uint16
		expectQuantity     uint16
	}{
		{
			name: "ok, int64",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 10, Type: FieldTypeFloat64},
			},
			expectStartAddress: 10,
			expectQuantity:     4,
		},
		{
			name: "ok, multiple fields",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 10, Type: FieldTypeInt8},
				{ServerAddress: ":502", UnitID: 0, Address: 74, Type: FieldTypeUint32},
			},
			expectStartAddress: 10,
			expectQuantity:     66,
		},
		{
			name: "ok, multiple fields int8",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 10, Type: FieldTypeInt8},
				{ServerAddress: ":502", UnitID: 0, Address: 75, Type: FieldTypeInt8},
			},
			expectStartAddress: 10,
			expectQuantity:     66,
		},
		{
			name: "ok, multiple fields int64",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 0, Type: FieldTypeInt8},
				{ServerAddress: ":502", UnitID: 0, Address: 1, Type: FieldTypeFloat64},
			},
			expectStartAddress: 0,
			expectQuantity:     5,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			batched, err := split(tc.givenFields, 3, ProtocolTCP)
			if assert.NoError(t, err) {
				assert.Len(t, batched, 1)
				req := batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP)
				assert.Equal(t, tc.expectStartAddress, req.StartAddress)
				assert.Equal(t, tc.expectQuantity, req.Quantity)
			}
		})
	}
}

func TestSplit_many(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 118, Length: 11, Type: FieldTypeString, // 118 + 6 + 124
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 121, Type: FieldTypeUint64,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeInt16,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeFloat32,
		},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 124)
	expect.TransactionID = 123

	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0].Request)
}

func TestSplit_to2RegisterBatches(t *testing.T) {
	given := []Field{
		{Name: "F1",
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeInt8,
		},
		{Name: "F2",
			ServerAddress: ":502", UnitID: 0,
			Address: 119, Length: 15, Type: FieldTypeString, // 119,120,121,122, 123,124,125,126 == new request
		},
		{Name: "F3",
			ServerAddress: ":502", UnitID: 0,
			Address: 121, Type: FieldTypeUint64, // 121,122,123,124
		},
		{Name: "F4",
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeFloat32, // 122, 123
		},
		{Name: "F5",
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeCoil, // should be ignored
		},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 2)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	expect.TransactionID = 123

	firstBatch := batched[0]
	firstBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, firstBatch.Request)
	assert.Len(t, firstBatch.Fields, 1)

	expect2, _ := packet.NewReadHoldingRegistersRequestTCP(0, 119, 8)
	expect2.TransactionID = 124

	secondBatch := batched[1]
	secondBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 124
	assert.Equal(t, expect2, secondBatch.Request)
	assert.Len(t, secondBatch.Fields, 3)
}

func TestSplit_to2CoilsBatches(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeCoil,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeCoil, // at same place previous field
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 100, Type: FieldTypeCoil,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 2001, Type: FieldTypeCoil, // should go to next batch
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeFloat32, // should be ignored
		},
	}

	batched, err := split(given, 1, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 2)

	expect, _ := packet.NewReadCoilsRequestTCP(0, 1, 100)
	expect.TransactionID = 123

	firstBatch := batched[0]
	firstBatch.Request.(*packet.ReadCoilsRequestTCP).TransactionID = 123
	assert.Equal(t, expect, firstBatch.Request)
	assert.Len(t, firstBatch.Fields, 3)

	expect2, _ := packet.NewReadCoilsRequestTCP(0, 2001, 1)
	expect2.TransactionID = 124

	secondBatch := batched[1]
	secondBatch.Request.(*packet.ReadCoilsRequestTCP).TransactionID = 124
	assert.Equal(t, expect2, secondBatch.Request)
	assert.Len(t, secondBatch.Fields, 1)
}
