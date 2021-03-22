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
			RegisterAddress: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: "", UnitID: 0, // ServerAddress is empty
			RegisterAddress: 1, Type: FieldTypeInt8,
		},
	}

	batched, err := split(given, "fc3_tcp")
	assert.EqualError(t, err, "field server address can not be empty")
	assert.Nil(t, batched)
}

func TestSplit_single(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 1, Type: FieldTypeInt8,
		},
	}

	batched, err := split(given, "fc3_tcp")
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	pReq, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	pReq.TransactionID = 123
	expect := RegisterRequest{
		serverAddress: ":502",
		startAddress:  1,
		Request:       pReq,
		fields: []Field{
			{
				ServerAddress: ":502", UnitID: 0,
				RegisterAddress: 1, Type: FieldTypeInt8,
			},
		},
	}
	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0])
}

func TestSplit_many(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 118, Length: 11, Type: FieldTypeString, // 118 + 6 + 124
		},
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 121, Type: FieldTypeUint64,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 122, Type: FieldTypeFloat32,
		},
	}

	batched, err := split(given, "fc3_tcp")
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 124)
	expect.TransactionID = 123

	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0].Request)
}

func TestSplit_to2batches(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 119, Length: 15, Type: FieldTypeString, // 119,120,121,122, 123,124,125,126 == new request
		},
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 121, Type: FieldTypeUint64, // 121,122,123,124
		},
		{
			ServerAddress: ":502", UnitID: 0,
			RegisterAddress: 122, Type: FieldTypeFloat32, // 122, 123
		},
	}

	batched, err := split(given, "fc3_tcp")
	assert.NoError(t, err)
	assert.Len(t, batched, 2)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	expect.TransactionID = 123

	firstBatch := batched[0]
	firstBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, firstBatch.Request)
	assert.Len(t, firstBatch.fields, 1)

	expect2, _ := packet.NewReadHoldingRegistersRequestTCP(0, 119, 8)
	expect2.TransactionID = 124

	secondBatch := batched[1]
	secondBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 124
	assert.Equal(t, expect2, secondBatch.Request)
	assert.Len(t, secondBatch.fields, 3)
}
