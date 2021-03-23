package modbus_test

import (
	"context"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/modbustest"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExternalUsage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	handler := func(received []byte, bytesRead int) (response []byte, closeConnection bool) {
		if bytesRead == 0 {
			return nil, false
		}
		resp := packet.ReadHoldingRegistersResponseTCP{
			MBAPHeader: packet.MBAPHeader{TransactionID: 123, ProtocolID: 0},
			ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
				UnitID:          0,
				RegisterByteLen: 10,
				Data:            []byte{0x0, 0x1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := modbus.NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Uint16(18).UnitID(0).Name("test_do")).
		Add(b.Int64(19).Name("alarm_do_1").UnitID(0)).
		ReadHoldingRegistersTCP()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := modbus.NewClient()
	if err := client.Connect(context.Background(), addr); err != nil {
		return
	}

	//for _, req := range reqs {
	//
	//}
	req := reqs[0] // skip looping as we always have 1 request in this example
	resp, err := client.Do(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	fields, err := req.ExtractFields(resp.(modbus.RegistersResponse), true)
	assert.NotNil(t, resp)
	assert.Len(t, fields, 2)

	assert.Equal(t, uint16(1), fields[0].Value)
	assert.Equal(t, "test_do", fields[0].Field.Name)

	assert.Equal(t, int64(-1), fields[1].Value)
	assert.Equal(t, "alarm_do_1", fields[1].Field.Name)
}
