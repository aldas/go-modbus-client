package packet

import (
	"encoding/binary"
	"errors"
	"math/rand"
)

// WriteSingleRegisterRequestTCP is TCP Request for Write Single Register (FC=06)
//
// Example packet: 0x00 0x01 0x00 0x00 0x00 0x06 0x11 0x06 0x00 0x6B 0x01 0x01
// 0x00 0x01 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x06 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0x01 0x01 - register data (10,11)
type WriteSingleRegisterRequestTCP struct {
	MBAPHeader
	WriteSingleRegisterRequest
}

// WriteSingleRegisterRequestRTU is RTU Request for Write Single Register (FC=06)
//
// Example packet: 0x11 0x06 0x00 0x6B 0x01 0x01 0x3a 0xd6
// 0x11 - unit id (0)
// 0x06 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x01 0x01 - register data (4,5)
// 0x3a 0xd6 - CRC16 (6,7)
type WriteSingleRegisterRequestRTU struct {
	WriteSingleRegisterRequest
}

// WriteSingleRegisterRequest is Request for Write Single Register (FC=06)
type WriteSingleRegisterRequest struct {
	UnitID  uint8
	Address uint16
	// Data must be in BigEndian byte order for server to interpret them correctly. We send them as is.
	Data [2]byte
}

// NewWriteSingleRegisterRequestTCP creates new instance of Write Single Register TCP request
// NB: byte slice for `data` must be in BigEndian byte order for server to interpret them correctly
func NewWriteSingleRegisterRequestTCP(unitID uint8, address uint16, data []byte) (*WriteSingleRegisterRequestTCP, error) {
	w := &WriteSingleRegisterRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			Address: address,
		},
	}
	copy(w.Data[:], data)
	return w, nil
}

// Bytes returns WriteSingleRegisterRequestTCP packet as bytes form
func (r WriteSingleRegisterRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteSingleRegisterRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteSingleRegisterRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitID + 1 fc + 2 address + 2 register data
	return 6 + 6
}

// ParseWriteSingleRegisterRequestTCP parses given bytes into WriteSingleRegisterRequestTCP
func ParseWriteSingleRegisterRequestTCP(data []byte) (*WriteSingleRegisterRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	if data[7] != FunctionWriteSingleRegister {
		return nil, errors.New("received function code in packet is not 0x06")
	}
	return &WriteSingleRegisterRequestTCP{
		MBAPHeader: header,
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID: data[6],
			// function code = data[7]
			Address: binary.BigEndian.Uint16(data[8:10]),
			Data:    [2]byte{data[10], data[11]},
		},
	}, nil
}

// NewWriteSingleRegisterRequestRTU creates new instance of Write Single Register RTU request
// NB: byte slice for `data` must be in BigEndian byte order for server to interpret them correctly
func NewWriteSingleRegisterRequestRTU(unitID uint8, address uint16, data []byte) (*WriteSingleRegisterRequestRTU, error) {
	w := &WriteSingleRegisterRequestRTU{
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			Address: address,
		},
	}
	copy(w.Data[:], data)
	return w, nil
}

// Bytes returns WriteSingleRegisterRequestRTU packet as bytes form
func (r WriteSingleRegisterRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.WriteSingleRegisterRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteSingleRegisterRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 address + 2 register data
	return 6
}

// ParseWriteSingleRegisterRequestRTU parses given bytes into WriteSingleRegisterRequestRTU
func ParseWriteSingleRegisterRequestRTU(data []byte) (*WriteSingleRegisterRequestRTU, error) {
	dLen := len(data)
	if dLen != 8 && dLen != 6 { // with or without CRC
		return nil, errors.New("received data length too short to be valid packet")
	}
	if data[1] != FunctionWriteSingleRegister {
		return nil, errors.New("received function code in packet is not 0x06")
	}
	return &WriteSingleRegisterRequestRTU{
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID: data[0],
			// function code = data[1]
			Address: binary.BigEndian.Uint16(data[2:4]),
			Data:    [2]byte{data[4], data[5]},
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteSingleRegisterRequest) FunctionCode() uint8 {
	return FunctionWriteSingleRegister
}

// Bytes returns WriteSingleRegisterRequest packet as bytes form
func (r WriteSingleRegisterRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r WriteSingleRegisterRequest) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteSingleRegister
	binary.BigEndian.PutUint16(bytes[2:4], r.Address)
	copy(bytes[4:6], r.Data[:])
	return bytes
}
