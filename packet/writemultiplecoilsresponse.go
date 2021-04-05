package packet

import (
	"encoding/binary"
	"errors"
)

// WriteMultipleCoilsResponseTCP is TCP Response for Write Multiple Coils (FC=15)
//
// Example packet: 0x01 0x38 0x00 0x00 0x00 0x06 0x11 0x0F 0x04 0x10 0x00 0x03
// 0x01 0x38 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x0F - function code (7)
// 0x04 0x10 - start address (8,9)
// 0x00 0x03 - count of coils written (10,11)
type WriteMultipleCoilsResponseTCP struct {
	MBAPHeader
	WriteMultipleCoilsResponse
}

// WriteMultipleCoilsResponseRTU is RTU Response for Write Multiple Coils (FC=15)
//
// Example packet: 0x11 0x0F 0x04 0x10 0x00 0x03 0x17 0xaf
// 0x11 - unit id (0)
// 0x0F - function code (1)
// 0x04 0x10 - start address (2,3)
// 0x00 0x03 - count of coils written (4,5)
// 0x17 0xaf - CRC16 (6,7)
type WriteMultipleCoilsResponseRTU struct {
	WriteMultipleCoilsResponse
}

// WriteMultipleCoilsResponse is Response for Write Multiple Coils (FC=15)
type WriteMultipleCoilsResponse struct {
	UnitID       uint8
	StartAddress uint16
	CoilCount    uint16
}

// Bytes returns WriteMultipleCoilsResponseTCP packet as bytes form
func (r WriteMultipleCoilsResponseTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteMultipleCoilsResponse.bytes(result[6 : 6+length])
	return result
}

// ParseWriteMultipleCoilsResponseTCP parses given bytes into ParseWriteMultipleCoilsResponseTCP
func ParseWriteMultipleCoilsResponseTCP(data []byte) (*WriteMultipleCoilsResponseTCP, error) {
	dLen := len(data)
	if dLen < 12 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	pduLen := binary.BigEndian.Uint16(data[4:6])
	if dLen != 6+int(pduLen) {
		return nil, errors.New("received data length does not match PDU len in packet")
	}
	return &WriteMultipleCoilsResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
			UnitID:       data[6],
			StartAddress: binary.BigEndian.Uint16(data[8:10]),
			CoilCount:    binary.BigEndian.Uint16(data[10:12]),
		},
	}, nil
}

// Bytes returns WriteMultipleCoilsResponseRTU packet as bytes form
func (r WriteMultipleCoilsResponseRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.WriteMultipleCoilsResponse.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ParseWriteMultipleCoilsResponseRTU parses given bytes into WriteMultipleCoilsResponseRTU
func ParseWriteMultipleCoilsResponseRTU(data []byte) (*WriteMultipleCoilsResponseRTU, error) {
	dLen := len(data)
	if dLen < 8 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	if dLen > 8 {
		return nil, errors.New("received data length too long to be valid packet")
	}
	return &WriteMultipleCoilsResponseRTU{
		WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
			UnitID: data[0],
			// data[1] function code
			StartAddress: binary.BigEndian.Uint16(data[2:4]),
			CoilCount:    binary.BigEndian.Uint16(data[4:6]),
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteMultipleCoilsResponse) FunctionCode() uint8 {
	return FunctionWriteMultipleCoils
}

// Bytes returns WriteMultipleCoilsResponse packet as bytes form
func (r WriteMultipleCoilsResponse) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r WriteMultipleCoilsResponse) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteMultipleCoils
	binary.BigEndian.PutUint16(bytes[2:4], r.StartAddress)
	binary.BigEndian.PutUint16(bytes[4:6], r.CoilCount)
	return bytes
}
