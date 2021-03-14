package packet

import (
	"encoding/binary"
	"errors"
)

// WriteSingleCoilResponseTCP is TCP Response for Write Single Coil (FC=05)
//
// Data part of packet is always 4 bytes - 2 byte for address and 2 byte for coil status (FF00 = on,  0000 = off).
// For example: coil at address 1 is turned on '0x00 0x01 0xFF 0x00'
// For example: coil at address 10 is turned off '0x00 0x0A 0x00 0x00'
//
// Example packet: 0x00 0x01 0x00 0x00 0x00 0x06 0x03 0x05 0x00 0x02 0xFF 0x00
// 0x00 0x01 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x03 - unit id (6)
// 0x05 - function code (7)
// 0x00 0x02 - start address (8,9)
// 0xFF 0x00 - coil data (true) (10,11)
type WriteSingleCoilResponseTCP struct {
	MBAPHeader
	WriteSingleCoilResponse
}

// WriteSingleCoilResponseRTU is RTU Response for Write Single Coil (FC=05)
//
// Data part of packet is always 4 bytes - 2 byte for address and 2 byte for coil status (FF00 = on,  0000 = off).
// For example: coil at address 1 is turned on '0x00 0x01 0xFF 0x00'
// For example: coil at address 10 is turned off '0x00 0x0A 0x00 0x00'
//
// Example packet: 0x03 0x05 0x00 0x02 0xFF 0x00 0xFF 0xFF
// 0x03 - unit id (0)
// 0x05 - function code (1)
// 0x00 0x02 - start address (2,3)
// 0xFF 0x00 - coil data (true) (4,5)
// 0xFF 0xFF - CRC16 (6,7) // FIXME: add correct crc value example
type WriteSingleCoilResponseRTU struct {
	WriteSingleCoilResponse
}

// WriteSingleCoilResponse is Response for Write Single Coil (FC=05)
type WriteSingleCoilResponse struct {
	UnitID       uint8
	StartAddress uint16
	CoilState    bool
}

// Bytes returns WriteSingleCoilResponseTCP packet as bytes form
func (r WriteSingleCoilResponseTCP) Bytes() []byte {
	result := make([]byte, tcpMBAPHeaderLen+6)
	r.MBAPHeader.bytes(result[0:6])
	r.WriteSingleCoilResponse.bytes(result[6:])
	return result
}

// ParseWriteSingleCoilResponseTCP parses given bytes into ParseWriteSingleCoilResponseTCP
func ParseWriteSingleCoilResponseTCP(data []byte) (*WriteSingleCoilResponseTCP, error) {
	dLen := len(data)
	if dLen < 12 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	pduLen := binary.BigEndian.Uint16(data[4:6])
	if dLen != 6+int(pduLen) {
		return nil, errors.New("received data length does not match PDU len in packet")
	}
	return &WriteSingleCoilResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
			Length:        binary.BigEndian.Uint16(data[4:6]),
		},
		WriteSingleCoilResponse: WriteSingleCoilResponse{
			UnitID:       data[6],
			StartAddress: binary.BigEndian.Uint16(data[8:10]),
			CoilState:    binary.BigEndian.Uint16(data[10:12]) == writeCoilOn, // FIXME: validate?
		},
	}, nil
}

// Bytes returns WriteSingleCoilResponseRTU packet as bytes form
func (r WriteSingleCoilResponseRTU) Bytes() []byte {
	result := make([]byte, 8)
	bytes := r.WriteSingleCoilResponse.bytes(result)
	binary.BigEndian.PutUint16(result[6:8], CRC16(bytes))
	return result
}

// ParseWriteSingleCoilResponseRTU parses given bytes into WriteSingleCoilResponseRTU
func ParseWriteSingleCoilResponseRTU(data []byte) (*WriteSingleCoilResponseRTU, error) {
	dLen := len(data)
	if dLen < 8 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	if dLen > 8 {
		return nil, errors.New("received data length too long to be valid packet")
	}
	// FIXME: check CRC
	return &WriteSingleCoilResponseRTU{
		WriteSingleCoilResponse: WriteSingleCoilResponse{
			UnitID: data[0],
			// data[1] function code
			StartAddress: binary.BigEndian.Uint16(data[2:4]),
			CoilState:    binary.BigEndian.Uint16(data[4:6]) == writeCoilOn, // FIXME: validate?
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r WriteSingleCoilResponse) FunctionCode() uint8 {
	return FunctionWriteSingleCoil
}

// Bytes returns WriteSingleCoilResponse packet as bytes form
func (r WriteSingleCoilResponse) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r WriteSingleCoilResponse) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionWriteSingleCoil
	binary.BigEndian.PutUint16(bytes[2:4], r.StartAddress)

	coilState := writeCoilOff
	if r.CoilState {
		coilState = writeCoilOn
	}
	binary.BigEndian.PutUint16(bytes[4:6], coilState)
	return bytes
}
