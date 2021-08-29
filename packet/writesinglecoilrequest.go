package packet

import (
	"encoding/binary"
	"errors"
	"math/rand"
)

// WriteSingleCoilRequestTCP is TCP Request for Write Single Coil (FC=05)
//
// Data part of packet is always 4 bytes - 2 byte for address and 2 byte for coil status (FF00 = on,  0000 = off).
// For example: coil at address 1 is turned on '0x00 0x01 0xFF 0x00'
// For example: coil at address 10 is turned off '0x00 0x0A 0x00 0x00'
//
// Example packet: 0x00 0x01 0x00 0x00 0x00 0x06 0x11 0x05 0x00 0x6B 0xFF 0x00
// 0x00 0x01 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x11 - unit id (6)
// 0x05 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0xFF 0x00 - coil data (true) (10,11)
type WriteSingleCoilRequestTCP struct {
	MBAPHeader
	WriteSingleCoilRequest
}

// WriteSingleCoilRequestRTU is RTU Request for Write Single Coil (FC=05)
//
// Data part of packet is always 4 bytes - 2 byte for address and 2 byte for coil status (FF00 = on,  0000 = off).
// For example: coil at address 1 is turned on '0x00 0x01 0xFF 0x00'
// For example: coil at address 10 is turned off '0x00 0x0A 0x00 0x00'
//
// Example packet: 0x11 0x05 0x00 0x6B 0xFF 0x00 0xff 0x76
// 0x11 - unit id (0)
// 0x05 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0xFF 0x00 - coil data (true) (4,5)
// 0xff 0x76 - CRC16 (6,7)
type WriteSingleCoilRequestRTU struct {
	WriteSingleCoilRequest
}

// WriteSingleCoilRequest is Request for Write Single Coil (FC=05)
type WriteSingleCoilRequest struct {
	UnitID    uint8
	Address   uint16
	CoilState bool
}

// NewWriteSingleCoilRequestTCP creates new instance of Write Single Coil TCP request
func NewWriteSingleCoilRequestTCP(unitID uint8, address uint16, coilState bool) (*WriteSingleCoilRequestTCP, error) {
	return &WriteSingleCoilRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			Address:   address,
			CoilState: coilState,
		},
	}, nil
}

// Bytes returns WriteSingleCoilRequestTCP packet as bytes form
func (r WriteSingleCoilRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.WriteSingleCoilRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteSingleCoilRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitID + 1 fc + 1 coil byte count + 2 data len
	return 6 + 3 + 2
}

// ParseWriteSingleCoilRequestTCP parses given bytes into WriteSingleCoilRequestTCP
func ParseWriteSingleCoilRequestTCP(data []byte) (*WriteSingleCoilRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	if data[7] != FunctionWriteSingleCoil {
		return nil, errors.New("received function code in packet is not 0x05")
	}
	coilStateRaw := binary.BigEndian.Uint16(data[10:12])
	if coilStateRaw != writeCoilOn && coilStateRaw != writeCoilOff {
		return nil, errors.New("coil state has invalid value")
	}
	return &WriteSingleCoilRequestTCP{
		MBAPHeader: header,
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID: data[6],
			// function code = data[7]
			Address:   binary.BigEndian.Uint16(data[8:10]),
			CoilState: coilStateRaw == writeCoilOn,
		},
	}, nil
}

// NewWriteSingleCoilRequestRTU creates new instance of Write Single Coil RTU request
func NewWriteSingleCoilRequestRTU(unitID uint8, address uint16, coilState bool) (*WriteSingleCoilRequestRTU, error) {
	return &WriteSingleCoilRequestRTU{
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			Address:   address,
			CoilState: coilState,
		},
	}, nil
}

// Bytes returns WriteSingleCoilRequestRTU packet as bytes form
func (r WriteSingleCoilRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.WriteSingleCoilRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r WriteSingleCoilRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 coils byte count + 2 coils data
	return 6
}

// ParseWriteSingleCoilRequestRTU parses given bytes into WriteSingleCoilRequestRTU
func ParseWriteSingleCoilRequestRTU(data []byte) (*WriteSingleCoilRequestRTU, error) {
	dLen := len(data)
	if dLen != 8 && dLen != 6 { // with or without CRC
		return nil, errors.New("received data length too short to be valid packet")
	}
	if data[1] != FunctionWriteSingleCoil {
		return nil, errors.New("received function code in packet is not 0x05")
	}
	coilStateRaw := binary.BigEndian.Uint16(data[4:6])
	if coilStateRaw != writeCoilOn && coilStateRaw != writeCoilOff {
		return nil, errors.New("coil state has invalid value")
	}
	return &WriteSingleCoilRequestRTU{
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID: data[0],
			// function code = data[7]
			Address:   binary.BigEndian.Uint16(data[2:4]),
			CoilState: coilStateRaw == writeCoilOn,
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteSingleCoilRequest) FunctionCode() uint8 {
	return FunctionWriteSingleCoil
}

// Bytes returns WriteSingleCoilRequest packet as bytes form
func (r WriteSingleCoilRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

const (
	writeCoilOn  = uint16(0xFF00)
	writeCoilOff = uint16(0x0000)
)

func (r WriteSingleCoilRequest) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteSingleCoil
	binary.BigEndian.PutUint16(bytes[2:4], r.Address)

	coilState := writeCoilOff
	if r.CoilState {
		coilState = writeCoilOn
	}
	binary.BigEndian.PutUint16(bytes[4:6], coilState)
	return bytes
}
