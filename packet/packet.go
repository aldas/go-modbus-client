package packet

import (
	"encoding/binary"
	"errors"
)

const (
	tcpMBAPHeaderLen         = 6
	functionCodeErrorBitmask = uint8(128)

	// MaxRegistersInReadResponse is maximum quantity of registers that can be returned by read request (fc03/fc04)
	MaxRegistersInReadResponse = uint16(125)
	// MaxCoilsInReadResponse is maximum quantity of discretes/coils that can be returned by read request (fc01/fc02)
	MaxCoilsInReadResponse = uint16(2000) // 2000/8=250 bytes
)

const (
	// FunctionReadCoils is function code for Read Coils (FC01)
	FunctionReadCoils = uint8(1) // 0x01
	// FunctionReadDiscreteInputs is function code for Read Discrete Inputs (FC02)
	FunctionReadDiscreteInputs = uint8(2) // 0x02
	// FunctionReadHoldingRegisters is function code for Read Holding Registers (FC03)
	FunctionReadHoldingRegisters = uint8(3) // 0x03
	// FunctionReadInputRegisters is function code for Read Input Registers (FC04)
	FunctionReadInputRegisters = uint8(4) // 0x04
	// FunctionWriteSingleCoil is function code for Write Single Coil (FC05)
	FunctionWriteSingleCoil = uint8(5) // 0x05
	// FunctionWriteSingleRegister is function code for Write Single Register (FC06)
	FunctionWriteSingleRegister = uint8(6) // 0x06
	// FunctionWriteMultipleCoils is function code for Write Multiple Coils (FC15)
	FunctionWriteMultipleCoils = uint8(15) // 0x0f
	// FunctionWriteMultipleRegisters is function code for Write Multiple Registers (FC16)
	FunctionWriteMultipleRegisters = uint8(16) // 0x10
	// FunctionReadServerID is function code for Read Server ID (FC16)
	FunctionReadServerID = uint8(17) // 0x11
	// FunctionReadWriteMultipleRegisters is function code for Read / Write Multiple Registers (FC23)
	FunctionReadWriteMultipleRegisters = uint8(23) // 0x17
)

var supportedFunctionCodes = [10]byte{
	FunctionReadCoils,
	FunctionReadDiscreteInputs,
	FunctionReadHoldingRegisters,
	FunctionReadInputRegisters,
	FunctionWriteSingleCoil,
	FunctionWriteSingleRegister,
	FunctionWriteMultipleCoils,
	FunctionWriteMultipleRegisters,
	FunctionReadServerID,
	FunctionReadWriteMultipleRegisters,
}

// MBAPHeader (Modbus Application Header) is header part of modbus TCP packet. NB: this library does pack unitID into header
type MBAPHeader struct {
	TransactionID uint16
	ProtocolID    uint16
}

func (h MBAPHeader) bytes(dst []byte, length uint16) {
	binary.BigEndian.PutUint16(dst[0:2], h.TransactionID)
	binary.BigEndian.PutUint16(dst[2:4], 0x0000)
	binary.BigEndian.PutUint16(dst[4:6], length)
}

// ParseMBAPHeader parses MBAPHeader from given bytes
func ParseMBAPHeader(data []byte) (MBAPHeader, error) {
	if len(data) < 6 {
		return MBAPHeader{}, NewErrorParseTCP(ErrServerFailure, "data to short to contain MBAPHeader")
	}
	if data[2] != 0x0 || data[3] != 0x00 {
		return MBAPHeader{}, NewErrorParseTCP(ErrServerFailure, "invalid protocol id")
	}
	pduLen := binary.BigEndian.Uint16(data[4:6])
	if pduLen == 0 {
		return MBAPHeader{}, NewErrorParseTCP(ErrServerFailure, "pdu length in header can not be 0")
	}
	if len(data) != 6+int(pduLen) {
		return MBAPHeader{}, NewErrorParseTCP(ErrServerFailure, "packet length does not match length in header")
	}
	return MBAPHeader{
		TransactionID: binary.BigEndian.Uint16(data[0:2]),
		ProtocolID:    0,
	}, nil
}

var (
	// ErrTCPDataTooShort is returned when received data is still too short to be actual Modbus TCP packet.
	ErrTCPDataTooShort = NewErrorParseTCP(ErrUnknown, "data is too short to be a Modbus TCP packet")
	// ErrIsNotTCPPacket is returned when received data does not look like Modbus TCP packet
	ErrIsNotTCPPacket = NewErrorParseTCP(ErrUnknown, "data does not like Modbus TCP packet")
)

// LooksLikeModbusTCP checks if given data starts with bytes that could be potentially parsed as Modbus TCP packet.
func LooksLikeModbusTCP(data []byte, allowUnSupportedFunctionCodes bool) (expectedLen int, error error) {
	// Example of first 8 bytes
	// 0x81 0x80 - transaction id (0,1)
	// 0x00 0x00 - protocol id (2,3)
	// 0x00 0x02 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
	// 0x10 - unit id (6)
	// 0x11 - function code (7)

	// minimal amount is 8 bytes (header + unit id + function code)
	if len(data) < 8 {
		return 0, ErrTCPDataTooShort
	}
	if !(data[2] == 0x0 && data[3] == 0x0) { // check protocol id
		return 0, ErrIsNotTCPPacket
	}
	pduLen := binary.BigEndian.Uint16(data[4:6]) // number of bytes in the message to follow
	if pduLen < 3 {                              // every request is more than 2 bytes of PDU
		return 0, ErrIsNotTCPPacket
	}
	functionCode := data[7] // function code
	if functionCode == 0 {
		return 0, ErrIsNotTCPPacket
	}
	expectedLen = int(pduLen) + 6
	if allowUnSupportedFunctionCodes {
		return expectedLen, nil
	}
	for _, fc := range supportedFunctionCodes {
		if fc == functionCode {
			return expectedLen, nil
		}
	}
	return expectedLen, &ErrorParseTCP{
		Message: "unsupported function code",
		Packet: ErrorResponseTCP{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			UnitID:        data[6],
			Function:      functionCode,
			Code:          ErrIllegalFunction,
		},
	}
}

//
//// IsCompleteRTUPacket checks if packet is complete valid modbus RTU packet
//func IsCompleteRTUPacket(data []byte) bool {
//	// minimal amount is 5 bytes (1 byte unitID + 1 byte function code + 1 byte of something ala error code + 2 bytes of crc)
//	packetLen := len(data)
//	if packetLen < 5 {
//		return false
//	}
//	// should check crc16 (or not)
//	// deduce necessary bytes from function code if crc is valid?
//	return true
//}

func putReadRequestBytes(dst []byte, unitID uint8, functionCode uint8, startAddress uint16, quantity uint16) {
	dst[0] = unitID
	dst[1] = functionCode
	binary.BigEndian.PutUint16(dst[2:4], startAddress)
	binary.BigEndian.PutUint16(dst[4:6], quantity)
}

// ErrInvalidCRC is error returned when packet data does not match its CRC value
var ErrInvalidCRC = errors.New("packet cyclic redundancy check does not match Modbus RTU packet bytes")

// CRC16 calculates 16 bit cyclic redundancy check (CRC) for given bytes
// Note about the CRC:
//
// Polynomial: x16 + x15 + x2 + 1 (CRC-16-ANSI also known as CRC-16-IBM, normal hexadecimal algebraic polynomial being 8005 and reversed A001).
// Initial value: 65,535.
// Example of frame in hexadecimal: 01 04 02 FF FF B8 80 (CRC-16-ANSI calculation from 01 to FF gives 80B8, which is transmitted least significant byte first).
func CRC16(data []byte) uint16 {
	crc := uint16(0xffff)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
