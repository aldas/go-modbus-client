package packet

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
)

// ReadCoilsRequestTCP is TCP Request for Read Coils function (FC=01)
//
// Example packet:  0x81 0x80 0x00 0x00 0x00 0x06 0x10 0x01 0x00 0x6B 0x00 0x03
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x10 - unit id (6)
// 0x01 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0x00 0x03 - coils quantity to return (10,11)
type ReadCoilsRequestTCP struct {
	MBAPHeader
	ReadCoilsRequest
}

// ReadCoilsRequestRTU is RTU Request for Read Coils function (FC=01)
//
// Example packet:  0x10 0x01 0x00 0x6B 0x00 0x03 0x0e 0x96
// 0x10 - unit id (0)
// 0x01 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x00 0x03 - coils quantity to return (4,5)
// 0x0e 0x96 - CRC16 (6,7)
type ReadCoilsRequestRTU struct {
	ReadCoilsRequest
}

// ReadCoilsRequest is Request for Read Coils function (FC=01)
type ReadCoilsRequest struct {
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16
}

// NewReadCoilsRequestTCP creates new instance of Read Coils TCP request
func NewReadCoilsRequestTCP(unitID uint8, startAddress uint16, quantity uint16) (*ReadCoilsRequestTCP, error) {
	if quantity == 0 || quantity > 2000 {
		// 2000 coils is due that in response data size field is 1 byte so max 250*8=2000 coils can be returned
		return nil, fmt.Errorf("quantity is out of range (1-2000): %v", quantity)
	}

	return &ReadCoilsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadCoilsRequestTCP packet as bytes form
func (r ReadCoilsRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadCoilsRequest.bytes(result[6 : 6+length])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadCoilsRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitID + 1 fc + 1 coil byte count + N data len (1-256)
	return 6 + 3 + r.coilByteLength()
}

// ParseReadCoilsRequestTCP parses given bytes into ReadCoilsRequestTCP
func ParseReadCoilsRequestTCP(data []byte) (*ReadCoilsRequestTCP, error) {
	header, err := ParseMBAPHeader(data)
	if err != nil {
		return nil, err
	}
	unitID := data[6]
	if data[7] != FunctionReadCoils {
		tmpErr := NewErrorParseTCP(ErrIllegalFunction, "received function code in packet is not 0x01")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadCoils
		return nil, tmpErr
	}
	quantity := binary.BigEndian.Uint16(data[10:12])
	if !(quantity >= 1 && quantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseTCP(ErrIllegalDataValue, "invalid quantity. valid range 1..125")
		tmpErr.Packet.TransactionID = header.TransactionID
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadCoils
		return nil, tmpErr
	}
	return &ReadCoilsRequestTCP{
		MBAPHeader: header,
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID: unitID,
			// function code = data[7]
			StartAddress: binary.BigEndian.Uint16(data[8:10]),
			Quantity:     quantity,
		},
	}, nil
}

// NewReadCoilsRequestRTU creates new instance of Read Coils RTU request
func NewReadCoilsRequestRTU(unitID uint8, startAddress uint16, quantity uint16) (*ReadCoilsRequestRTU, error) {
	if quantity == 0 || quantity > 2000 {
		// 2000 coils is due that in response data size field is 1 byte so max 250*8=2000 coils can be returned
		return nil, fmt.Errorf("quantity is out of range (1-2000): %v", quantity)
	}

	return &ReadCoilsRequestRTU{
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadCoilsRequestRTU packet as bytes form
func (r ReadCoilsRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.ReadCoilsRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadCoilsRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 coils byte count + N coils data
	return 4 + r.coilByteLength()
}

// ParseReadCoilsRequestRTU parses given bytes into ReadCoilsRequestRTU
// Does not check CRC
func ParseReadCoilsRequestRTU(data []byte) (*ReadCoilsRequestRTU, error) {
	dLen := len(data)
	if dLen != 8 && dLen != 6 { // with or without CRC bytes
		return nil, NewErrorParseRTU(ErrServerFailure, "invalid data length to be valid packet")
	}
	unitID := data[0]
	if data[1] != FunctionReadCoils {
		tmpErr := NewErrorParseRTU(ErrIllegalFunction, "received function code in packet is not 0x01")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadCoils
		return nil, tmpErr
	}
	quantity := binary.BigEndian.Uint16(data[4:6])
	if !(quantity >= 1 && quantity <= 125) { // 0x0001 to 0x007D
		tmpErr := NewErrorParseRTU(ErrIllegalDataValue, "invalid quantity. valid range 1..125")
		tmpErr.Packet.UnitID = unitID
		tmpErr.Packet.Function = FunctionReadCoils
		return nil, tmpErr
	}
	return &ReadCoilsRequestRTU{
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID: unitID,
			// function code = data[1]
			StartAddress: binary.BigEndian.Uint16(data[2:4]),
			Quantity:     quantity,
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadCoilsRequest) FunctionCode() uint8 {
	return FunctionReadCoils
}

func (r ReadCoilsRequest) coilByteLength() int {
	return int(math.Ceil(float64(r.Quantity) / 8))
}

// Bytes returns ReadCoilsRequest packet as bytes form
func (r ReadCoilsRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r ReadCoilsRequest) bytes(bytes []byte) []byte {
	putReadRequestBytes(bytes, r.UnitID, FunctionReadCoils, r.StartAddress, r.Quantity)
	return bytes
}
