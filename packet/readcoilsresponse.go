package packet

import (
	"encoding/binary"
	"errors"
)

// ReadCoilsResponseTCP is TCP Response for Read Coils (FC=01)
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x05 0x03 0x01 0x02 0xCD 0x6B
// 0x81 0x80 - transaction id  (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x05 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x03 - unit id (6)
// 0x01 - function code (7)
// 0x02 - coils byte count (8)
// 0xCD 0x6B - coils data (2 bytes = 2 // 8 coils) (9,10, ...)
type ReadCoilsResponseTCP struct {
	MBAPHeader
	ReadCoilsResponse
}

// ReadCoilsResponseRTU is RTU Response for Read Coils (FC=01)
//
// Example packet: 0x03 0x01 0x02 0xCD 0x6B 0xd5 0x43
// 0x03 - unit id (0)
// 0x01 - function code (1)
// 0x02 - coils byte count (2)
// 0xCD 0x6B - coils data (2 bytes = 2 // 8 coils) (3,4, ...)
// 0xd5 0x43 - CRC16 (n-2,n-1)
type ReadCoilsResponseRTU struct {
	ReadCoilsResponse
}

// ReadCoilsResponse is Response for Read Coils (FC=01)
type ReadCoilsResponse struct {
	UnitID          uint8
	CoilsByteLength uint8
	Data            []byte
}

// Bytes returns ReadCoilsResponseTCP packet as bytes form
func (r ReadCoilsResponseTCP) Bytes() []byte {
	length := r.ReadCoilsResponse.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadCoilsResponse.bytes(result[6:])
	return result
}

// ParseReadCoilsResponseTCP parses given bytes into ReadCoilsResponseTCP
func ParseReadCoilsResponseTCP(data []byte) (*ReadCoilsResponseTCP, error) {
	dLen := len(data)
	if dLen < 10 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := int(data[8])
	if dLen != 9+byteLen {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadCoilsResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		ReadCoilsResponse: ReadCoilsResponse{
			UnitID: data[6],
			// function code = data[7]
			CoilsByteLength: data[8],
			Data:            data[9 : 9+byteLen],
		},
	}, nil
}

// Bytes returns ReadCoilsResponseRTU packet as bytes form
func (r ReadCoilsResponseRTU) Bytes() []byte {
	length := r.len() + 2
	result := make([]byte, length)
	bytes := r.ReadCoilsResponse.bytes(result)
	crc := CRC16(bytes[:length-2])
	result[length-2] = uint8(crc)
	result[length-1] = uint8(crc >> 8)
	return result
}

// ParseReadCoilsResponseRTU parses given bytes into ReadCoilsResponseRTU
func ParseReadCoilsResponseRTU(data []byte) (*ReadCoilsResponseRTU, error) {
	dLen := len(data)
	if dLen < 6 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := int(data[2])
	if dLen != 3+byteLen+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadCoilsResponseRTU{
		ReadCoilsResponse: ReadCoilsResponse{
			UnitID: data[0],
			// function code = data[1]
			CoilsByteLength: data[2],
			Data:            data[3 : 3+byteLen],
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadCoilsResponse) FunctionCode() uint8 {
	return FunctionReadCoils
}

func (r ReadCoilsResponse) len() uint16 {
	return 3 + uint16(len(r.Data))
}

// Bytes returns ReadCoilsResponse packet as bytes form
func (r ReadCoilsResponse) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
}

func (r ReadCoilsResponse) bytes(data []byte) []byte {
	data[0] = r.UnitID
	data[1] = FunctionReadCoils
	coilsByteLen := uint8(len(r.Data))
	data[2] = coilsByteLen
	copy(data[3:3+coilsByteLen], r.Data)

	return data
}

// IsCoilSet checks if N-th coil is set in response data. Coils are counted from `startAddress` (see ReadCoilsRequest) and right to left.
func (r ReadCoilsResponse) IsCoilSet(startAddress uint16, coilAddress uint16) (bool, error) {
	return isBitSet(r.Data, startAddress, coilAddress)
}
