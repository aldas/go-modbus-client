package modbus_test

import (
	"context"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func xTestReadHoldingRegistersRequestTCP(t *testing.T) {
	client := modbus.NewTCPClient()
	if err := client.Connect(context.Background(), "localhost:5020"); err != nil {
		return
	}
	defer client.Close()
	startAddress := uint16(10)
	req, err := packet.NewReadHoldingRegistersRequestTCP(0, startAddress, 9)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Do(ctx, req)

	assert.NoError(t, err)

	fc3Resp, ok := resp.(*packet.ReadHoldingRegistersResponseTCP)
	assert.True(t, ok)
	assert.Equal(t, uint8(0), fc3Resp.UnitID)
	assert.Equal(t, packet.FunctionReadHoldingRegisters, fc3Resp.FunctionCode())

	register, err := fc3Resp.AsRegisters(startAddress)
	assert.NoError(t, err)

	at19, err := register.Uint16(18)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0), at19)

	at17, err := register.Uint32(17)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), at17)

	at15, err := register.Uint64(15)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), at15)
}

func xTestReadCoilsRequestTCP(t *testing.T) {
	client := modbus.NewTCPClient()
	if err := client.Connect(context.Background(), "localhost:5020"); err != nil {
		return
	}
	defer client.Close()

	req, err := packet.NewReadCoilsRequestTCP(0, 10, 9)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Do(ctx, req)

	assert.NoError(t, err)

	coilResp, ok := resp.(*packet.ReadCoilsResponseTCP)
	assert.True(t, ok)
	assert.Equal(t, uint8(0), coilResp.UnitID)
	assert.Equal(t, packet.FunctionReadCoils, coilResp.FunctionCode())
}
