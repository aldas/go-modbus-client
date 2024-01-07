package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Request is common interface of modbus request packets
type Request interface {
	// FunctionCode returns function code of this request
	FunctionCode() uint8
	// Bytes returns packet as bytes form
	Bytes() []byte
	// ExpectedResponseLength returns length of bytes that valid response to this request would be
	ExpectedResponseLength() int
}

// ParseTCPRequest parses given bytes into modbus TCP request packet or returns error
func ParseTCPRequest(data []byte) (Request, error) {
	if len(data) < 8 {
		return nil, ErrTCPDataTooShort
	}
	functionCode := data[7]
	switch functionCode {
	case FunctionReadCoils: // 0x01
		return ParseReadCoilsRequestTCP(data)
	case FunctionReadDiscreteInputs: // 0x02
		return ParseReadDiscreteInputsRequestTCP(data)
	case FunctionReadHoldingRegisters: // 0x03
		return ParseReadHoldingRegistersRequestTCP(data)
	case FunctionReadInputRegisters: // 0x04
		return ParseReadInputRegistersRequestTCP(data)
	case FunctionWriteSingleCoil: // 0x05
		return ParseWriteSingleCoilRequestTCP(data)
	case FunctionWriteSingleRegister: // 0x06
		return ParseWriteSingleRegisterRequestTCP(data)
	case FunctionWriteMultipleCoils: // 0x0f
		return ParseWriteMultipleCoilsRequestTCP(data)
	case FunctionWriteMultipleRegisters: // 0x10
		return ParseWriteMultipleRegistersRequestTCP(data)
	case FunctionReadServerID: // 0x11
		return ParseReadServerIDRequestTCP(data)
	case FunctionReadWriteMultipleRegisters: // 0x17
		return ParseReadWriteMultipleRegistersRequestTCP(data)
	default:
		return nil, NewErrorParseTCP(ErrIllegalFunction, fmt.Sprintf("unknown function code parsed: %v", functionCode))
	}
}

// ParseRTURequestWithCRC checks packet CRC and parses given bytes into modbus RTU request packet or returns error
func ParseRTURequestWithCRC(data []byte) (Response, error) {
	dataLen := len(data)
	if dataLen < 4 {
		return nil, errors.New("data is too short to be a Modbus RTU packet")
	}
	packetCRC := binary.LittleEndian.Uint16(data[dataLen-2:])
	actualCRC := CRC16(data[:dataLen-2])
	if packetCRC != actualCRC {
		return nil, ErrInvalidCRC
	}
	return ParseRTURequest(data)
}

// ParseRTURequest parses given bytes into modbus RTU request packet or returns error
// Does not check CRC.
func ParseRTURequest(data []byte) (Request, error) {
	if len(data) < 4 {
		return nil, errors.New("data is too short to be a Modbus RTU packet")
	}
	functionCode := data[1]
	switch functionCode {
	case FunctionReadCoils: // 0x01
		return ParseReadCoilsRequestRTU(data)
	case FunctionReadDiscreteInputs: // 0x02
		return ParseReadDiscreteInputsRequestRTU(data)
	case FunctionReadHoldingRegisters: // 0x03
		return ParseReadHoldingRegistersRequestRTU(data)
	case FunctionReadInputRegisters: // 0x04
		return ParseReadInputRegistersRequestRTU(data)
	case FunctionWriteSingleCoil: // 0x05
		return ParseWriteSingleCoilRequestRTU(data)
	case FunctionWriteSingleRegister: // 0x06
		return ParseWriteSingleRegisterRequestRTU(data)
	case FunctionWriteMultipleCoils: // 0x0f
		return ParseWriteMultipleCoilsRequestRTU(data)
	case FunctionWriteMultipleRegisters: // 0x10
		return ParseWriteMultipleRegistersRequestRTU(data)
	case FunctionReadServerID: // 0x11
		return ParseReadServerIDRequestRTU(data)
	case FunctionReadWriteMultipleRegisters: // 0x17
		return ParseReadWriteMultipleRegistersRequestRTU(data)
	default:
		return nil, fmt.Errorf("unknown function code parsed: %v", functionCode)
	}
}
