package modbus

import (
	"errors"
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

	maxFieldTypeValue = uint8(13)
)

// FieldType is enum type for data types that Field can represent
type FieldType uint8

// Fields is slice of Field instances
type Fields []Field

// Field is distinct field be requested and extracted from response
// Tag `mapstructure` allows you to marshal https://github.com/spf13/viper supported configuration format to the Field
type Field struct {
	ServerAddress string `json:"server_address" mapstructure:"server_address"` // [network://]host:port
	UnitID        uint8  `json:"unit_id" mapstructure:"unit_id"`

	RegisterAddress uint16    `json:"register_address" mapstructure:"register_address"`
	Type            FieldType `json:"type" mapstructure:"type"`
	Bit             uint8     `json:"bit" mapstructure:"bit"`
	FromHighByte    bool      `json:"from_high_byte" mapstructure:"from_high_byte"`
	Length          uint8     `json:"Length" mapstructure:"Length"`

	ByteOrder packet.ByteOrder `json:"byte_order" mapstructure:"byte_order"`
	Name      string           `json:"Name" mapstructure:"Name"`
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
func (f Field) Validate() error {
	if f.ServerAddress == "" {
		return errors.New("field server address can not be empty")
	}
	if f.RegisterAddress == 0 {
		return errors.New("field register address can not be 0")
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
	for _, f := range fields {
		b.fields = append(b.fields, f)
	}
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

			RegisterAddress: registerAddress,
			Bit:             bit,
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

			RegisterAddress: registerAddress,
			FromHighByte:    fromHighByte,
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

			RegisterAddress: registerAddress,
			FromHighByte:    fromHighByte,
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

			RegisterAddress: registerAddress,
			FromHighByte:    fromHighByte,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
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

			RegisterAddress: registerAddress,
		},
	}
}

// RegisterRequest helps to connect requested fields to responses
type RegisterRequest struct {
	packet.Request

	serverAddress string
	unitID        uint8
	startAddress  uint16

	fields Fields
}

// ServerAddress returns modbus server address of contained request
func (r RegisterRequest) ServerAddress() string {
	return r.serverAddress
}

// UnitID returns UnitID of contained request
func (r RegisterRequest) UnitID() uint8 {
	return r.unitID
}

// StartAddress returns start Address of contained request
func (r RegisterRequest) StartAddress() uint16 {
	return r.startAddress
}

// RegistersResponse is marker interface for responses returning register data
type RegistersResponse interface {
	packet.Response
	AsRegisters(requestStartAddress uint16) (*packet.Registers, error)
}

// AsRegisters returns response data as Register to more convenient access
func (r RegisterRequest) AsRegisters(response RegistersResponse) (*packet.Registers, error) {
	return response.AsRegisters(r.startAddress)
}

// ReadHoldingRegistersTCP combines fields into TCP Read Holding Registers (FC3) requests
func (b *Builder) ReadHoldingRegistersTCP() ([]RegisterRequest, error) {
	return split(b.fields, "fc3_tcp")
}

// ReadHoldingRegistersRTU combines fields into RTU Read Holding Registers (FC3) requests
func (b *Builder) ReadHoldingRegistersRTU() ([]RegisterRequest, error) {
	return split(b.fields, "fc3_rtu")
}

// ReadInputRegistersTCP combines fields into TCP Read Input Registers (FC4) requests
func (b *Builder) ReadInputRegistersTCP() ([]RegisterRequest, error) {
	return split(b.fields, "fc4_tcp")
}

// ReadInputRegistersRTU combines fields into RTU Read Input Registers (FC4) requests
func (b *Builder) ReadInputRegistersRTU() ([]RegisterRequest, error) {
	return split(b.fields, "fc4_rtu")
}
