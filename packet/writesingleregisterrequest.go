package packet

import (
	"encoding/binary"
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
// Example packet: 0x11 0x06 0x00 0x6B 0x01 0x01 0xFF 0xFF
// 0x11 - unit id (0)
// 0x06 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x01 0x01 - register data (4,5)
// 0xFF 0xFF - CRC16 (6,7) // FIXME: add correct crc value example
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
			Length:        6,
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
	result := make([]byte, tcpMBAPHeaderLen+6)
	r.MBAPHeader.bytes(result[0:6])
	r.WriteSingleRegisterRequest.bytes(result[6:12])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteSingleRegisterRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitID + 1 fc + 2 address + 2 register data
	return 6 + 6
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
	binary.BigEndian.PutUint16(result[6:8], CRC16(bytes))
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteSingleRegisterRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 address + 2 register data
	return 6
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
