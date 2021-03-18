package modbus

import (
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplit_single(t *testing.T) {
	given := []*Field{
		{
			address: ":502", unitID: 0,
			registerAddress: 1, fieldType: fieldTypeInt8,
		},
	}

	batched, err := split(given, "fc3_tcp")
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	pReq, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	pReq.TransactionID = 123
	expect := RegisterRequest{
		startAddress: 1,
		Request:      pReq,
		fields: []*Field{
			{
				address: ":502", unitID: 0,
				registerAddress: 1, fieldType: fieldTypeInt8,
			},
		},
	}
	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0])
}

func TestSplit_many(t *testing.T) {
	given := []*Field{
		{
			address: ":502", unitID: 0,
			registerAddress: 1, fieldType: fieldTypeInt8,
		},
		{
			address: ":502", unitID: 0,
			registerAddress: 120, length: 7, fieldType: fieldTypeString,
		},
		{
			address: ":502", unitID: 0,
			registerAddress: 123, fieldType: fieldTypeFloat32,
		},
		{
			address: ":502", unitID: 0,
			registerAddress: 121, fieldType: fieldTypeUint64,
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
	given := []*Field{
		{
			address: ":502", unitID: 0,
			registerAddress: 1, fieldType: fieldTypeInt8,
		},
		{
			address: ":502", unitID: 0,
			registerAddress: 120, length: 7, fieldType: fieldTypeString,
		},
		{
			address: ":502", unitID: 0,
			registerAddress: 123, fieldType: fieldTypeFloat32,
		},
		{
			address: ":502", unitID: 0,
			registerAddress: 122, fieldType: fieldTypeUint64,
		},
	}

	batched, err := split(given, "fc3_tcp")
	assert.NoError(t, err)
	assert.Len(t, batched, 2)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 123)
	expect.TransactionID = 123

	firstBatch := batched[0]
	firstBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, firstBatch.Request)
	assert.Len(t, firstBatch.fields, 2)

	expect2, _ := packet.NewReadHoldingRegistersRequestTCP(0, 122, 3)
	expect2.TransactionID = 124

	secondBatch := batched[1]
	secondBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 124
	assert.Equal(t, expect2, secondBatch.Request)
	assert.Len(t, secondBatch.fields, 2)
}
