package packet

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
)

// ReadInputRegistersRequestTCP is TCP Request for Read Input Registers (FC=04)
//
// Example packet: 0x00 0x01 0x00 0x00 0x00 0x06 0x01 0x04 0x00 0x6B 0x00 0x01
// 0x00 0x01 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x01 - unit id (6)
// 0x04 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0x00 0x01 - input registers quantity to return (10,11)
type ReadInputRegistersRequestTCP struct {
	MBAPHeader
	ReadInputRegistersRequest
}

// ReadInputRegistersRequestRTU is RTU Request for Read Input Registers (FC=04)
//
// Example packet: 0x01 0x04 0x00 0x6B 0x00 0x01 0x40 0x16
// 0x01 - unit id (0)
// 0x04 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x00 0x01 - input registers quantity to return (4,5)
// 0x40 0x16 - CRC16 (6,7)
type ReadInputRegistersRequestRTU struct {
	ReadInputRegistersRequest
}

// ReadInputRegistersRequest is Request for Read Input Registers (FC=04)
type ReadInputRegistersRequest struct {
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16
}

// NewReadInputRegistersRequestTCP creates new instance of Read Input Registers TCP request
func NewReadInputRegistersRequestTCP(unitID uint8, startAddress uint16, quantity uint16) (*ReadInputRegistersRequestTCP, error) {
	if quantity == 0 || quantity > MaxRegistersInReadResponse {
		return nil, fmt.Errorf("quantity is out of range (1-125): %v", quantity)
	}

	return &ReadInputRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 1 + rand.N(uint16(65534)), // #nosec G404
			ProtocolID:    0,
		},
		ReadInputRegistersRequest: ReadInputRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadInputRegistersRequestTCP packet as bytes form
func (r ReadInputRegistersRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadInputRegistersRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadInputRegistersRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitid + 1 fc + 1 register byte count + N data len (2-256)
	return 6 + 3 + 2*int(r.Quantity)
}

// ParseReadInputRegistersRequestTCP parses given bytes into ReadInputRegistersRequestTCP
func ParseReadInputRegistersRequestTCP(data []byte) (*ReadInputRegistersRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	unitID := data[6]
	if data[7] != FunctionReadInputRegisters {
		tmpErr := NewErrorParseTCP(ErrIllegalFunction, "received function code in packet is not 0x04")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadInputRegisters
		return nil, tmpErr
	}
	quantity := binary.BigEndian.Uint16(data[10:12])
	if !(quantity >= 1 && quantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseTCP(ErrIllegalDataValue, "invalid quantity. valid range 1..125")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadInputRegisters
		return nil, tmpErr
	}
	return &ReadInputRegistersRequestTCP{
		MBAPHeader: header,
		ReadInputRegistersRequest: ReadInputRegistersRequest{
			UnitID: unitID,
			// function code = data[7]
			StartAddress: binary.BigEndian.Uint16(data[8:10]),
			Quantity:     quantity,
		},
	}, nil
}

// NewReadInputRegistersRequestRTU creates new instance of Read Input Registers RTU request
func NewReadInputRegistersRequestRTU(unitID uint8, startAddress uint16, quantity uint16) (*ReadInputRegistersRequestRTU, error) {
	if quantity == 0 || quantity > MaxRegistersInReadResponse {
		return nil, fmt.Errorf("quantity is out of range (1-125): %v", quantity)
	}

	return &ReadInputRegistersRequestRTU{
		ReadInputRegistersRequest: ReadInputRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadInputRegistersRequestRTU packet as bytes form
func (r ReadInputRegistersRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.ReadInputRegistersRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ParseReadInputRegistersRequestRTU parses given bytes into ReadInputRegistersRequestRTU
func ParseReadInputRegistersRequestRTU(data []byte) (*ReadInputRegistersRequestRTU, error) {
	dLen := len(data)
	if dLen != 8 && dLen != 6 { // with or without CRC bytes
		return nil, NewErrorParseRTU(ErrServerFailure, "invalid data length to be valid packet")
	}
	unitID := data[0]
	if data[1] != FunctionReadInputRegisters {
		tmpErr := NewErrorParseRTU(ErrIllegalFunction, "received function code in packet is not 0x04")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadInputRegisters
		return nil, tmpErr
	}
	quantity := binary.BigEndian.Uint16(data[4:6])
	if !(quantity >= 1 && quantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseRTU(ErrIllegalDataValue, "invalid quantity. valid range 1..125")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadInputRegisters
		return nil, tmpErr
	}
	return &ReadInputRegistersRequestRTU{
		ReadInputRegistersRequest: ReadInputRegistersRequest{
			UnitID: unitID,
			// function code = data[1]
			StartAddress: binary.BigEndian.Uint16(data[2:4]),
			Quantity:     quantity,
		},
	}, nil
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadInputRegistersRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 1 register byte count + N register data + 2 CRC
	return 3 + 2*int(r.Quantity) + 2
}

// FunctionCode returns function code of this request
func (r ReadInputRegistersRequest) FunctionCode() uint8 {
	return FunctionReadInputRegisters
}

// Bytes returns ReadInputRegistersRequest packet as bytes form
func (r ReadInputRegistersRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r ReadInputRegistersRequest) bytes(bytes []byte) []byte {
	putReadRequestBytes(bytes, r.UnitID, FunctionReadInputRegisters, r.StartAddress, r.Quantity)
	return bytes
}
