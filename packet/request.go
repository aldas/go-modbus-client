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

// RequestDestination represents the addressing and request parameters
// used in a Modbus read/write operation. It identifies the target device,
// the function to execute, and the register range to read/write.
type RequestDestination struct {
	// Address specifies the target device's network or serial address.
	// For Modbus RTU, this is the filesystem path to device.
	// For Modbus TCP, this may represent the host or IP address.
	Address string

	// UnitID is the Modbus unit identifier used primarily in Modbus TCP
	// to address devices behind a gateway. In RTU mode, this is the slave address.
	UnitID uint8

	// FunctionCode specifies the Modbus function to execute.
	FunctionCode uint8

	// StartAddress defines the (starting) register or coil address
	// for the read/write operation.
	// Optional: This field is 0 when request does not have this field.
	StartAddress uint16

	// Quantity indicates how many registers or coils to read,
	// depending on the function code.
	// Optional: This field is 0 when request does not have this field.
	Quantity uint16
}

// ExtractRequestDestination extracts destination related fields from Modbus requests.
// Note: this function does not support FC23 (ReadWriteMultipleRegisters) fully, it is missing write related fields.
func ExtractRequestDestination(req Request) (RequestDestination, error) {
	d := RequestDestination{
		Address:      "", // up to the user to fill
		UnitID:       0,
		FunctionCode: req.FunctionCode(),
		StartAddress: 0,
		Quantity:     0,
	}
	switch r := req.(type) {
	case *ReadCoilsRequestTCP: // fc1
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadCoilsRequestRTU: // fc1
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadDiscreteInputsRequestTCP: // fc2
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadDiscreteInputsRequestRTU: // fc2
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadHoldingRegistersRequestTCP: // fc3
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadHoldingRegistersRequestRTU: // fc3
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadInputRegistersRequestTCP: // fc4
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *ReadInputRegistersRequestRTU: // fc4
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.Quantity
	case *WriteSingleCoilRequestRTU: // fc5
		d.UnitID = r.UnitID
		d.StartAddress = r.Address
	case *WriteSingleCoilRequestTCP: // fc5
		d.UnitID = r.UnitID
		d.StartAddress = r.Address
	case *WriteSingleRegisterRequestRTU: // fc6
		d.UnitID = r.UnitID
		d.StartAddress = r.Address
	case *WriteSingleRegisterRequestTCP: // fc6
		d.UnitID = r.UnitID
		d.StartAddress = r.Address
	case *WriteMultipleCoilsRequestRTU: // fc15
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.CoilCount
	case *WriteMultipleCoilsRequestTCP: // fc15
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.CoilCount
	case *WriteMultipleRegistersRequestRTU: // fc16
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.RegisterCount
	case *WriteMultipleRegistersRequestTCP: // fc16
		d.UnitID = r.UnitID
		d.StartAddress = r.StartAddress
		d.Quantity = r.RegisterCount
	case *ReadServerIDRequestRTU: // fc17
		d.UnitID = r.UnitID
	case *ReadServerIDRequestTCP: // fc17
		d.UnitID = r.UnitID
	case *ReadWriteMultipleRegistersRequestRTU: // fc23
		d.UnitID = r.UnitID
		d.StartAddress = r.ReadStartAddress
		d.Quantity = r.ReadQuantity
		// r.WriteStartAddress and r.ReadQuantity are at the moment intentionally left out
		// if these fields are needed then open an issue / send PR
	case *ReadWriteMultipleRegistersRequestTCP: // fc23
		d.UnitID = r.UnitID
		d.StartAddress = r.ReadStartAddress
		d.Quantity = r.ReadQuantity
		// r.WriteStartAddress and r.ReadQuantity are at the moment intentionally left out
		// if these fields are needed then open an issue / send PR
	default:
		return RequestDestination{}, fmt.Errorf("extract request destination: unknown function code parsed: %v", req.FunctionCode())
	}
	return d, nil
}
