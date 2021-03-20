package packet

import (
	"encoding/binary"
	"fmt"
)

func errorText(code uint8) string {
	switch code {
	case 1:
		// The function code received in the query is not an allowable action for the server. This may be
		// because the function code is only applicable to newer devices, and was not implemented in the
		// unit selected. It could also indicate that the server is in the wrong state to process a request of this
		// type, for example because it is unconfigured and is being asked to return register values.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Illegal function"
	case 2:
		// The data address received in the query is not an allowable address for the server. More specifically,
		// the combination of reference number and transfer length is invalid. For a controller with 100 registers,
		// the PDU addresses the first register as 0, and the last one as 99. If a request is submitted with a
		// starting register address of 96 and a quantity of registers of 4, then this request will successfully
		// operate (address-wise at least) on registers 96, 97, 98, 99. If a request is submitted with a starting
		// register address of 96 and a quantity of registers of 5, then this request will fail with Exception
		// Code 0x02 “Illegal Data Address” since it attempts to operate on registers 96, 97, 98, 99 and 100, and
		// there is no register with address 100.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Illegal data address"
	case 3:
		// A value contained in the query data field is not an allowable value for server. This indicates a fault in
		// the structure of the remainder of a complex request, such as that the implied length is incorrect.
		// It specifically does NOT mean that a data item submitted for storage in a register has a value outside
		// the expectation of the application program, since the MODBUS protocol is unaware of the significance of any
		// particular value of any particular register.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Illegal data value"
	case 4:
		// An unrecoverable error occurred while the server was attempting to perform the requested action.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Server failure"
	case 5:
		// Specialized use in conjunction with programming commands.
		// The server has accepted the request and is processing it, but a long duration of time will be required to
		// do so. This response is returned to prevent a timeout error from occurring in the client. The client can
		// next issue a Poll Program Complete message to determine if processing is completed.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Acknowledge"
	case 6:
		// Specialized use in conjunction with programming commands.
		// The server is engaged in processing a long duration program command. The client should retransmit the
		// message later when the server is free.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Server busy"
	case 8:
		// Specialized use in conjunction with function codes 20 and 21 and reference type 6, to indicate that the
		// extended file area failed to pass a consistency check.
		// The server attempted to read record file, but detected a parity error in the memory. The client can retry
		// the request, but service may be required on the server device.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 48
		return "Memory parity error"
	case 10:
		// Specialized use in conjunction with gateways, indicates that the gateway was unable to allocate an internal
		// communication path from the input port to the output port for processing the request. Usually means that
		// the gateway is misconfigured or overloaded.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 49
		return "Gateway path unavailable"
	case 11:
		// Specialized use in conjunction with gateways, indicates that no response was obtained from the target device.
		// Usually means that the device is not present on the network.
		// Quote from: `MODBUS Application Protocol Specification V1.1b3`, page 49
		return "Gateway targeted device failed to respond"
	default:
		return fmt.Sprintf("Unknown error code: %v", code)
	}
}

// ErrorResponseTCP is TCP error response send by server to client
type ErrorResponseTCP struct {
	TransactionID uint16
	UnitID        uint8
	Function      uint8
	Code          uint8
}

// Error translates error code to error message.
func (re ErrorResponseTCP) Error() string {
	return errorText(re.Code)
}

// Bytes returns ErrorResponseTCP packet as bytes form
func (re ErrorResponseTCP) Bytes() []byte {
	result := make([]byte, 9)

	binary.BigEndian.PutUint16(result[0:2], re.TransactionID)
	binary.BigEndian.PutUint16(result[2:4], 0)
	binary.BigEndian.PutUint16(result[4:6], 3)
	result[6] = re.UnitID
	result[7] = re.Function + functionCodeErrorBitmask
	result[8] = re.Code

	return result
}

// FunctionCode returns function code to which error response originates from / was responded to
func (re ErrorResponseTCP) FunctionCode() uint8 {
	return re.Function
}

// ErrorResponseRTU is RTU error response send by server to client
type ErrorResponseRTU struct {
	UnitID   uint8
	Function uint8
	Code     uint8
}

// Error translates error code to error message.
func (re ErrorResponseRTU) Error() string {
	return errorText(re.Code)
}

// Bytes returns ErrorResponseRTU packet as bytes form
func (re ErrorResponseRTU) Bytes() []byte {
	result := make([]byte, 5)

	result[0] = re.UnitID
	result[1] = re.Function + functionCodeErrorBitmask
	result[2] = re.Code
	crc := CRC16(result[0:3])
	result[3] = uint8(crc)
	result[4] = uint8(crc >> 8)

	return result
}

// FunctionCode returns function code to which error response originates from / was responded to
func (re ErrorResponseRTU) FunctionCode() uint8 {
	return re.Function
}

// AsTCPErrorPacket converts raw packet bytes to Modbus TCP error response if possible
//
// Example packet: 0xda 0x87 0x00 0x00 0x00 0x03 0x01 0x81 0x03
// 0xda 0x87 - transaction id (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x03 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x01 - unit id (6)
// 0x81 - function code + 128 (error bitmask) (7)
// 0x03 - error code (8)
func AsTCPErrorPacket(data []byte) error {
	// a) data is too short. can not determine packet.
	// b) data is too long. can not be an error packet
	// Actual packet is at least 9 bytes. 7 bytes for Modbus TCP header and at least 2 bytes for PDU
	if len(data) != 9 {
		return nil
	}
	errorFunctionCode := data[7] & functionCodeErrorBitmask
	if errorFunctionCode != 0 {
		return &ErrorResponseTCP{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			UnitID:        data[6],
			Function:      data[7] - functionCodeErrorBitmask,
			Code:          data[8],
		}
	}
	return nil // probably start of valid packet
}

// AsRTUErrorPacket converts raw packet bytes to Modbus RTU error response if possible
//
// Example packet: 0x0a 0x81 0x02 0xb0 0x53
// 0x0a - unit id (0)
// 0x81 - function code + 128 (error bitmask) (1)
// 0x02 - error code (2)
// 0xb0 0x53 - CRC (3,4)
func AsRTUErrorPacket(data []byte) error {
	// a) data is too short. can not determine packet.
	// b) data is too long. can not be an error packet
	// Actual packet is at least 5 bytes. 1 unitID + 1 function code + 1 error code + 2 crc bytes
	if len(data) != 5 {
		return nil
	}
	errorFunctionCode := data[1] & functionCodeErrorBitmask
	if errorFunctionCode != 0 {
		return &ErrorResponseRTU{
			UnitID:   data[0],
			Function: data[1] - functionCodeErrorBitmask,
			Code:     data[2],
		}
	}
	return nil // probably start of valid packet
}
