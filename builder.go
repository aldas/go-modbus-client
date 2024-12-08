package modbus

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aldas/go-modbus-client/packet"
	"strconv"
	"strings"
	"time"
)

const (
	protocolAny ProtocolType = 0 // any or unknown protocol

	// ProtocolTCP represents Modbus TCP encoding which could be transferred over TCP/UDP connection.
	ProtocolTCP ProtocolType = 1
	// ProtocolRTU represents Modbus RTU encoding with a Cyclic-Redundant Checksum. Could be transferred
	// over TCP/UDP and serial connection.
	ProtocolRTU ProtocolType = 2
	// protocolASCII represents Modbus ASCII encoding where each data byte is split into the two bytes
	// representing the two ASCII characters in the Hexadecimal value
	// NOTE: NOT YET IMPLEMENTED
	// protocolASCII ProtocolType = 3
)

// ProtocolType represents which Modbus encoding is being used.
type ProtocolType uint8

// UnmarshalJSON converts raw bytes from JSON to Invalid
func (pt *ProtocolType) UnmarshalJSON(raw []byte) error {
	t := string(raw)
	switch strings.ToLower(t) {
	case `"tcp"`:
		*pt = ProtocolTCP
	case `"rtu"`:
		*pt = ProtocolRTU

	default:
		return fmt.Errorf("unknown protocol value, given: '%s'", t)
	}
	return nil
}

// BField is distinct field be requested and extracted from response
type BField struct {
	Field
	// note: this struct exists solely to provide fluent methods to set fields
}

// ServerAddress sets modbus server address for Field. Usage `[network://]host:port`
func (f *BField) ServerAddress(serverAddress string) *BField {
	f.Field.ServerAddress = serverAddress
	return f
}

// FunctionCode sets FunctionCode for Field
func (f *BField) FunctionCode(functionCode uint8) *BField {
	f.Field.FunctionCode = functionCode
	return f
}

// RequestInterval sets RequestInterval for Field
func (f *BField) RequestInterval(requestInterval time.Duration) *BField {
	f.Field.RequestInterval = Duration(requestInterval)
	return f
}

// UnitID sets UnitID for Field
func (f *BField) UnitID(unitID uint8) *BField {
	f.Field.UnitID = unitID
	return f
}

// Protocol sets Protocol for Field
func (f *BField) Protocol(protocol ProtocolType) *BField {
	f.Field.Protocol = protocol
	return f
}

// ByteOrder sets word and byte order for Field to be used when extracting values from response
func (f *BField) ByteOrder(byteOrder packet.ByteOrder) *BField {
	f.Field.ByteOrder = byteOrder
	return f
}

// Name sets name/identifier for Field to be used to uniquely identify value when extracting values from response
func (f *BField) Name(name string) *BField {
	f.Field.Name = name
	return f
}

// A Duration represents the elapsed time between two instants. This library type extends time.Duration
// with JSON marshalling and unmarshalling support to/from string. i.e. "1s" unmarshalled to `1 * time.Second`
type Duration time.Duration

// MarshalJSON converts Duration to JSON bytes
func (d Duration) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.WriteRune('"')
	buf.WriteString(time.Duration(d).String())
	buf.WriteRune('"')
	return buf.Bytes(), nil
}

// UnmarshalJSON converts raw bytes from JSON to Duration
func (d *Duration) UnmarshalJSON(raw []byte) error {
	if raw[0] != '"' {
		v, err := strconv.ParseInt(string(raw), 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse Duration as int, err: %w", err)
		}
		*d = Duration(v)
		return nil
	}

	if len(raw) < 3 {
		return fmt.Errorf("duration value too short, given: '%s'", raw)
	}
	e := len(raw) - 1
	if raw[e] != '"' {
		return fmt.Errorf("duration value does not end with quote mark, given: '%s'", raw)
	}
	tmp, err := time.ParseDuration(string(raw[1:e]))
	if err != nil {
		return fmt.Errorf("could not parse Duration from string, err: %w", err)
	}
	*d = Duration(tmp)
	return nil
}

// Builder helps to group extractable field values of different types into modbus requests with minimal amount of separate requests produced
type Builder struct {
	fields Fields
	config BuilderDefaults
}

// BuilderDefaults holds Builder default values for adding/creating Fields
type BuilderDefaults struct {
	// Should be formatted as url.URL scheme `[scheme:][//[userinfo@]host][/]path[?query]`
	// Example:
	// * `127.0.0.1:502` (library defaults to `tcp` as scheme)
	// * `udp://127.0.0.1:502`
	// * `/dev/ttyS0?BaudRate=4800`
	// * `file:///dev/ttyUSB?BaudRate=4800`
	ServerAddress string       `json:"server_address" mapstructure:"server_address"`
	FunctionCode  uint8        `json:"function_code" mapstructure:"function_code"`
	UnitID        uint8        `json:"unit_id" mapstructure:"unit_id"`
	Protocol      ProtocolType `json:"protocol" mapstructure:"protocol"`
	Interval      Duration     `json:"interval" mapstructure:"interval"`
}

// NewRequestBuilderWithConfig creates new instance of Builder with given defaults.
// Arguments can be left empty and ServerAddress+UnitID provided for each field separately
func NewRequestBuilderWithConfig(config BuilderDefaults) *Builder {
	return &Builder{
		fields: make(Fields, 0, 5),
		config: config,
	}
}

// NewRequestBuilder creates new instance of Builder
func NewRequestBuilder(serverAddress string, unitID uint8) *Builder {
	return NewRequestBuilderWithConfig(BuilderDefaults{
		ServerAddress: serverAddress,
		UnitID:        unitID,
	})
}

// AddAll adds field into Builder. AddAll does not set ServerAddress and UnitID values.
func (b *Builder) AddAll(fields Fields) *Builder {
	for _, field := range fields {
		b.add(field)
	}
	return b
}

// AddField adds field into Builder
func (b *Builder) AddField(field Field) *Builder {
	return b.add(field)
}

// Add adds field into Builder
func (b *Builder) Add(field *BField) *Builder {
	b.fields = append(b.fields, field.Field)
	return b
}

// Add adds field into Builder
func (b *Builder) add(field Field) *Builder {
	if field.ServerAddress == "" {
		field.ServerAddress = b.config.ServerAddress
	}
	if field.FunctionCode == 0 {
		field.FunctionCode = b.config.FunctionCode
	}
	if field.UnitID == 0 && b.config.UnitID != 0 {
		field.UnitID = b.config.UnitID
	}
	if field.Protocol == protocolAny {
		field.Protocol = b.config.Protocol
	}
	if field.RequestInterval == 0 {
		field.RequestInterval = b.config.Interval
	}
	b.fields = append(b.fields, field)
	return b
}

// Bit add bit (0-15) field to Builder to be requested and extracted
func (b *Builder) Bit(registerAddress uint16, bit uint8) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeBit,

			Address: registerAddress,
			Bit:     bit,
		},
	}
}

// Coil adds discrete/coil field to Builder to be requested and extracted by FC1/FC2.
func (b *Builder) Coil(address uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeCoil,

			Address: address,
		},
	}
}

// Byte add byte field to Builder to be requested and extracted
func (b *Builder) Byte(registerAddress uint16, fromHighByte bool) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeByte,

			Address:      registerAddress,
			FromHighByte: fromHighByte,
		},
	}
}

// Uint8 add uint8 field to Builder to be requested and extracted
func (b *Builder) Uint8(registerAddress uint16, fromHighByte bool) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeUint8,

			Address:      registerAddress,
			FromHighByte: fromHighByte,
		},
	}
}

// Int8 add int8 field to Builder to be requested and extracted
func (b *Builder) Int8(registerAddress uint16, fromHighByte bool) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeInt8,

			Address:      registerAddress,
			FromHighByte: fromHighByte,
		},
	}
}

// Uint16 add uint16 field to Builder to be requested and extracted
func (b *Builder) Uint16(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeUint16,

			Address: registerAddress,
		},
	}
}

// Int16 add int16 field to Builder to be requested and extracted
func (b *Builder) Int16(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeInt16,

			Address: registerAddress,
		},
	}
}

// Uint32 add uint32 field to Builder to be requested and extracted
func (b *Builder) Uint32(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeUint32,

			Address: registerAddress,
		},
	}
}

// Int32 add int32 field to Builder to be requested and extracted
func (b *Builder) Int32(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeInt32,

			Address: registerAddress,
		},
	}
}

// Uint64 add uint64 field to Builder to be requested and extracted
func (b *Builder) Uint64(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeUint64,

			Address: registerAddress,
		},
	}
}

// Int64 add int64 field to Builder to be requested and extracted
func (b *Builder) Int64(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeInt64,

			Address: registerAddress,
		},
	}
}

// Float32 add float32 field to Builder to be requested and extracted
func (b *Builder) Float32(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeFloat32,

			Address: registerAddress,
		},
	}
}

// Float64 add float64 field to Builder to be requested and extracted
func (b *Builder) Float64(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type: FieldTypeFloat64,

			Address: registerAddress,
		},
	}
}

// String add string field to Builder to be requested and extracted
func (b *Builder) String(registerAddress uint16, length uint8) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type:   FieldTypeString,
			Length: length,

			Address: registerAddress,
		},
	}
}

// Bytes add raw bytes field to Builder to be requested and extracted. byteLength is length in bytes (1 register is 2 bytes)
func (b *Builder) Bytes(registerAddress uint16, byteLength uint8) *BField {
	return &BField{
		Field{
			ServerAddress:   b.config.ServerAddress,
			FunctionCode:    b.config.FunctionCode,
			UnitID:          b.config.UnitID,
			Protocol:        b.config.Protocol,
			RequestInterval: b.config.Interval,

			Type:   FieldTypeRawBytes,
			Length: byteLength,

			Address: registerAddress,
		},
	}
}

// BuilderRequest helps to connect requested fields to responses
type BuilderRequest struct {
	packet.Request

	// ServerAddress is modbus server address where request should be sent
	ServerAddress string
	// UnitID is unit identifier of modbus slave device
	UnitID uint8
	// StartAddress is start register address for request
	StartAddress uint16

	Protocol        ProtocolType
	RequestInterval time.Duration

	// Fields is slice of field use to construct the request and to be extracted from response
	Fields Fields
}

// RegistersResponse is marker interface for responses returning register data
type RegistersResponse interface {
	packet.Response
	AsRegisters(requestStartAddress uint16) (*packet.Registers, error)
}

// CoilsResponse is marker interface for responses returning coil/discrete data
type CoilsResponse interface {
	packet.Response
	IsCoilSet(startAddress uint16, coilAddress uint16) (bool, error)
}

// AsRegisters returns response data as Register to more convenient access
func (r BuilderRequest) AsRegisters(response RegistersResponse) (*packet.Registers, error) {
	return response.AsRegisters(r.StartAddress)
}

// FieldValue is concrete value extracted from register data using field data type and byte order
type FieldValue struct {
	Field Field

	// Value contains extracted value
	// possible types:
	// * bool
	// * byte
	// * []byte
	// * uint[8/16/32/64]
	// * int[8/16/32/64]
	// * float[32/64]
	// * string
	Value any

	// Error contains error that occurred during extracting field from response.
	// In case Field.Invalid was set and response data contained it the Error is set to modbus.ErrInvalidValue
	Error error
}

// ErrorFieldExtractHadError is returned when ExtractFields could not extract value from Field
var ErrorFieldExtractHadError = errors.New("field extraction had an error. check FieldValue.Error for details")

// ExtractFields extracts Field values from given response. When continueOnExtractionErrors is true and error occurs
// during extraction, this method does not end but continues to extract all Fields and returns ErrorFieldExtractHadError
// at the end. To distinguish errors check FieldValue.Error field.
func (r BuilderRequest) ExtractFields(response packet.Response, continueOnExtractionErrors bool) ([]FieldValue, error) {
	switch resp := response.(type) {
	case RegistersResponse:
		return r.extractRegisterFields(resp, continueOnExtractionErrors)
	case CoilsResponse:
		return r.extractCoilFields(resp, continueOnExtractionErrors)
	}
	return nil, errors.New("can not extract fields from unsupported response type")
}

func (r BuilderRequest) extractRegisterFields(response RegistersResponse, continueOnExtractionErrors bool) ([]FieldValue, error) {
	regs, err := response.AsRegisters(r.StartAddress)
	if err != nil {
		return nil, err
	}

	hadErrors := false
	capacity := 0
	if continueOnExtractionErrors {
		capacity = len(r.Fields)
	}
	result := make([]FieldValue, 0, capacity)
	for _, f := range r.Fields {
		vTmp, err := f.ExtractFrom(regs)
		if err != nil && !continueOnExtractionErrors {
			return nil, fmt.Errorf("field extraction failed. name: %v err: %w", f.Name, err)
		}
		if !hadErrors && err != nil {
			hadErrors = true
		}
		tmp := FieldValue{
			Field: f,
			Value: vTmp,
			Error: err,
		}
		result = append(result, tmp)
	}
	if hadErrors {
		return result, ErrorFieldExtractHadError
	}
	return result, nil
}

func (r BuilderRequest) extractCoilFields(response CoilsResponse, continueOnExtractionErrors bool) ([]FieldValue, error) {
	hadErrors := false
	capacity := 0
	if continueOnExtractionErrors {
		capacity = len(r.Fields)
	}
	result := make([]FieldValue, 0, capacity)
	for _, f := range r.Fields {
		vTmp, err := response.IsCoilSet(r.StartAddress, f.Address)

		if err != nil && !continueOnExtractionErrors {
			return nil, fmt.Errorf("field extraction failed. name: %v err: %w", f.Name, err)
		}
		if !hadErrors && err != nil {
			hadErrors = true
		}
		tmp := FieldValue{
			Field: f,
			Value: vTmp,
			Error: err,
		}
		result = append(result, tmp)
	}
	if hadErrors {
		return result, ErrorFieldExtractHadError
	}
	return result, nil
}

// Split combines fields into requests by their ServerAddress+FunctionCode+UnitID+Protocol+RequestInterval
func (b *Builder) Split() ([]BuilderRequest, error) {
	return split(b.fields, 0, protocolAny)
}

// ReadHoldingRegistersTCP combines fields into TCP Read Holding Registers (FC3) requests
func (b *Builder) ReadHoldingRegistersTCP() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadHoldingRegisters, ProtocolTCP)
}

// ReadHoldingRegistersRTU combines fields into RTU Read Holding Registers (FC3) requests
func (b *Builder) ReadHoldingRegistersRTU() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadHoldingRegisters, ProtocolRTU)
}

// ReadInputRegistersTCP combines fields into TCP Read Input Registers (FC4) requests
func (b *Builder) ReadInputRegistersTCP() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadInputRegisters, ProtocolTCP)
}

// ReadInputRegistersRTU combines fields into RTU Read Input Registers (FC4) requests
func (b *Builder) ReadInputRegistersRTU() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadInputRegisters, ProtocolRTU)
}

// ReadCoilsTCP combines fields into TCP Read Coils (FC1) requests
func (b *Builder) ReadCoilsTCP() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadCoils, ProtocolTCP)
}

// ReadCoilsRTU combines fields into RTU Read Coils (FC1) requests
func (b *Builder) ReadCoilsRTU() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadCoils, ProtocolRTU)
}

// ReadDiscreteInputsTCP combines fields into TCP Read Discrete Inputs (FC2) requests
func (b *Builder) ReadDiscreteInputsTCP() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadDiscreteInputs, ProtocolTCP)
}

// ReadDiscreteInputsRTU combines fields into RTU Read Discrete Inputs (FC2) requests
func (b *Builder) ReadDiscreteInputsRTU() ([]BuilderRequest, error) {
	return split(b.fields, packet.FunctionReadDiscreteInputs, ProtocolRTU)
}
