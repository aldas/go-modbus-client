package packet

import (
	"encoding/binary"
	"errors"
)

// ReadWriteMultipleRegistersResponseTCP is TCP Response for Read / Write Multiple Registers request (FC=23)
//
// Example packet: 0x01 0x38 0x00 0x00 0x00 0x05 0x11 0x17 0x02 0xCD 0x6B
// 0x01 0x38 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x05 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x17 - function code (7)
// 0x02 - registers bytes count (8)
// 0xCD 0x6B - write registers data (1 registers) (9, 10, ...)
type ReadWriteMultipleRegistersResponseTCP struct {
	MBAPHeader
	ReadWriteMultipleRegistersResponse
}

// ReadWriteMultipleRegistersResponseRTU is RTU Response for Read / Write Multiple Registers request (FC=23)
//
// Example packet: 0x11 0x17 0x02 0xCD 0x6B 0xFF 0xFF
// 0x11 - unit id (0)
// 0x17 - function code (1)
// 0x02 - registers bytes count (2)
// 0xCD 0x6B - write registers data (1 registers) (3, 4, ...)
// 0xFF 0xFF - CRC16 (n-2,n-1) // FIXME: add correct crc value example
type ReadWriteMultipleRegistersResponseRTU struct {
	ReadWriteMultipleRegistersResponse
}

// ReadWriteMultipleRegistersResponse is Response for Read / Write Multiple Registers request (FC=23)
type ReadWriteMultipleRegistersResponse struct {
	UnitID          uint8
	RegisterByteLen uint8
	Data            []byte
}

// Bytes returns ReadWriteMultipleRegistersResponseTCP packet as bytes form
func (r ReadWriteMultipleRegistersResponseTCP) Bytes() []byte {
	dataLen := len(r.Data)
	result := make([]byte, tcpMBAPHeaderLen+3+dataLen)
	r.MBAPHeader.bytes(result[0:6])
	r.ReadWriteMultipleRegistersResponse.bytes(result[6:])
	return result
}

// ParseReadWriteMultipleRegistersResponseTCP parses given bytes into ReadWriteMultipleRegistersResponseTCP
func ParseReadWriteMultipleRegistersResponseTCP(data []byte) (*ReadWriteMultipleRegistersResponseTCP, error) {
	dLen := len(data)
	if dLen < 11 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[8]
	if dLen != 9+int(byteLen) {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadWriteMultipleRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
			Length:        binary.BigEndian.Uint16(data[4:6]),
		},
		ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
			UnitID: data[6],
			// function code = data[7]
			RegisterByteLen: data[8],
			Data:            data[9 : 9+byteLen],
		},
	}, nil
}

// Bytes returns ReadWriteMultipleRegistersResponseRTU packet as bytes form
func (r ReadWriteMultipleRegistersResponseRTU) Bytes() []byte {
	byteLen := r.RegisterByteLen
	result := make([]byte, 3+byteLen+2)
	bytes := r.ReadWriteMultipleRegistersResponse.bytes(result)
	binary.BigEndian.PutUint16(result[3+byteLen:3+byteLen+2], CRC16(bytes))
	return result
}

// ParseReadWriteMultipleRegistersResponseRTU parses given bytes into ReadWriteMultipleRegistersResponseTCP
func ParseReadWriteMultipleRegistersResponseRTU(data []byte) (*ReadWriteMultipleRegistersResponseRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[2]
	if dLen != 3+int(byteLen)+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	// FIXME: check CRC
	return &ReadWriteMultipleRegistersResponseRTU{
		ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
			UnitID: data[0],
			// function code = data[1]
			RegisterByteLen: data[2],
			Data:            data[3 : 3+byteLen],
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadWriteMultipleRegistersResponse) FunctionCode() uint8 {
	return FunctionReadWriteMultipleRegisters
}

// Bytes returns ReadWriteMultipleRegistersResponse packet as bytes form
func (r ReadWriteMultipleRegistersResponse) Bytes() []byte {
	return r.bytes(make([]byte, 3+r.RegisterByteLen))
}

func (r ReadWriteMultipleRegistersResponse) bytes(data []byte) []byte {
	data[0] = r.UnitID
	data[1] = FunctionReadWriteMultipleRegisters
	data[2] = r.RegisterByteLen
	copy(data[3:], r.Data)

	return data
}

// AsRegisters returns response data as Register to more convenient access
func (r ReadWriteMultipleRegistersResponse) AsRegisters(requestStartAddress uint16) (*Registers, error) {
	return NewRegisters(r.Data, requestStartAddress)
}
