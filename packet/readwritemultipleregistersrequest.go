package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
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
// Example packet: 0x11 0x17 0x04 0x10 0x00 0x01 0x01 0x12 0x00 0x02 0x04 0x00 0xc8 0x00 0x82 0x64 0xe2
// 0x11 - unit id (0)
// 0x17 - function code (1)
// 0x04 0x10 - read registers start address (2,3)
// 0x00 0x01 - read registers quantity (4,5)
// 0x01 0x12 - write register start address (6,7)
// 0x00 0x02 - write quantity (8,9)
// 0x04 - write bytes count (10)
// 0x00 0xc8 0x00 0x82 - write registers data (2 registers) (11,12, ...)
// 0x64 0xe2 - CRC16 (n-2,n-1)
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
			TransactionID: 1 + rand.N(uint16(65534)), // #nosec G404
			ProtocolID:    0,
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
	length := r.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadWriteMultipleRegistersRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadWriteMultipleRegistersRequestTCP) ExpectedResponseLength() int {
	return 6 + 11 + int(r.ReadQuantity)*2
}

// ParseReadWriteMultipleRegistersRequestTCP parses given bytes into ReadWriteMultipleRegistersRequestTCP
func ParseReadWriteMultipleRegistersRequestTCP(data []byte) (*ReadWriteMultipleRegistersRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	unitID := data[6]
	if data[7] != FunctionReadWriteMultipleRegisters {
		tmpErr := NewErrorParseTCP(ErrIllegalFunction, "received function code in packet is not 0x17")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	readQuantity := binary.BigEndian.Uint16(data[10:12])
	if !(readQuantity >= 1 && readQuantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseTCP(ErrIllegalDataValue, "invalid read quantity. valid range 1..125")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	writeQuantity := binary.BigEndian.Uint16(data[14:16])
	if !(writeQuantity >= 1 && writeQuantity <= 121) { // 0x0001 to 0x0079
		tmpErr := NewErrorParseTCP(ErrIllegalDataValue, "invalid write quantity. valid range 1..121")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	writeBytesCount := data[16]
	if len(data) < 17+int(writeBytesCount) {
		tmpErr := NewErrorParseTCP(ErrIllegalDataValue, "received data write bytes length does not match write data length")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	var writeData []byte
	if writeBytesCount > 0 {
		writeData = make([]byte, writeBytesCount)
		copy(writeData, data[17:17+writeBytesCount])
	}
	return &ReadWriteMultipleRegistersRequestTCP{
		MBAPHeader: header,
		ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
			UnitID: unitID,
			// function code = data[7]
			ReadStartAddress:  binary.BigEndian.Uint16(data[8:10]),
			ReadQuantity:      readQuantity,
			WriteStartAddress: binary.BigEndian.Uint16(data[12:14]),
			WriteQuantity:     writeQuantity,
			WriteData:         writeData,
		},
	}, nil
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
	pduLen := r.len() + 2
	result := make([]byte, pduLen)
	bytes := r.ReadWriteMultipleRegistersRequest.bytes(result)
	crc := CRC16(bytes[:pduLen-2])
	result[pduLen-2] = uint8(crc)
	result[pduLen-1] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadWriteMultipleRegistersRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 registers bytes count + N registers data + 2 CRC
	return 4 + 2*int(r.ReadQuantity) + 2
}

// ParseReadWriteMultipleRegistersRequestRTU parses given bytes into ReadWriteMultipleRegistersRequestRTU
func ParseReadWriteMultipleRegistersRequestRTU(data []byte) (*ReadWriteMultipleRegistersRequestRTU, error) {
	dLen := len(data)
	if dLen < 12 {
		return nil, NewErrorParseRTU(ErrServerFailure, "received data length too short to be valid packet")
	}
	unitID := data[0]
	if data[1] != FunctionReadWriteMultipleRegisters {
		tmpErr := NewErrorParseRTU(ErrIllegalFunction, "received function code in packet is not 0x17")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	readQuantity := binary.BigEndian.Uint16(data[4:6])
	if !(readQuantity >= 1 && readQuantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseRTU(ErrIllegalDataValue, "invalid read quantity. valid range 1..125")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	writeQuantity := binary.BigEndian.Uint16(data[8:10])
	if !(writeQuantity >= 1 && writeQuantity <= 121) { // 0x0001 to 0x0079
		tmpErr := NewErrorParseRTU(ErrIllegalDataValue, "invalid write quantity. valid range 1..121")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	writeBytesCount := data[10]
	expectedLen := 11 + int(writeBytesCount)
	if dLen != expectedLen && dLen != expectedLen+2 { // without crc and with crc
		tmpErr := NewErrorParseRTU(ErrIllegalDataValue, "received data write bytes length does not match write data length")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadWriteMultipleRegisters
		return nil, tmpErr
	}
	var writeData []byte
	if writeBytesCount > 0 {
		writeData = make([]byte, writeBytesCount)
		copy(writeData, data[11:11+writeBytesCount])
	}
	return &ReadWriteMultipleRegistersRequestRTU{
		ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
			UnitID: unitID,
			// function code = data[1]
			ReadStartAddress:  binary.BigEndian.Uint16(data[2:4]),
			ReadQuantity:      readQuantity,
			WriteStartAddress: binary.BigEndian.Uint16(data[6:8]),
			WriteQuantity:     writeQuantity,
			WriteData:         writeData,
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadWriteMultipleRegistersRequest) FunctionCode() uint8 {
	return FunctionReadWriteMultipleRegisters
}

func (r ReadWriteMultipleRegistersRequest) len() uint16 {
	return 11 + uint16(len(r.WriteData))
}

// Bytes returns ReadWriteMultipleRegistersRequest packet as bytes form
func (r ReadWriteMultipleRegistersRequest) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
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
