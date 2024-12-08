package packet

import (
	"math/rand/v2"
)

// ReadServerIDRequestTCP is TCP Request for Read Server ID function (FC=17, 0x11)
//
// Example packet:  0x81 0x80 0x00 0x00 0x00 0x02 0x10 0x11
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x02 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x10 - unit id (6)
// 0x11 - function code (7)
type ReadServerIDRequestTCP struct {
	MBAPHeader
	ReadServerIDRequest
}

// ReadServerIDRequestRTU is RTU Request for Read Server ID function (FC=17, 0x11)
//
// Example packet:  0x10 0x11 0xcc 0x7c
// 0x10 - unit id (0)
// 0x11 - function code (1)
// 0xcc 0x7c - CRC16 (6,7)
type ReadServerIDRequestRTU struct {
	ReadServerIDRequest
}

// ReadServerIDRequest is Request for Read Server ID function (FC=17, 0x11)
type ReadServerIDRequest struct {
	UnitID uint8
}

// NewReadServerIDRequestTCP creates new instance of Read Server ID TCP request
func NewReadServerIDRequestTCP(unitID uint8) (*ReadServerIDRequestTCP, error) {
	return &ReadServerIDRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 1 + rand.N(uint16(65534)), // #nosec G404
			ProtocolID:    0,
		},
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: unitID,
		},
	}, nil
}

// Bytes returns ReadServerIDRequestTCP packet as bytes form
func (r ReadServerIDRequestTCP) Bytes() []byte {
	length := uint16(2)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadServerIDRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadServerIDRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitID + 1 fc + 1 byte count + N (unknown amount of bytes, at least 1)
	return 6 + 4 // at least this amount
}

// ParseReadServerIDRequestTCP parses given bytes into ReadServerIDRequestTCP
func ParseReadServerIDRequestTCP(data []byte) (*ReadServerIDRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	unitID := data[6]
	if data[7] != FunctionReadServerID {
		tmpErr := NewErrorParseTCP(ErrIllegalFunction, "received function code in packet is not 0x11")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadServerID
		return nil, tmpErr
	}
	return &ReadServerIDRequestTCP{
		MBAPHeader: header,
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: unitID,
		},
	}, nil
}

// NewReadServerIDRequestRTU creates new instance of Read Server ID RTU request
func NewReadServerIDRequestRTU(unitID uint8) (*ReadServerIDRequestRTU, error) {
	return &ReadServerIDRequestRTU{
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: unitID,
		},
	}, nil
}

// Bytes returns ReadServerIDRequestRTU packet as bytes form
func (r ReadServerIDRequestRTU) Bytes() []byte {
	result := make([]byte, 2+2)
	bytes := r.ReadServerIDRequest.bytes(result)
	crc := CRC16(bytes[:2])
	result[2] = uint8(crc)
	result[3] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadServerIDRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 1 byte count for server id + 1 run status + 2 CRC
	return 2 + 2 + 2 // at least this amount
}

// ParseReadServerIDRequestRTU parses given bytes into ReadServerIDRequestRTU
// Does not check CRC
func ParseReadServerIDRequestRTU(data []byte) (*ReadServerIDRequestRTU, error) {
	dLen := len(data)
	if dLen != 4 && dLen != 2 { // with or without CRC bytes
		return nil, NewErrorParseRTU(ErrServerFailure, "invalid data length to be valid packet")
	}
	unitID := data[0]
	if data[1] != FunctionReadServerID {
		tmpErr := NewErrorParseRTU(ErrIllegalFunction, "received function code in packet is not 0x11")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadServerID
		return nil, tmpErr
	}
	return &ReadServerIDRequestRTU{
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: unitID,
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadServerIDRequest) FunctionCode() uint8 {
	return FunctionReadServerID
}

// Bytes returns ReadServerIDRequest packet as bytes form
func (r ReadServerIDRequest) Bytes() []byte {
	return r.bytes(make([]byte, 2))
}

func (r ReadServerIDRequest) bytes(bytes []byte) []byte {
	bytes[0] = r.UnitID
	bytes[1] = FunctionReadServerID
	return bytes
}
