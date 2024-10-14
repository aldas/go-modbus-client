package modbus

import (
	"errors"
	"fmt"
	"github.com/aldas/go-modbus-client/packet"
)

const (
	// FieldTypeBit represents single bit out 16 bit register. Use `Field.Bit` (0-15) to indicate which bit is meant.
	FieldTypeBit FieldType = 1
	// FieldTypeByte represents single byte of 16 bit, 2 byte, single register. Use `Field.FromHighByte` to indicate is high or low byte is meant.
	FieldTypeByte FieldType = 2
	// FieldTypeUint8 represents uint8 value of 2 byte, single register. Use `Field.FromHighByte` to indicate is high or low byte value is meant.
	FieldTypeUint8 FieldType = 3
	// FieldTypeInt8 represents int8 value of 2 byte, single register. Use `Field.FromHighByte` to indicate is high or low byte value is meant.
	FieldTypeInt8 FieldType = 4
	// FieldTypeUint16 represents single register (16 bit) as uint16 value
	FieldTypeUint16 FieldType = 5
	// FieldTypeInt16 represents single register (16 bit) as int16 value
	FieldTypeInt16 FieldType = 6
	// FieldTypeUint32 represents 2 registers (32 bit) as uint32 value. Use `Field.ByteOrder` to indicate byte and word order of register data.
	FieldTypeUint32 FieldType = 7
	// FieldTypeInt32 represents 2 registers (32 bit) as int32 value. Use `Field.ByteOrder` to indicate byte and word order of register data.
	FieldTypeInt32 FieldType = 8
	// FieldTypeUint64 represents 4 registers (64 bit) as uint64 value. Use `Field.ByteOrder` to indicate byte and word order of register data.
	FieldTypeUint64 FieldType = 9
	// FieldTypeInt64 represents 4 registers (64 bit) as int64 value. Use `Field.ByteOrder` to indicate byte and word order of register data.
	FieldTypeInt64 FieldType = 10
	// FieldTypeFloat32 represents 2 registers (32 bit) as float32 value. Use `Field.ByteOrder` to indicate byte and word order of register data.
	FieldTypeFloat32 FieldType = 11
	// FieldTypeFloat64 represents 4 registers (64 bit) as float64 value. Use `Field.ByteOrder` to indicate byte and word order of register data.
	FieldTypeFloat64 FieldType = 12
	// FieldTypeString represents N registers as string value. Use `Field.Length` to length of string.
	FieldTypeString FieldType = 13

	// FieldTypeCoil represents single discrete/coil value (used by FC1/FC2).
	FieldTypeCoil FieldType = 14

	maxFieldTypeValue = uint8(14)
)

// FieldType is enum type for data types that Field can represent
type FieldType uint8

// Fields is slice of Field instances
type Fields []Field

// Field is distinct field be requested and extracted from response
// Tag `mapstructure` allows you to marshal https://github.com/spf13/viper supported configuration format to the Field
type Field struct {
	Name string `json:"Name" mapstructure:"Name"`

	ServerAddress string `json:"server_address" mapstructure:"server_address"` // [network://]host:port
	UnitID        uint8  `json:"unit_id" mapstructure:"unit_id"`
	// Address of the register (first register of that data type) or discrete/coil address in modbus. Addresses are 0-based.
	Address uint16    `json:"address" mapstructure:"address"`
	Type    FieldType `json:"type" mapstructure:"type"`

	// Only relevant to register function fields
	Bit          uint8            `json:"bit" mapstructure:"bit"`
	FromHighByte bool             `json:"from_high_byte" mapstructure:"from_high_byte"`
	Length       uint8            `json:"Length" mapstructure:"Length"`
	ByteOrder    packet.ByteOrder `json:"byte_order" mapstructure:"byte_order"`
}

// registerSize returns how many register/words does this field would take in modbus response
func (f *Field) registerSize() uint16 {
	switch f.Type {
	case FieldTypeFloat64, FieldTypeInt64, FieldTypeUint64:
		return 4
	case FieldTypeFloat32, FieldTypeInt32, FieldTypeUint32:
		return 2
	case FieldTypeString:
		if f.Length%2 == 0 { // even
			return uint16(f.Length) / 2
		}
		return (uint16(f.Length) / 2) + 1 // odd
	default:
		return 1
	}
}

// Validate checks if Field is values are correctly filled
func (f *Field) Validate() error {
	if f.ServerAddress == "" {
		return errors.New("field server address can not be empty")
	}
	if f.Type == 0 {
		return errors.New("field type must be set")
	}
	if uint8(f.Type) > maxFieldTypeValue {
		return errors.New("field type has invalid value")
	}
	if f.Bit > 15 {
		return errors.New("field bit value must be in range (0-15)")
	}
	if f.Type == FieldTypeString && f.Length == 0 {
		return errors.New("field with type string must have length set")
	}
	return nil
}

// ExtractFrom extracts field value from given registers data
func (f *Field) ExtractFrom(registers *packet.Registers) (interface{}, error) {
	switch f.Type {
	case FieldTypeBit:
		return registers.Bit(f.Address, f.Bit)
	case FieldTypeByte:
		return registers.Byte(f.Address, f.FromHighByte)
	case FieldTypeUint8:
		return registers.Uint8(f.Address, f.FromHighByte)
	case FieldTypeInt8:
		return registers.Int8(f.Address, f.FromHighByte)
	case FieldTypeUint16:
		return registers.Uint16(f.Address)
	case FieldTypeInt16:
		return registers.Int16(f.Address)
	case FieldTypeUint32:
		return registers.Uint32WithByteOrder(f.Address, f.ByteOrder)
	case FieldTypeInt32:
		return registers.Int32WithByteOrder(f.Address, f.ByteOrder)
	case FieldTypeUint64:
		return registers.Uint64WithByteOrder(f.Address, f.ByteOrder)
	case FieldTypeInt64:
		return registers.Int64WithByteOrder(f.Address, f.ByteOrder)
	case FieldTypeFloat32:
		return registers.Float32WithByteOrder(f.Address, f.ByteOrder)
	case FieldTypeFloat64:
		return registers.Float64WithByteOrder(f.Address, f.ByteOrder)
	case FieldTypeString:
		return registers.StringWithByteOrder(f.Address, f.Length, f.ByteOrder)
	}
	return nil, errors.New("extraction failure due unknown field type")
}

// BField is distinct field be requested and extracted from response
type BField struct {
	Field
}

// ServerAddress sets modbus server address for Field. Usage `[network://]host:port`
func (f *BField) ServerAddress(serverAddress string) *BField {
	f.Field.ServerAddress = serverAddress
	return f
}

// UnitID sets unitID for Field
func (f *BField) UnitID(unitID uint8) *BField {
	f.Field.UnitID = unitID
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

// Builder helps to group extractable field values of different types into modbus requests with minimal amount of separate requests produced
type Builder struct {
	fields Fields

	serverAddress string // [network://]host:port
	unitID        uint8
}

// NewRequestBuilder creates new instance of Builder with given defaults.
// Arguments can be left empty and ServerAddress+UnitID provided for each field separately
func NewRequestBuilder(serverAddress string, unitID uint8) *Builder {
	return &Builder{
		serverAddress: serverAddress,
		unitID:        unitID,
		fields:        make(Fields, 0, 5),
	}
}

// AddAll adds field into Builder. AddAll does not set ServerAddress and UnitID values.
func (b *Builder) AddAll(fields Fields) *Builder {
	b.fields = append(b.fields, fields...)
	return b
}

// Add adds field into Builder
func (b *Builder) Add(field *BField) *Builder {
	b.fields = append(b.fields, field.Field)
	return b
}

// Bit add bit (0-15) field to Builder to be requested and extracted
func (b *Builder) Bit(registerAddress uint16, bit uint8) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeBit,

			Address: registerAddress,
			Bit:     bit,
		},
	}
}

// Coil adds discrete/coil field to Builder to be requested and extracted by FC1/FC2.
func (b *Builder) Coil(address uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeCoil,

			Address: address,
		},
	}
}

// Byte add byte field to Builder to be requested and extracted
func (b *Builder) Byte(registerAddress uint16, fromHighByte bool) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeByte,

			Address:      registerAddress,
			FromHighByte: fromHighByte,
		},
	}
}

// Uint8 add uint8 field to Builder to be requested and extracted
func (b *Builder) Uint8(registerAddress uint16, fromHighByte bool) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeUint8,

			Address:      registerAddress,
			FromHighByte: fromHighByte,
		},
	}
}

// Int8 add int8 field to Builder to be requested and extracted
func (b *Builder) Int8(registerAddress uint16, fromHighByte bool) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeInt8,

			Address:      registerAddress,
			FromHighByte: fromHighByte,
		},
	}
}

// Uint16 add uint16 field to Builder to be requested and extracted
func (b *Builder) Uint16(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeUint16,

			Address: registerAddress,
		},
	}
}

// Int16 add int16 field to Builder to be requested and extracted
func (b *Builder) Int16(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeInt16,

			Address: registerAddress,
		},
	}
}

// Uint32 add uint32 field to Builder to be requested and extracted
func (b *Builder) Uint32(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeUint32,

			Address: registerAddress,
		},
	}
}

// Int32 add int32 field to Builder to be requested and extracted
func (b *Builder) Int32(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeInt32,

			Address: registerAddress,
		},
	}
}

// Uint64 add uint64 field to Builder to be requested and extracted
func (b *Builder) Uint64(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeUint64,

			Address: registerAddress,
		},
	}
}

// Int64 add int64 field to Builder to be requested and extracted
func (b *Builder) Int64(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeInt64,

			Address: registerAddress,
		},
	}
}

// Float32 add float32 field to Builder to be requested and extracted
func (b *Builder) Float32(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeFloat32,

			Address: registerAddress,
		},
	}
}

// Float64 add float64 field to Builder to be requested and extracted
func (b *Builder) Float64(registerAddress uint16) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeFloat64,

			Address: registerAddress,
		},
	}
}

// String add string field to Builder to be requested and extracted
func (b *Builder) String(registerAddress uint16, length uint8) *BField {
	return &BField{
		Field{
			ServerAddress: b.serverAddress,
			UnitID:        b.unitID,
			Type:          FieldTypeString,
			Length:        length,

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
	Value interface{}
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

// ReadHoldingRegistersTCP combines fields into TCP Read Holding Registers (FC3) requests
func (b *Builder) ReadHoldingRegistersTCP() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC3TCP)
}

// ReadHoldingRegistersRTU combines fields into RTU Read Holding Registers (FC3) requests
func (b *Builder) ReadHoldingRegistersRTU() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC3RTU)
}

// ReadInputRegistersTCP combines fields into TCP Read Input Registers (FC4) requests
func (b *Builder) ReadInputRegistersTCP() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC4TCP)
}

// ReadInputRegistersRTU combines fields into RTU Read Input Registers (FC4) requests
func (b *Builder) ReadInputRegistersRTU() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC4RTU)
}

// ReadCoilsTCP combines fields into TCP Read Coils (FC1) requests
func (b *Builder) ReadCoilsTCP() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC1TCP)
}

// ReadCoilsRTU combines fields into RTU Read Coils (FC1) requests
func (b *Builder) ReadCoilsRTU() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC1RTU)
}

// ReadDiscreteInputsTCP combines fields into TCP Read Discrete Inputs (FC2) requests
func (b *Builder) ReadDiscreteInputsTCP() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC2TCP)
}

// ReadDiscreteInputsRTU combines fields into RTU Read Discrete Inputs (FC2) requests
func (b *Builder) ReadDiscreteInputsRTU() ([]BuilderRequest, error) {
	return split(b.fields, splitToFC2RTU)
}
