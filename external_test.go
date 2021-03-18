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

func TestReadCoilsRequestTCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	addr := startServer(ctx, t)
	//addr := "localhost:5020"

	client := modbus.NewTCPClient()
	if err := client.Connect(context.Background(), addr); err != nil {
		return
	}
	defer client.Close()

	req, err := packet.NewReadCoilsRequestTCP(0, 10, 9)
	//req, err := packet.NewReadDiscreteInputsRequestTCP(0, 10, 9)
	//req, err := packet.NewReadHoldingRegistersRequestTCP(0, 10, 9)
	//req, err := packet.NewReadInputRegistersRequestTCP(0, 10, 9)
	//req, err := packet.NewWriteSingleCoilRequestTCP(0, 10, true)
	//req, err := packet.NewWriteSingleRegisterRequestTCP(0, 10, []byte{0xCA, 0xFE})
	//req, err := packet.NewWriteMultipleCoilsRequestTCP(0, 10, []bool{true, false, true})
	//req, err := packet.NewWriteMultipleRegistersRequestTCP(0, 10, []byte{0xCA, 0xFE, 0xBA, 0xBE})
	assert.NoError(t, err)

	clientCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	resp, err := client.Do(clientCtx, req)

	assert.NoError(t, err)
	assert.Equal(t, packet.FunctionReadHoldingRegisters, resp.FunctionCode())
}

func startServer(ctx context.Context, t *testing.T) string {
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
	return addr
}
