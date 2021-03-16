package packet

import (
	"encoding/binary"
	"errors"
)

// ReadDiscreteInputsResponseTCP is TCP Response for Read Discrete Inputs (FC=02)
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x05 0x03 0x02 0x02 0xCD 0x6B
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x05 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x03 - unit id (6)
// 0x02 - function code (7)
// 0x02 - inputs byte count (8)
// 0xCD 0x6B - inputs discrete data (2 bytes = 2 // 8 inputs) (9, ...)
type ReadDiscreteInputsResponseTCP struct {
	MBAPHeader
	ReadDiscreteInputsResponse
}

// ReadDiscreteInputsResponseRTU is RTU Response for Read Discrete Inputs (FC=02)
//
// Example packet: 0x03 0x02 0x02 0xCD 0x6B 0xFF 0xFF
// 0x03 - unit id (0)
// 0x02 - function code (1)
// 0x02 - inputs byte count (2)
// 0xCD 0x6B - inputs data (2 bytes = 2 // 8 inputs) (3,4, ...)
// 0xFF 0xFF - CRC16 (n-2,n-1) // FIXME: add correct crc value example
type ReadDiscreteInputsResponseRTU struct {
	ReadDiscreteInputsResponse
}

// ReadDiscreteInputsResponse is Response for Read Discrete Inputs (FC=02)
type ReadDiscreteInputsResponse struct {
	UnitID           uint8
	InputsByteLength uint8
	Data             []byte
}

// Bytes returns ReadDiscreteInputsResponseTCP packet as bytes form
func (r ReadDiscreteInputsResponseTCP) Bytes() []byte {
	coilsByteLen := len(r.Data)
	result := make([]byte, tcpMBAPHeaderLen+3+coilsByteLen)
	r.MBAPHeader.bytes(result[0:6])
	r.ReadDiscreteInputsResponse.bytes(result[6:])
	return result
}

// ParseReadDiscreteInputsResponseTCP parses given bytes into ReadDiscreteInputsResponseTCP
func ParseReadDiscreteInputsResponseTCP(data []byte) (*ReadDiscreteInputsResponseTCP, error) {
	dLen := len(data)
	if dLen < 10 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[8]
	if dLen != 9+int(byteLen) {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadDiscreteInputsResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
			Length:        binary.BigEndian.Uint16(data[4:6]),
		},
		ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
			UnitID: data[6],
			// function code = data[7]
			InputsByteLength: data[8],
			Data:             data[9 : 9+byteLen],
		},
	}, nil
}

// Bytes returns ReadDiscreteInputsResponseRTU packet as bytes form
func (r ReadDiscreteInputsResponseRTU) Bytes() []byte {
	inputsByteLen := len(r.Data)
	result := make([]byte, 3+inputsByteLen+2)
	bytes := r.ReadDiscreteInputsResponse.bytes(result)
	binary.BigEndian.PutUint16(result[3+inputsByteLen:3+inputsByteLen+2], CRC16(bytes))
	return result
}

// ParseReadDiscreteInputsResponseRTU parses given bytes into ReadDiscreteInputsResponseRTU
func ParseReadDiscreteInputsResponseRTU(data []byte) (*ReadDiscreteInputsResponseRTU, error) {
	dLen := len(data)
	if dLen < 6 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	byteLen := data[2]
	if dLen != 3+int(byteLen)+2 {
		return nil, errors.New("received data length does not match byte len in packet")
	}
	return &ReadDiscreteInputsResponseRTU{
		ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
			UnitID: data[0],
			// function code = data[1]
			InputsByteLength: data[2],
			Data:             data[3 : 3+byteLen],
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadDiscreteInputsResponse) FunctionCode() uint8 {
	return FunctionReadDiscreteInputs
}

// Bytes returns ReadDiscreteInputsResponse packet as bytes form
func (r ReadDiscreteInputsResponse) Bytes() []byte {
	return r.bytes(make([]byte, 3+len(r.Data)))
}

func (r ReadDiscreteInputsResponse) bytes(data []byte) []byte {
	data[0] = r.UnitID
	data[1] = FunctionReadDiscreteInputs
	coilsByteLen := uint8(len(r.Data))
	data[2] = coilsByteLen
	copy(data[3:3+coilsByteLen], r.Data)

	return data
}

// IsInputSet checks if N-th discrete input is set in response data. Inputs are counted from `startAddress` (see ReadDiscreteInputsRequest) and right to left.
func (r ReadDiscreteInputsResponse) IsInputSet(startAddress uint16, inputAddress uint16) (bool, error) {
	return isBitSet(r.Data, startAddress, inputAddress)
}
