package packet

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
)

// ReadHoldingRegistersRequestTCP is TCP Request for Read Holding Registers (FC=03)
//
// Example packet: 0x00 0x01 0x00 0x00 0x00 0x06 0x01 0x03 0x00 0x6B 0x00 0x01
// 0x00 0x01 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x01 - unit id (6)
// 0x03 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0x00 0x01 - holding registers quantity to return (10,11)
type ReadHoldingRegistersRequestTCP struct {
	MBAPHeader
	ReadHoldingRegistersRequest
}

// ReadHoldingRegistersRequestRTU is RTU Request for Read Holding Registers (FC=03)
//
// Example packet: 0x01 0x03 0x00 0x6B 0x00 0x01 0xf5 0xd6
// 0x01 - unit id (0)
// 0x03 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x00 0x01 - holding registers quantity to return (4,5)
// 0xf5 0xd6 - CRC16 (6,7)
type ReadHoldingRegistersRequestRTU struct {
	ReadHoldingRegistersRequest
}

// ReadHoldingRegistersRequest is Request for Read Holding Registers (FC=03)
type ReadHoldingRegistersRequest struct {
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16
}

// NewReadHoldingRegistersRequestTCP creates new instance of Read Holding Registers TCP request
func NewReadHoldingRegistersRequestTCP(unitID uint8, startAddress uint16, quantity uint16) (*ReadHoldingRegistersRequestTCP, error) {
	if quantity == 0 || quantity > MaxRegistersInReadResponse {
		return nil, fmt.Errorf("quantity is out of range (1-125): %v", quantity)
	}

	return &ReadHoldingRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 1 + rand.N(uint16(65534)), // #nosec G404
			ProtocolID:    0,
		},
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadHoldingRegistersRequestTCP packet as bytes form
func (r ReadHoldingRegistersRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadHoldingRegistersRequest.bytes(result[6:12])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadHoldingRegistersRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitid + 1 fc + 1 register byte count + N data len (2-256)
	return 6 + 3 + 2*int(r.Quantity)
}

// ParseReadHoldingRegistersRequestTCP parses given bytes into ReadHoldingRegistersRequestTCP
func ParseReadHoldingRegistersRequestTCP(data []byte) (*ReadHoldingRegistersRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	unitID := data[6]
	if data[7] != FunctionReadHoldingRegisters {
		tmpErr := NewErrorParseTCP(ErrIllegalFunction, "received function code in packet is not 0x03")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadHoldingRegisters
		return nil, tmpErr
	}
	quantity := binary.BigEndian.Uint16(data[10:12])
	if !(quantity >= 1 && quantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseTCP(ErrIllegalDataValue, "invalid quantity. valid range 1..125")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadHoldingRegisters
		return nil, tmpErr
	}
	return &ReadHoldingRegistersRequestTCP{
		MBAPHeader: header,
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID: unitID,
			// function code = data[7]
			StartAddress: binary.BigEndian.Uint16(data[8:10]),
			Quantity:     quantity,
		},
	}, nil
}

// NewReadHoldingRegistersRequestRTU creates new instance of Read Holding Registers RTU request
func NewReadHoldingRegistersRequestRTU(unitID uint8, startAddress uint16, quantity uint16) (*ReadHoldingRegistersRequestRTU, error) {
	if quantity == 0 || quantity > MaxRegistersInReadResponse {
		return nil, fmt.Errorf("quantity is out of range (1-125): %v", quantity)
	}

	return &ReadHoldingRegistersRequestRTU{
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadHoldingRegistersRequestRTU packet as bytes form
func (r ReadHoldingRegistersRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.ReadHoldingRegistersRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ParseReadHoldingRegistersRequestRTU parses given bytes into ReadHoldingRegistersRequestRTU
func ParseReadHoldingRegistersRequestRTU(data []byte) (*ReadHoldingRegistersRequestRTU, error) {
	dLen := len(data)
	if dLen != 8 && dLen != 6 { // with or without CRC bytes
		return nil, NewErrorParseRTU(ErrServerFailure, "invalid data length to be valid packet")
	}
	unitID := data[0]
	if data[1] != FunctionReadHoldingRegisters {
		tmpErr := NewErrorParseRTU(ErrIllegalFunction, "received function code in packet is not 0x03")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadHoldingRegisters
		return nil, tmpErr
	}
	quantity := binary.BigEndian.Uint16(data[4:6])
	if !(quantity >= 1 && quantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseRTU(ErrIllegalDataValue, "invalid quantity. valid range 1..125")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadHoldingRegisters
		return nil, tmpErr
	}
	return &ReadHoldingRegistersRequestRTU{
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID: unitID,
			// function code = data[1]
			StartAddress: binary.BigEndian.Uint16(data[2:4]),
			Quantity:     quantity,
		},
	}, nil
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadHoldingRegistersRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 1 register byte count + N register data + 2 crc
	return 3 + 2*int(r.Quantity) + 2
}

// FunctionCode returns function code of this request
func (r ReadHoldingRegistersRequest) FunctionCode() uint8 {
	return FunctionReadHoldingRegisters
}

// Bytes returns ReadHoldingRegistersRequest packet as bytes form
func (r ReadHoldingRegistersRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r ReadHoldingRegistersRequest) bytes(bytes []byte) []byte {
	putReadRequestBytes(bytes, r.UnitID, FunctionReadHoldingRegisters, r.StartAddress, r.Quantity)
	return bytes
}
