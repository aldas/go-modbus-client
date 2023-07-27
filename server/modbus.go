package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
)

// ModbusTCPAssembler assembles read data into complete packets and calls ModbusHandler with assembled packet
type ModbusTCPAssembler struct {
	handler  ModbusHandler
	received bytes.Buffer
}

// ReceiveRead assembles read byte until full TCP packet is formed or return an error when received data does not look like TCP packet
func (m *ModbusTCPAssembler) ReceiveRead(ctx context.Context, received []byte, bytesRead int) (response []byte, closeConnection bool) {
	m.received.Write(received)

	n, err := packet.LooksLikeModbusTCP(m.received.Bytes(), false)
	if err == packet.ErrTCPDataTooShort {
		return nil, false // wait for more data to arrive
	} else if err != nil {
		return err.(*packet.ErrorParseTCP).Bytes(), false
	}

	p, err := packet.ParseTCPRequest(m.received.Next(n))
	if err != nil {
		return err.(*packet.ErrorParseTCP).Bytes(), false
	}

	resp, err := m.handler.Handle(ctx, p)
	if err != nil {
		var target *packet.ErrorParseTCP
		if errors.As(err, &target) {
			return target.Bytes(), false
		}
		return packet.NewErrorParseTCP(packet.ErrUnknown, err.Error()).Bytes(), false
	}

	return resp.Bytes(), false
}
