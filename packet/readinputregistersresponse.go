package packet

import (
	"encoding/binary"
	"errors"
)

// ReadInputRegistersResponseTCP is TCP Request for Read Input Registers (FC=04)
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x05 0x01 0x04 0x02 0xCD 0x6B
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x05 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x01 - unit id (6)
// 0x04 - function code (7)
// 0x02 - returned registers byte count (8)
// 0xCD 0x6B - input registers data (1 register) (9,10, ... 2 bytes for each register)
type ReadInputRegistersResponseTCP struct {
	MBAPHeader
	ReadInputRegistersResponse
}

// ReadInputRegistersResponseRTU is RTU Request for Read Input Registers (FC=04)
//
// Example packet: 0x01 0x04 0x02 0xCD 0x6B 0xac 0x4f
// 0x01 - unit id (0)
// 0x04 - function code (1)
// 0x02 - returned registers byte count (2)
// 0xCD 0x6B - input registers data (1 register) (3,4, ... 2 bytes for each register)
// 0xac 0x4f - CRC16 (n-2,n-1)
type ReadInputRegistersResponseRTU struct {
	ReadInputRegistersResponse
}

// ReadInputRegistersResponse is Request for Read Input Registers (FC=04)
type ReadInputRegistersResponse struct {
	UnitID          uint8
	RegisterByteLen uint8
	Data            []byte
}

// Bytes returns ReadInputRegistersResponseTCP packet as bytes form
func (r ReadInputRegistersResponseTCP) Bytes() []byte {
	length := r.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadInputRegistersResponse.bytes(result[6 : 6+length])
	return result
}

// ParseReadInputRegistersResponseTCP parses given bytes into ReadInputRegistersResponseTCP
func ParseReadInputRegistersResponseTCP(data []byte) (*ReadInputRegistersResponseTCP, error) {
	dLen := len(data)
	if dLen < 11 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := int(data[8])
	if dLen != 9+byteLen {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadInputRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		ReadInputRegistersResponse: ReadInputRegistersResponse{
			UnitID: data[6],
			// function code = data[7]
			RegisterByteLen: data[8],
			Data:            data[9 : 9+byteLen],
		},
	}, nil
}

// Bytes returns ReadInputRegistersResponseRTU packet as bytes form
func (r ReadInputRegistersResponseRTU) Bytes() []byte {
	length := r.len() + 2
	result := make([]byte, length)
	bytes := r.ReadInputRegistersResponse.bytes(result)
	crc := CRC16(bytes[:length-2])
	result[length-2] = uint8(crc)
	result[length-1] = uint8(crc >> 8)
	return result
}

// ParseReadInputRegistersResponseRTU parses given bytes into ParseReadInputRegistersResponseRTU
func ParseReadInputRegistersResponseRTU(data []byte) (*ReadInputRegistersResponseRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := int(data[2])
	if dLen != 3+byteLen+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadInputRegistersResponseRTU{
		ReadInputRegistersResponse: ReadInputRegistersResponse{
			UnitID: data[0],
			// function code = data[1]
			RegisterByteLen: data[2],
			Data:            data[3 : 3+byteLen],
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadInputRegistersResponse) FunctionCode() uint8 {
	return FunctionReadInputRegisters
}

func (r ReadInputRegistersResponse) len() uint16 {
	return 3 + uint16(r.RegisterByteLen)
}

// Bytes returns ReadInputRegistersResponse packet as bytes form
func (r ReadInputRegistersResponse) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
}

func (r ReadInputRegistersResponse) bytes(data []byte) []byte {
	data[0] = r.UnitID
	data[1] = FunctionReadInputRegisters
	data[2] = r.RegisterByteLen
	copy(data[3:], r.Data)

	return data
}

// AsRegisters returns response data as Register to more convenient access
func (r ReadInputRegistersResponse) AsRegisters(requestStartAddress uint16) (*Registers, error) {
	return NewRegisters(r.Data, requestStartAddress)
}
