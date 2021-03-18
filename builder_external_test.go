package modbus_test

import (
	"context"
	"fmt"
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
				RegisterByteLen: 2,
				Data:            []byte{0xca, 0xfe},
			},
		}
		return resp.Bytes(), true
	}
	addr, err := modbustest.RunServerOnRandomPort(ctx, handler)
	if err != nil {
		t.Fatal(err)
	}

	b := modbus.NewRequestBuilder(addr, 1)

	reqs, err := b.Add(b.Int64(18).UnitID(0).Name("test_do")).
		Add(b.Int64(18).Name("alarm_do_1").UnitID(0)).
		ReadHoldingRegistersTCP()
	assert.NoError(t, err)
	assert.Len(t, reqs, 1)

	client := modbus.NewClient()
	if err := client.Connect(context.Background(), addr); err != nil {
		return
	}
	for _, req := range reqs {
		resp, err := client.Do(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	}
}

func TestExternalUsage2(t *testing.T) {
	_ = ExternalUsage2()
}

func ExternalUsage2() error {
	b := modbus.NewRequestBuilder("localhost:5020", 1)

	requests, _ := b.Add(b.Int64(18).UnitID(0).Name("test_do")).
		Add(b.Int64(18).Name("alarm_do_1").UnitID(0)).
		ReadHoldingRegistersTCP() // split added fields into multiple requests with suitable quantity size

	client := modbus.NewClient()
	if err := client.Connect(context.Background(), "localhost:5020"); err != nil {
		return err
	}
	for _, req := range requests {
		resp, err := client.Do(context.Background(), req)
		if err != nil {
			return err
		}
		// extract response as packet.Registers instance to have access to convenience methods to extracting registers
		// as different data types
		registers, _ := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(req.StartAddress())
		alarmDo1, _ := registers.Int64(18)
		fmt.Printf("int64 @ address 18: %v", alarmDo1)
	}

	return nil
}
