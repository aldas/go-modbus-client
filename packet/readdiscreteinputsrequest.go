package packet

import (
	"fmt"
	"math"
	"math/rand"
)

// ReadDiscreteInputsRequestTCP is TCP Request for Read Discrete Inputs (FC=02)
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x06 0x10 0x02 0x00 0x6B 0x00 0x03
// 0x81 0x80 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x10 - unit id (6)
// 0x02 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0x00 0x03 - discrete inputs quantity to return (10,11)
type ReadDiscreteInputsRequestTCP struct {
	MBAPHeader
	ReadDiscreteInputsRequest
}

// ReadDiscreteInputsRequestRTU is RTU Request for Read Discrete Inputs (FC=02)
//
// Example packet: 0x10 0x02 0x00 0x6B 0x00 0x03 0xFF 0xFF
// 0x10 - unit id (0)
// 0x02 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x00 0x03 - discrete inputs quantity to return (4,5)
// 0xFF 0xFF - CRC16 (6,7) // FIXME: add correct crc value example
type ReadDiscreteInputsRequestRTU struct {
	ReadDiscreteInputsRequest
}

// ReadDiscreteInputsRequest is Request for Read Discrete Inputs (FC=02)
type ReadDiscreteInputsRequest struct {
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16
}

// NewReadDiscreteInputsRequestTCP creates new instance of Read Discrete Inputs TCP request
func NewReadDiscreteInputsRequestTCP(unitID uint8, startAddress uint16, quantity uint16) (*ReadDiscreteInputsRequestTCP, error) {
	if quantity == 0 || quantity > 2048 {
		// 2048 coils is due that in response data size field is 1 byte so max 256*8=2048 coils can be returned
		return nil, fmt.Errorf("quantity is out of range (1-2048): %v", quantity)
	}

	return &ReadDiscreteInputsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadDiscreteInputsRequestTCP packet as bytes form
func (r ReadDiscreteInputsRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadDiscreteInputsRequest.bytes(result[6:12])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadDiscreteInputsRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitid + 1 fc + 1 coil byte count + N data len (1-256)
	return 6 + 3 + r.coilByteLength()
}

// NewReadDiscreteInputsRequestRTU creates new instance of Read Discrete Inputs RTU request
func NewReadDiscreteInputsRequestRTU(unitID uint8, startAddress uint16, quantity uint16) (*ReadDiscreteInputsRequestRTU, error) {
	if quantity == 0 || quantity > 2048 {
		// 2048 coils is due that in response data size field is 1 byte so max 256*8=2048 coils can be returned
		return nil, fmt.Errorf("quantity is out of range (1-2048): %v", quantity)
	}

	return &ReadDiscreteInputsRequestRTU{
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadDiscreteInputsRequestRTU packet as bytes form
func (r ReadDiscreteInputsRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.ReadDiscreteInputsRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadDiscreteInputsRequestRTU) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 coils byte count + N coils data
	return 4 + r.coilByteLength()
}

// FunctionCode returns function code of this request
func (r ReadDiscreteInputsRequest) FunctionCode() uint8 {
	return FunctionReadDiscreteInputs
}

func (r ReadDiscreteInputsRequest) coilByteLength() int {
	return int(math.Ceil(float64(r.Quantity) / 8))
}

// Bytes returns ReadDiscreteInputsRequest packet as bytes form
func (r ReadDiscreteInputsRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r ReadDiscreteInputsRequest) bytes(bytes []byte) []byte {
	putReadRequestBytes(bytes, r.UnitID, FunctionReadDiscreteInputs, r.StartAddress, r.Quantity)
	return bytes
}
