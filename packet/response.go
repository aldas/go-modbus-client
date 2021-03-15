package packet

import (
	"errors"
	"fmt"
)

// Response is common interface of modbus response packets
type Response interface {
	// FunctionCode returns function code of this request
	FunctionCode() uint8
	// Bytes returns packet as bytes form
	Bytes() []byte
}

// ParseTCPResponse parses given bytes into modbus TCP response packet or into ErrorResponseTCP or returns error
func ParseTCPResponse(data []byte) (Response, error) {
	if len(data) < 9 {
		return nil, errors.New("data is too short to be a Modbus TCP packet")
	}
	if err := AsTCPErrorPacket(data); err != nil {
		return nil, err
	}

	functionCode := data[7]
	switch functionCode {
	case FunctionReadCoils: // 0x01
		return ParseReadCoilsResponseTCP(data)
	case FunctionReadDiscreteInputs: // 0x02
		return ParseReadDiscreteInputsResponseTCP(data)
	case FunctionReadHoldingRegisters: // 0x03
		return ParseReadHoldingRegistersResponseTCP(data)
	case FunctionReadInputRegisters: // 0x04
		return ParseReadInputRegistersResponseTCP(data)
	case FunctionWriteSingleCoil: // 0x05
		return ParseWriteSingleCoilResponseTCP(data)
	case FunctionWriteSingleRegister: // 0x06
		return ParseWriteSingleRegisterResponseTCP(data)
	case FunctionWriteMultipleCoils: // 0x0f
		return ParseWriteMultipleCoilsResponseTCP(data)
	case FunctionWriteMultipleRegisters: // 0x10
		return ParseWriteMultipleRegistersResponseTCP(data)
	default:
		return nil, fmt.Errorf("unknown function code parsed: %v", functionCode)
	}
}

// ParseRTUResponse parses given bytes into modbus RTU response packet or into ErrorResponseRTU or returns error
func ParseRTUResponse(data []byte) (Response, error) {
	if len(data) < 5 {
		return nil, errors.New("data is too short to be a Modbus RTU packet")
	}
	if err := AsRTUErrorPacket(data); err != nil {
		return nil, err
	}

	functionCode := data[1]
	switch functionCode {
	case FunctionReadCoils: // 0x01
		return ParseReadCoilsResponseRTU(data)
	case FunctionReadDiscreteInputs: // 0x02
		return ParseReadDiscreteInputsResponseRTU(data)
	case FunctionReadHoldingRegisters: // 0x03
		return ParseReadHoldingRegistersResponseRTU(data)
	case FunctionReadInputRegisters: // 0x04
		return ParseReadInputRegistersResponseRTU(data)
	case FunctionWriteSingleCoil: // 0x05
		return ParseWriteSingleCoilResponseRTU(data)
	case FunctionWriteSingleRegister: // 0x06
		return ParseWriteSingleRegisterResponseRTU(data)
	case FunctionWriteMultipleCoils: // 0x0f
		return ParseWriteMultipleCoilsResponseRTU(data)
	case FunctionWriteMultipleRegisters: // 0x10
		return ParseWriteMultipleRegistersResponseRTU(data)
	default:
		return nil, fmt.Errorf("unknown function code parsed: %v", functionCode)
	}
}

// isBitSet checks if N-th bit is set in data. NB: Bits are counted from `startBit` and right to left.
func isBitSet(data []byte, startBit uint16, bit uint16) (bool, error) {
	targetBit := bit - startBit
	if bit < startBit {
		return false, errors.New("bit can not be before startBit")
	}
	if len(data)*8 <= int(targetBit) {
		return false, errors.New("bit value more than data contains bits")
	}
	nThByte := len(data) - 1 - int(targetBit/8)
	nThBit := targetBit % 8
	b := data[nThByte]
	return b&(1<<nThBit) != 0, nil
}
