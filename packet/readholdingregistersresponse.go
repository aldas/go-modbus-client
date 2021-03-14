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
	dataLen := len(r.Data)
	result := make([]byte, tcpMBAPHeaderLen+3+dataLen)
	r.MBAPHeader.bytes(result[0:6])
	r.ReadHoldingRegistersResponse.bytes(result[6:])
	return result
}

// ParseReadHoldingRegistersResponseTCP parses given bytes into ReadHoldingRegistersResponseTCP
func ParseReadHoldingRegistersResponseTCP(data []byte) (*ReadHoldingRegistersResponseTCP, error) {
	dLen := len(data)
	if dLen < 11 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[8]
	if dLen != 9+int(byteLen) {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadHoldingRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
			Length:        binary.BigEndian.Uint16(data[4:6]),
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
	byteLen := r.RegisterByteLen
	result := make([]byte, 3+byteLen+2)
	bytes := r.ReadHoldingRegistersResponse.bytes(result)
	binary.BigEndian.PutUint16(result[3+byteLen:3+byteLen+2], CRC16(bytes))
	return result
}

// ParseReadHoldingRegistersResponseRTU parses given bytes into ReadHoldingRegistersResponseTCP
func ParseReadHoldingRegistersResponseRTU(data []byte) (*ReadHoldingRegistersResponseRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[2]
	if dLen != 3+int(byteLen)+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	// FIXME: check CRC
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

// Bytes returns ReadHoldingRegistersResponse packet as bytes form
func (r ReadHoldingRegistersResponse) Bytes() []byte {
	return r.bytes(make([]byte, 3+r.RegisterByteLen))
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
