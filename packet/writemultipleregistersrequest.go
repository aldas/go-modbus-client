package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
)

// WriteMultipleRegistersRequestTCP is TCP Request for Write Multiple Registers (FC=16)
//
// Example packet: 0x01 0x38 0x00 0x00 0x00 0x0d 0x11 0x10 0x04 0x10 0x00 0x03 0x06 0x00 0xC8 0x00 0x82 0x87 0x01
// 0x01 0x38 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x0d - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x10 - function code (7)
// 0x04 0x10 - start address (8,9)
// 0x00 0x03 - count of register to write (10,11)
// 0x06 - registers byte count (12)
// 0x00 0xC8 0x00 0x82 0x87 0x01 - registers data (13, ...)
type WriteMultipleRegistersRequestTCP struct {
	MBAPHeader
	WriteMultipleRegistersRequest
}

// WriteMultipleRegistersRequestRTU is RTU Request for Write Multiple Registers (FC=16)
//
// Example packet: 0x11 0x10 0x04 0x10 0x00 0x03 0x06 0x00 0xC8 0x00 0x82 0x87 0x01 0x2f 0x7d
// 0x11 - unit id (0)
// 0x10 - function code (1)
// 0x04 0x10 - start address (2,3)
// 0x00 0x03 - count of register to write (4,5)
// 0x06 - registers byte count (6)
// 0x00 0xC8 0x00 0x82 0x87 0x01 - registers data (7,8, ...)
// 0x2f 0x7d - CRC16 (n-2,n-1)
type WriteMultipleRegistersRequestRTU struct {
	WriteMultipleRegistersRequest
}

// WriteMultipleRegistersRequest is Request for Write Multiple Registers (FC=16)
type WriteMultipleRegistersRequest struct {
	UnitID        uint8
	StartAddress  uint16
	RegisterCount uint16
	// Data must be in BigEndian byte order for server to interpret them correctly. We send them as is.
	Data []byte
}

// NewWriteMultipleRegistersRequestTCP creates new instance of Write Multiple Registers TCP request
// NB: bytes for `data` must be in BigEndian byte order for server to interpret them correctly
func NewWriteMultipleRegistersRequestTCP(unitID uint8, startAddress uint16, data []byte) (*WriteMultipleRegistersRequestTCP, error) {
	registerByteCount := len(data)
	if registerByteCount%2 != 0 {
		return nil, errors.New("data length must be even number of bytes")
	}
	registerCount := uint16(registerByteCount / 2)
	if registerCount == 0 || registerCount > 124 {
		return nil, fmt.Errorf("registers count out of range (1-124): %v", registerCount)
	}

	return &WriteMultipleRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress:  startAddress,
			RegisterCount: registerCount,
			Data:          data,
		},
	}, nil
}

// Bytes returns WriteMultipleRegistersRequestTCP packet as bytes form
func (r WriteMultipleRegistersRequestTCP) Bytes() []byte {
	length := r.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteMultipleRegistersRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteMultipleRegistersRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 UnitID + 1 functionCode + 2 start addr + 2 count of registers
	return 6 + 6
}

// NewWriteMultipleRegistersRequestRTU creates new instance of Write Multiple Registers RTU request
// NB: bytes for `data` must be in BigEndian byte order for server to interpret them correctly
func NewWriteMultipleRegistersRequestRTU(unitID uint8, startAddress uint16, data []byte) (*WriteMultipleRegistersRequestRTU, error) {
	registerByteCount := len(data)
	if registerByteCount%2 != 0 {
		return nil, errors.New("data length must be even number of bytes")
	}
	registerCount := uint16(registerByteCount / 2)
	if registerCount == 0 || registerCount > 124 {
		return nil, fmt.Errorf("registers count out of range (1-124): %v", registerCount)
	}

	return &WriteMultipleRegistersRequestRTU{
		WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress:  startAddress,
			RegisterCount: registerCount,
			Data:          data,
		},
	}, nil
}

// Bytes returns WriteMultipleRegistersRequestRTU packet as bytes form
func (r WriteMultipleRegistersRequestRTU) Bytes() []byte {
	pduLen := r.len() + 2
	result := make([]byte, pduLen)
	bytes := r.WriteMultipleRegistersRequest.bytes(result)
	crc := CRC16(bytes[:pduLen-2])
	result[pduLen-2] = uint8(crc)
	result[pduLen-1] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteMultipleRegistersRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 start addr + 2 count of registers + 2 CRC
	return 6 + 2
}

// FunctionCode returns function code of this request
func (r WriteMultipleRegistersRequest) FunctionCode() uint8 {
	return FunctionWriteMultipleRegisters
}

func (r WriteMultipleRegistersRequest) len() uint16 {
	return 7 + uint16(len(r.Data))
}

// Bytes returns WriteMultipleRegistersRequest packet as bytes form
func (r WriteMultipleRegistersRequest) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
}

func (r WriteMultipleRegistersRequest) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteMultipleRegisters
	binary.BigEndian.PutUint16(bytes[2:4], r.StartAddress)
	binary.BigEndian.PutUint16(bytes[4:6], r.RegisterCount)
	bytes[6] = uint8(len(r.Data))
	copy(bytes[7:], r.Data)
	return bytes
}
