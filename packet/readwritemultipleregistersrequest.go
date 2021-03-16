package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
)

// ReadWriteMultipleRegistersRequestTCP is TCP Request for Read / Write Multiple Registers (FC=23)
//
// Example packet: 0x01 0x38 0x00 0x00 0x00 0x0f 0x11 0x17 0x04 0x10 0x00 0x01 0x01 0x12 0x00 0x02 0x04 0x00 0xc8 0x00 0x82
// 0x01 0x38 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x0f - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x17 - function code (7)
// 0x04 0x10 - read registers start address (8,9)
// 0x00 0x01 - read registers quantity (10,11)
// 0x01 0x12 - write register start address (12,13)
// 0x00 0x02 - write quantity (14,15)
// 0x04 - write bytes count (16)
// 0x00 0xc8 0x00 0x82 - write registers data (2 registers) (17,18, ...)
type ReadWriteMultipleRegistersRequestTCP struct {
	MBAPHeader
	ReadWriteMultipleRegistersRequest
}

// ReadWriteMultipleRegistersRequestRTU is RTU Request for Read / Write Multiple Registers (FC=23)
//
// Example packet: 0x11 0x17 0x04 0x10 0x00 0x01 0x01 0x12 0x00 0x02 0x04 0x00 0xc8 0x00 0x82 0xFF 0xFF
// 0x11 - unit id (0)
// 0x17 - function code (1)
// 0x04 0x10 - read registers start address (2,3)
// 0x00 0x01 - read registers quantity (4,5)
// 0x01 0x12 - write register start address (6,7)
// 0x00 0x02 - write quantity (8,9)
// 0x04 - write bytes count (10)
// 0x00 0xc8 0x00 0x82 - write registers data (2 registers) (11,12, ...)
// 0xFF 0xFF - CRC16 (n-2,n-1) // FIXME: add correct crc value example
type ReadWriteMultipleRegistersRequestRTU struct {
	ReadWriteMultipleRegistersRequest
}

// ReadWriteMultipleRegistersRequest is Request for Read / Write Multiple Registers (FC=23)
type ReadWriteMultipleRegistersRequest struct {
	UnitID uint8

	ReadStartAddress uint16
	ReadQuantity     uint16

	WriteStartAddress uint16
	WriteQuantity     uint16
	// WriteData must be in BigEndian byte order for server to interpret them correctly. We send them as is.
	WriteData []byte
}

// NewReadWriteMultipleRegistersRequestTCP creates new instance of Write Multiple Registers TCP request
// NB: bytes for `data` must be in BigEndian byte order for server to interpret them correctly
func NewReadWriteMultipleRegistersRequestTCP(
	unitID uint8,
	readStartAddress uint16,
	readQuantity uint16,
	writeStartAddress uint16,
	writeData []byte,
) (*ReadWriteMultipleRegistersRequestTCP, error) {
	if readQuantity == 0 || readQuantity > 124 {
		return nil, fmt.Errorf("read registers count out of range (1-124): %v", readQuantity)
	}
	writeByteCount := len(writeData)
	if writeByteCount%2 != 0 {
		return nil, errors.New("write data length must be even number of bytes")
	}
	writeRegisterCount := uint16(writeByteCount / 2)
	if writeRegisterCount == 0 || writeRegisterCount > 124 {
		return nil, fmt.Errorf("write registers count out of range (1-124): %v", writeRegisterCount)
	}

	return &ReadWriteMultipleRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
			Length:        11 + uint16(writeByteCount),
		},
		ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			ReadStartAddress: readStartAddress,
			ReadQuantity:     readQuantity,

			WriteStartAddress: writeStartAddress,
			WriteQuantity:     uint16(writeByteCount / 2),
			WriteData:         writeData,
		},
	}, nil
}

// Bytes returns ReadWriteMultipleRegistersRequestTCP packet as bytes form
func (r ReadWriteMultipleRegistersRequestTCP) Bytes() []byte {
	result := make([]byte, tcpMBAPHeaderLen+r.Length)
	r.MBAPHeader.bytes(result[0:6])
	r.ReadWriteMultipleRegistersRequest.bytes(result[6 : 6+r.Length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadWriteMultipleRegistersRequestTCP) ExpectedResponseLength() int {
	return 6 + 11 + int(r.ReadQuantity)*2
}

// NewReadWriteMultipleRegistersRequestRTU creates new instance of Write Multiple Registers RTU request
// NB: bytes for `data` must be in BigEndian byte order for server to interpret them correctly
func NewReadWriteMultipleRegistersRequestRTU(
	unitID uint8,
	readStartAddress uint16,
	readQuantity uint16,
	writeStartAddress uint16,
	writeData []byte,
) (*ReadWriteMultipleRegistersRequestRTU, error) {
	if readQuantity == 0 || readQuantity > 124 {
		return nil, fmt.Errorf("read registers count out of range (1-124): %v", readQuantity)
	}
	writeByteCount := len(writeData)
	if writeByteCount%2 != 0 {
		return nil, errors.New("write data length must be even number of bytes")
	}
	registerCount := uint16(writeByteCount / 2)
	if registerCount == 0 || registerCount > 124 {
		return nil, fmt.Errorf("write registers count out of range (1-124): %v", registerCount)
	}

	return &ReadWriteMultipleRegistersRequestRTU{
		ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			ReadStartAddress: readStartAddress,
			ReadQuantity:     readQuantity,

			WriteStartAddress: writeStartAddress,
			WriteQuantity:     uint16(writeByteCount / 2),
			WriteData:         writeData,
		},
	}, nil
}

// Bytes returns ReadWriteMultipleRegistersRequestRTU packet as bytes form
func (r ReadWriteMultipleRegistersRequestRTU) Bytes() []byte {
	pduLen := 11 + uint16(len(r.WriteData)) + 2
	result := make([]byte, pduLen)
	bytes := r.ReadWriteMultipleRegistersRequest.bytes(result)
	binary.BigEndian.PutUint16(result[pduLen-2:pduLen], CRC16(bytes))
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadWriteMultipleRegistersRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 registers bytes count + N registers data + 2 CRC
	return 4 + 2*int(r.ReadQuantity) + 2
}

// FunctionCode returns function code of this request
func (r ReadWriteMultipleRegistersRequest) FunctionCode() uint8 {
	return FunctionReadWriteMultipleRegisters
}

// Bytes returns ReadWriteMultipleRegistersRequest packet as bytes form
func (r ReadWriteMultipleRegistersRequest) Bytes() []byte {
	return r.bytes(make([]byte, 11+len(r.WriteData)))
}

func (r ReadWriteMultipleRegistersRequest) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionReadWriteMultipleRegisters
	binary.BigEndian.PutUint16(bytes[2:4], r.ReadStartAddress)
	binary.BigEndian.PutUint16(bytes[4:6], r.ReadQuantity)
	binary.BigEndian.PutUint16(bytes[6:8], r.WriteStartAddress)
	binary.BigEndian.PutUint16(bytes[8:10], r.WriteQuantity)
	bytes[10] = uint8(len(r.WriteData))
	copy(bytes[11:], r.WriteData)
	return bytes
}
