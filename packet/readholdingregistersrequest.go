package packet

import (
	"fmt"
	"math/rand"
)

// ReadHoldingRegistersRequestTCP is TCP Request for Read Holding Registers (FC=03)
//
// Example packet: 0x00 0x01 0x00 0x00 0x00 0x06 0x01 0x03 0x00 0x6B 0x00 0x01
// 0x00 0x01 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x06 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x01 - unit id (6)
// 0x03 - function code (7)
// 0x00 0x6B - start address (8,9)
// 0x00 0x01 - holding registers quantity to return (10,11)
type ReadHoldingRegistersRequestTCP struct {
	MBAPHeader
	ReadHoldingRegistersRequest
}

// ReadHoldingRegistersRequestRTU is RTU Request for Read Holding Registers (FC=03)
//
// Example packet: 0x01 0x03 0x00 0x6B 0x00 0x01 0xFF 0xFF
// 0x01 - unit id (0)
// 0x03 - function code (1)
// 0x00 0x6B - start address (2,3)
// 0x00 0x01 - holding registers quantity to return (4,5)
// 0xFF 0xFF - CRC16 (6,7) // FIXME: add correct crc value example
type ReadHoldingRegistersRequestRTU struct {
	ReadHoldingRegistersRequest
}

// ReadHoldingRegistersRequest is Request for Read Holding Registers (FC=03)
type ReadHoldingRegistersRequest struct {
	UnitID       uint8
	StartAddress uint16
	Quantity     uint16
}

// NewReadHoldingRegistersRequestTCP creates new instance of Read Holding Registers TCP request
func NewReadHoldingRegistersRequestTCP(unitID uint8, startAddress uint16, quantity uint16) (*ReadHoldingRegistersRequestTCP, error) {
	if quantity == 0 || quantity > MaxRegistersInReadResponse {
		return nil, fmt.Errorf("quantity is out of range (1-125): %v", quantity)
	}

	return &ReadHoldingRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: uint16(1 + rand.Intn(65534)),
			ProtocolID:    0,
		},
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadHoldingRegistersRequestTCP packet as bytes form
func (r ReadHoldingRegistersRequestTCP) Bytes() []byte {
	length := uint16(6)
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadHoldingRegistersRequest.bytes(result[6:12])
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadHoldingRegistersRequestTCP) ExpectedResponseLength() int {
	// response = 6 header len + 1 unitid + 1 fc + 1 register byte count + N data len (2-256)
	return 6 + 3 + 2*int(r.Quantity)
}

// NewReadHoldingRegistersRequestRTU creates new instance of Read Holding Registers RTU request
func NewReadHoldingRegistersRequestRTU(unitID uint8, startAddress uint16, quantity uint16) (*ReadHoldingRegistersRequestRTU, error) {
	if quantity == 0 || quantity > MaxRegistersInReadResponse {
		return nil, fmt.Errorf("quantity is out of range (1-125): %v", quantity)
	}

	return &ReadHoldingRegistersRequestRTU{
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID: unitID,
			// function code is added by Bytes()
			StartAddress: startAddress,
			Quantity:     quantity,
		},
	}, nil
}

// Bytes returns ReadHoldingRegistersRequestRTU packet as bytes form
func (r ReadHoldingRegistersRequestRTU) Bytes() []byte {
	result := make([]byte, 6+2)
	bytes := r.ReadHoldingRegistersRequest.bytes(result)
	crc := CRC16(bytes[:6])
	result[6] = uint8(crc)
	result[7] = uint8(crc >> 8)
	return result
}

// ExpectedResponseLength returns length of bytes that valid response to this request would be
func (r ReadHoldingRegistersRequest) ExpectedResponseLength() int {
	// response = 1 UnitID + 1 functionCode + 2 register byte count + N register data
	return 4 + 2*int(r.Quantity)
}

// FunctionCode returns function code of this request
func (r ReadHoldingRegistersRequest) FunctionCode() uint8 {
	return FunctionReadHoldingRegisters
}

// Bytes returns ReadHoldingRegistersRequest packet as bytes form
func (r ReadHoldingRegistersRequest) Bytes() []byte {
	return r.bytes(make([]byte, 6))
}

func (r ReadHoldingRegistersRequest) bytes(bytes []byte) []byte {
	putReadRequestBytes(bytes, r.UnitID, FunctionReadHoldingRegisters, r.StartAddress, r.Quantity)
	return bytes
}
