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
)

const (
	// FunctionReadCoils is function code for Read Coils (FC01)
	FunctionReadCoils = uint8(1)
	// FunctionReadDiscreteInputs is function code for Read Discrete Inputs (FC02)
	FunctionReadDiscreteInputs = uint8(2)
	// FunctionReadHoldingRegisters is function code for Read Holding Registers (FC03)
	FunctionReadHoldingRegisters = uint8(3)
	// FunctionReadInputRegisters is function code for Read Input Registers (FC04)
	FunctionReadInputRegisters = uint8(4)
	// FunctionWriteSingleCoil is function code for Write Single Coil (FC05)
	FunctionWriteSingleCoil = uint8(5)
	// FunctionWriteSingleRegister is function code for Write Single Register (FC06)
	FunctionWriteSingleRegister = uint8(6)
	// FunctionWriteMultipleCoils is function code for Write Multiple Coils (FC15)
	FunctionWriteMultipleCoils = uint8(15)
	// FunctionWriteMultipleRegisters is function code for Write Multiple Registers (FC16)
	FunctionWriteMultipleRegisters = uint8(16)
	// FunctionReadWriteMultipleRegisters is function code for Read / Write Multiple Registers (FC23)
	FunctionReadWriteMultipleRegisters = uint8(23)
)

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

//// IsCompleteTCPPacket checks if packet is complete valid modbus TCP packet
//func IsCompleteTCPPacket(data []byte) bool {
//	// minimal amount is 9 bytes (header + function code + 1 byte of something ala error code)
//	packetLen := len(data)
//	if packetLen < 9 {
//		return false
//	}
//	// modbus header 6 bytes are = transaction id + protocol id + length of PDU part.
//	// so adding these number is what complete packet would be
//	expectedLength := int(6 + binary.BigEndian.Uint16(data[4:5]))
//	if packetLen > expectedLength {
//		return true // this is error situation. we received more bytes than length in packet indicates
//	}
//
//	return packetLen == expectedLength
//}
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
