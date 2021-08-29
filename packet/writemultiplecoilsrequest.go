package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
)

// WriteMultipleCoilsRequestTCP is TCP Request for Write Multiple Coils (FC=15)
//
// Example packet: 0x01 0x38 0x00 0x00 0x00 0x08 0x11 0x0F 0x04 0x10 0x00 0x03 0x01 0x05
// 0x01 0x38 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x08 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x0F - function code (7)
// 0x04 0x10 - start address (8,9)
// 0x00 0x03 - count of coils to write (10,11)
// 0x01 - coils byte count (12)
// 0x05 - coils data (13, ...)
type WriteMultipleCoilsRequestTCP struct {
	MBAPHeader
	WriteMultipleCoilsRequest
}

// WriteMultipleCoilsRequestRTU is RTU Request for Write Multiple Coils (FC=15)
//
// Example packet: 0x11 0x0F 0x04 0x10 0x00 0x03 0x01 0x05 0x8e 0x1f
// 0x11 - unit id (0)
// 0x0F - function code (1)
// 0x04 0x10 - start address (2,3)
// 0x00 0x03 - count of coils to write (4,5)
// 0x01 - coils byte count (6)
// 0x05 - coils data (7, ...)
// 0x8e 0x1f - CRC16 (n-2,n-1)
type WriteMultipleCoilsRequestRTU struct {
	WriteMultipleCoilsRequest
}

// WriteMultipleCoilsRequest is Request for Write Multiple Coils (FC=15)
type WriteMultipleCoilsRequest struct {
	UnitID       uint8
	StartAddress uint16
	CoilCount    uint16
	Data         []byte
}

// NewWriteMultipleCoilsRequestTCP creates new instance of Write Multiple Coils TCP request
func NewWriteMultipleCoilsRequestTCP(unitID uint8, startAddress uint16, coils []bool) (*WriteMultipleCoilsRequestTCP, error) {
	coilsCount := len(coils)
	if coilsCount == 0 || coilsCount > 1968 {
		// 1968 coils is due that coils byte len size field is 1 byte so max 246*8=1968 coils can be sent
		return nil, fmt.Errorf("coils count is out of range (1-1968): %v", coilsCount)
	}

	coilsBytes := CoilsToBytes(coils)
	return &WriteMultipleCoilsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			CoilCount:    uint16(coilsCount),
			Data:         coilsBytes,
		},
	}, nil
}

// Bytes returns WriteMultipleCoilsRequestTCP packet as bytes form
func (r WriteMultipleCoilsRequestTCP) Bytes() []byte {
	length := r.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteMultipleCoilsRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteMultipleCoilsRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 UnitID + 1 functionCode + 2 start addr + 2 count of coils
	return 6 + 6
}

// ParseWriteMultipleCoilsRequestTCP parses given bytes into WriteMultipleCoilsRequestTCP
func ParseWriteMultipleCoilsRequestTCP(data []byte) (*WriteMultipleCoilsRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	if data[7] != FunctionWriteMultipleCoils {
		return nil, errors.New("received function code in packet is not 0x0f")
	}
	coilCount := binary.BigEndian.Uint16(data[10:12])
	if !(coilCount >= 1 && coilCount <= 1968) { // 0x0001 to 0x07B0
		return nil, errors.New("invalid coils count. valid range 1..1968")
	}
	coilsBytesCount := data[12]
	if len(data) < 13+int(coilsBytesCount) {
		return nil, errors.New("received data coils bytes length does not match write data length")
	}
	var coilsData []byte
	if coilsBytesCount > 0 {
		coilsData = make([]byte, coilsBytesCount)
		copy(coilsData, data[13:13+coilsBytesCount])
	}
	return &WriteMultipleCoilsRequestTCP{
		MBAPHeader: header,
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID: data[6],
			// function code = data[7]
			StartAddress: binary.BigEndian.Uint16(data[8:10]),
			CoilCount:    coilCount,
			Data:         coilsData,
		},
	}, nil
}

// NewWriteMultipleCoilsRequestRTU creates new instance of Write Multiple Coils RTU request
func NewWriteMultipleCoilsRequestRTU(unitID uint8, startAddress uint16, coils []bool) (*WriteMultipleCoilsRequestRTU, error) {
	coilsCount := len(coils)
	if coilsCount == 0 || coilsCount > 1968 {
		// 1968 coils is due that coils byte len size field is 1 byte so max 246*8=1968 coils can be sent
		return nil, fmt.Errorf("coils count is out of range (1-1968): %v", coilsCount)
	}

	coilsBytes := CoilsToBytes(coils)
	return &WriteMultipleCoilsRequestRTU{
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			CoilCount:    uint16(coilsCount),
			Data:         coilsBytes,
		},
	}, nil
}

// Bytes returns WriteMultipleCoilsRequestRTU packet as bytes form
func (r WriteMultipleCoilsRequestRTU) Bytes() []byte {
	pduLen := r.len() + 2
	result := make([]byte, pduLen)
	bytes := r.WriteMultipleCoilsRequest.bytes(result)
	crc := CRC16(bytes[:pduLen-2])
	result[pduLen-2] = uint8(crc)
	result[pduLen-1] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteMultipleCoilsRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 start addr + 2 count of coils + 2 CRC
	return 6 + 2
}

// ParseWriteMultipleCoilsRequestRTU parses given bytes into WriteMultipleCoilsRequestRTU
func ParseWriteMultipleCoilsRequestRTU(data []byte) (*WriteMultipleCoilsRequestRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	if data[1] != FunctionWriteMultipleCoils {
		return nil, errors.New("received function code in packet is not 0x0f")
	}
	coilCount := binary.BigEndian.Uint16(data[4:6])
	if !(coilCount >= 1 && coilCount <= 1968) { // 0x0001 to 0x07B0
		return nil, errors.New("invalid coils count. valid range 1..1968")
	}
	coilsBytesCount := data[6]
	expectedLen := 7 + int(coilsBytesCount)
	if dLen != expectedLen && dLen != expectedLen+2 { // without crc and with crc
		return nil, errors.New("received data coils bytes length does not match write data length")
	}
	var coilsData []byte
	if coilsBytesCount > 0 {
		coilsData = make([]byte, coilsBytesCount)
		copy(coilsData, data[7:7+coilsBytesCount])
	}
	return &WriteMultipleCoilsRequestRTU{
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID: data[0],
			// function code = data[1]
			StartAddress: binary.BigEndian.Uint16(data[2:4]),
			CoilCount:    coilCount,
			Data:         coilsData,
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteMultipleCoilsRequest) FunctionCode() uint8 {
	return FunctionWriteMultipleCoils
}

func (r WriteMultipleCoilsRequest) len() uint16 {
	return 7 + uint16(len(r.Data))
}

// Bytes returns WriteMultipleCoilsRequest packet as bytes form
func (r WriteMultipleCoilsRequest) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
}

func (r WriteMultipleCoilsRequest) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteMultipleCoils
	binary.BigEndian.PutUint16(bytes[2:4], r.StartAddress)
	binary.BigEndian.PutUint16(bytes[4:6], r.CoilCount)
	bytes[6] = uint8(len(r.Data))
	copy(bytes[7:], r.Data)
	return bytes
}

// CoilsToBytes converts slice of coil states (as bool values) to byte slice.
func CoilsToBytes(coils []bool) []byte {
	cLen := len(coils)
	cnt := cLen / 8
	if cLen%8 != 0 {
		cnt++
	}
	result := make([]byte, cnt)
	for i := 0; i < cLen; i++ {
		bit := i % 8
		nthByte := i / 8
		if coils[i] {
			v := result[nthByte] | (1 << bit)
			result[nthByte] = v
		}
	}
	return result
}
