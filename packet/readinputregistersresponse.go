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
// Example packet: 0x01 0x04 0x02 0xCD 0x6B 0xFF 0xFF
// 0x01 - unit id (0)
// 0x04 - function code (1)
// 0x02 - returned registers byte count (2)
// 0xCD 0x6B - input registers data (1 register) (3,4, ... 2 bytes for each register)
// 0xFF 0xFF - CRC16 (n-2,n-1) // FIXME: add correct crc value example
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
	dataLen := len(r.Data)
	result := make([]byte, tcpMBAPHeaderLen+3+dataLen)
	r.MBAPHeader.bytes(result[0:6])
	r.ReadInputRegistersResponse.bytes(result[6:])
	return result
}

// ParseReadInputRegistersResponseTCP parses given bytes into ReadInputRegistersResponseTCP
func ParseReadInputRegistersResponseTCP(data []byte) (*ReadInputRegistersResponseTCP, error) {
	dLen := len(data)
	if dLen < 11 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[8]
	if dLen != 9+int(byteLen) {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadInputRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
			Length:        binary.BigEndian.Uint16(data[4:6]),
		},
		ReadInputRegistersResponse: ReadInputRegistersResponse{
			UnitID: data[6],
			// function code = data[7]
			RegisterByteLen: data[8],
			Data:            data[9:],
		},
	}, nil
}

// Bytes returns ReadInputRegistersResponseRTU packet as bytes form
func (r ReadInputRegistersResponseRTU) Bytes() []byte {
	byteLen := r.RegisterByteLen
	result := make([]byte, 3+byteLen+2)
	bytes := r.ReadInputRegistersResponse.bytes(result)
	binary.BigEndian.PutUint16(result[3+byteLen:3+byteLen+2], CRC16(bytes))
	return result
}

// ParseReadInputRegistersResponseRTU parses given bytes into ParseReadInputRegistersResponseRTU
func ParseReadInputRegistersResponseRTU(data []byte) (*ReadInputRegistersResponseRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[2]
	if dLen != 3+int(byteLen)+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	// FIXME: check CRC
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

// Bytes returns ReadInputRegistersResponse packet as bytes form
func (r ReadInputRegistersResponse) Bytes() []byte {
	return r.bytes(make([]byte, 3+r.RegisterByteLen))
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
