package packet

import (
	"encoding/binary"
	"errors"
)

// ReadHoldingRegistersResponseTCP is TCP Request for Read Holding Registers (FC=03)
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x05 0x01 0x03 0x02 0xCD 0x6B
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x05 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x01 - unit id (6)
// 0x03 - function code (7)
// 0x02 - returned registers byte count (8)
// 0xCD 0x6B - holding registers data (1 register) (9,10, ... 2 bytes for each register)
type ReadHoldingRegistersResponseTCP struct {
	MBAPHeader
	ReadHoldingRegistersResponse
}

// ReadHoldingRegistersResponseRTU is RTU Request for Read Holding Registers (FC=03)
//
// Example packet: 0x01 0x03 0x02 0xCD 0x6B 0xFF 0xFF
// 0x01 - unit id (0)
// 0x03 - function code (1)
// 0x02 - returned registers byte count (2)
// 0xCD 0x6B - holding registers data (1 register) (3,4, ... 2 bytes for each register)
// 0xFF 0xFF - CRC16 (n-2,n-1) // FIXME: add correct crc value example
type ReadHoldingRegistersResponseRTU struct {
	ReadHoldingRegistersResponse
}

// ReadHoldingRegistersResponse is Request for Read Holding Registers (FC=03)
type ReadHoldingRegistersResponse struct {
	UnitID          uint8
	RegisterByteLen uint8
	Data            []byte
}

// Bytes returns ReadHoldingRegistersResponseTCP packet as bytes form
func (r ReadHoldingRegistersResponseTCP) Bytes() []byte {
	length := r.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadHoldingRegistersResponse.bytes(result[6 : 6+length])
	return result
}

// ParseReadHoldingRegistersResponseTCP parses given bytes into ReadHoldingRegistersResponseTCP
func ParseReadHoldingRegistersResponseTCP(data []byte) (*ReadHoldingRegistersResponseTCP, error) {
	dLen := len(data)
	if dLen < 11 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := int(data[8])
	if dLen != 9+byteLen {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadHoldingRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
			UnitID: data[6],
			// function code = data[7]
			RegisterByteLen: data[8],
			Data:            data[9 : 9+byteLen],
		},
	}, nil
}

// Bytes returns ReadHoldingRegistersResponseRTU packet as bytes form
func (r ReadHoldingRegistersResponseRTU) Bytes() []byte {
	length := r.len()
	result := make([]byte, length+2)
	bytes := r.ReadHoldingRegistersResponse.bytes(result)
	crc := CRC16(bytes[:length])
	result[length] = uint8(crc)
	result[length+1] = uint8(crc >> 8)
	return result
}

// ParseReadHoldingRegistersResponseRTU parses given bytes into ReadHoldingRegistersResponseTCP
func ParseReadHoldingRegistersResponseRTU(data []byte) (*ReadHoldingRegistersResponseRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := int(data[2])
	if dLen != 3+byteLen+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadHoldingRegistersResponseRTU{
		ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
			UnitID: data[0],
			// function code = data[1]
			RegisterByteLen: data[2],
			Data:            data[3 : 3+byteLen],
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadHoldingRegistersResponse) FunctionCode() uint8 {
	return FunctionReadHoldingRegisters
}

func (r ReadHoldingRegistersResponse) len() uint16 {
	return 3 + uint16(r.RegisterByteLen)
}

// Bytes returns ReadHoldingRegistersResponse packet as bytes form
func (r ReadHoldingRegistersResponse) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
}

func (r ReadHoldingRegistersResponse) bytes(data []byte) []byte {
	data[0] = r.UnitID
	data[1] = FunctionReadHoldingRegisters
	data[2] = r.RegisterByteLen
	copy(data[3:], r.Data)

	return data
}

// AsRegisters returns response data as Register to more convenient access
func (r ReadHoldingRegistersResponse) AsRegisters(requestStartAddress uint16) (*Registers, error) {
	return NewRegisters(r.Data, requestStartAddress)
}
