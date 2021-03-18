package packet

import (
	"encoding/binary"
	"errors"
)

// WriteMultipleRegistersResponseTCP is TCP Response for Write Multiple Registers (FC=16)
//
// Example packet: 0x01 0x38 0x00 0x00 0x00 0x06 0x11 0x10 0x04 0x10 0x00 0x03
// 0x01 0x38 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x10 - function code (7)
// 0x04 0x10 - start address (8,9)
// 0x00 0x03 - count of registers written (10,11)
type WriteMultipleRegistersResponseTCP struct {
	MBAPHeader
	WriteMultipleRegistersResponse
}

// WriteMultipleRegistersResponseRTU is RTU Response for Write Multiple Registers (FC=16)
//
// Example packet: 0x11 0x10 0x04 0x10 0x00 0x03 0xFF 0xFF
// 0x11 - unit id (0)
// 0x10 - function code (1)
// 0x04 0x10 - start address (2,3)
// 0x00 0x03 - count of registers written (4,5)
// 0xFF 0xFF - CRC16 (6,7) // FIXME: add correct crc value example
type WriteMultipleRegistersResponseRTU struct {
	WriteMultipleRegistersResponse
}

// WriteMultipleRegistersResponse is Response for Write Multiple Registers (FC=16)
type WriteMultipleRegistersResponse struct {
	UnitID        uint8
	StartAddress  uint16
	RegisterCount uint16
}

// Bytes returns WriteMultipleRegistersResponseTCP packet as bytes form
func (r WriteMultipleRegistersResponseTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteMultipleRegistersResponse.bytes(result[6 : 6+length])
	return result
}

// ParseWriteMultipleRegistersResponseTCP parses given bytes into ParseWriteMultipleRegistersResponseTCP
func ParseWriteMultipleRegistersResponseTCP(data []byte) (*WriteMultipleRegistersResponseTCP, error) {
	dLen := len(data)
	if dLen < 12 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	pduLen := binary.BigEndian.Uint16(data[4:6])
	if dLen != 6+int(pduLen) {
		return nil, errors.New("received data length does not match PDU len in packet")
	}
	return &WriteMultipleRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
			UnitID:        data[6],
			StartAddress:  binary.BigEndian.Uint16(data[8:10]),
			RegisterCount: binary.BigEndian.Uint16(data[10:12]),
		},
	}, nil
}

// Bytes returns WriteMultipleRegistersResponseRTU packet as bytes form
func (r WriteMultipleRegistersResponseRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.WriteMultipleRegistersResponse.bytes(result)
	binary.BigEndian.PutUint16(result[6:8], CRC16(bytes[:6]))
	return result
}

// ParseWriteMultipleRegistersResponseRTU parses given bytes into WriteMultipleRegistersResponseRTU
func ParseWriteMultipleRegistersResponseRTU(data []byte) (*WriteMultipleRegistersResponseRTU, error) {
	dLen := len(data)
	if dLen < 8 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	if dLen > 8 {
		return nil, errors.New("received data length too long to be valid packet")
	}
	return &WriteMultipleRegistersResponseRTU{
		WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
			UnitID: data[0],
			// data[1] function code
			StartAddress:  binary.BigEndian.Uint16(data[2:4]),
			RegisterCount: binary.BigEndian.Uint16(data[4:6]),
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteMultipleRegistersResponse) FunctionCode() uint8 {
	return FunctionWriteMultipleRegisters
}

// Bytes returns WriteMultipleRegistersResponse packet as bytes form
func (r WriteMultipleRegistersResponse) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r WriteMultipleRegistersResponse) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteMultipleRegisters
	binary.BigEndian.PutUint16(bytes[2:4], r.StartAddress)
	binary.BigEndian.PutUint16(bytes[4:6], r.RegisterCount)
	return bytes
}
