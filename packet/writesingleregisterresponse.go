package packet

import (
	"encoding/binary"
	"errors"
)

// WriteSingleRegisterResponseTCP is TCP Response for Write Single Register (FC=06)
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x06 0x03 0x06 0x00 0x02 0xFF 0x00
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x03 - unit id (6)
// 0x06 - function code (7)
// 0x00 0x02 - start address (8,9)
// 0xFF 0x00 - register data (10,11)
type WriteSingleRegisterResponseTCP struct {
	MBAPHeader
	WriteSingleRegisterResponse
}

// WriteSingleRegisterResponseRTU is RTU Response for Write Single Register (FC=06)
//
// Example packet: 0x11 0x06 0x00 0x6B 0x01 0x01 0x3a 0xd6
// 0x11 - unit id (0)
// 0x06 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x01 0x01 - register data (4,5)
// 0x3a 0xd6 - CRC16 (6,7)
type WriteSingleRegisterResponseRTU struct {
	WriteSingleRegisterResponse
}

// WriteSingleRegisterResponse is Response for Write Single Register (FC=06)
type WriteSingleRegisterResponse struct {
	UnitID  uint8
	Address uint16
	Data    [2]byte
}

// Bytes returns WriteSingleRegisterResponseTCP packet as bytes form
func (r WriteSingleRegisterResponseTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteSingleRegisterResponse.bytes(result[6 : 6+length])
	return result
}

// ParseWriteSingleRegisterResponseTCP parses given bytes into WriteSingleRegisterResponseTCP
func ParseWriteSingleRegisterResponseTCP(data []byte) (*WriteSingleRegisterResponseTCP, error) {
	dLen := len(data)
	if dLen < 12 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	pduLen := binary.BigEndian.Uint16(data[4:6])
	if dLen != 6+int(pduLen) {
		return nil, errors.New("received data length does not match PDU len in packet")
	}

	return &WriteSingleRegisterResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		WriteSingleRegisterResponse: WriteSingleRegisterResponse{
			UnitID:  data[6],
			Address: binary.BigEndian.Uint16(data[8:10]),
			Data:    [2]byte{data[10], data[11]},
		},
	}, nil
}

// Bytes returns WriteSingleRegisterResponseRTU packet as bytes form
func (r WriteSingleRegisterResponseRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.WriteSingleRegisterResponse.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ParseWriteSingleRegisterResponseRTU parses given bytes into WriteSingleRegisterResponseRTU
func ParseWriteSingleRegisterResponseRTU(data []byte) (*WriteSingleRegisterResponseRTU, error) {
	dLen := len(data)
	if dLen < 8 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	if dLen > 8 {
		return nil, errors.New("received data length too long to be valid packet")
	}
	return &WriteSingleRegisterResponseRTU{
		WriteSingleRegisterResponse: WriteSingleRegisterResponse{
			UnitID: data[0],
			// data[1] function code
			Address: binary.BigEndian.Uint16(data[2:4]),
			Data:    [2]byte{data[4], data[5]},
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteSingleRegisterResponse) FunctionCode() uint8 {
	return FunctionWriteSingleRegister
}

// Bytes returns WriteSingleRegisterResponse packet as bytes form
func (r WriteSingleRegisterResponse) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r WriteSingleRegisterResponse) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteSingleRegister
	binary.BigEndian.PutUint16(bytes[2:4], r.Address)
	copy(bytes[4:6], r.Data[:])
	return bytes
}

// AsRegisters returns response data as Register to more convenient access
func (r WriteSingleRegisterResponse) AsRegisters(address uint16) (*Registers, error) {
	return NewRegisters(r.Data[:], address)
}
