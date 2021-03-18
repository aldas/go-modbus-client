package modbus

import (
	"github.com/aldas/go-modbus-client/packet"
)

const (
	fieldTypeBit = iota
	fieldTypeByte
	fieldTypeUint8
	fieldTypeInt8
	fieldTypeUint16
	fieldTypeInt16
	fieldTypeUint32
	fieldTypeInt32
	fieldTypeUint64
	fieldTypeInt64
	fieldTypeFloat32
	fieldTypeFloat64
	fieldTypeString
)

// Fields is slice of Fields
type Fields []*Field

// Field is distinct field be requested and extracted from response
type Field struct {
	address string // [network://]host:port
	unitID  uint8

	registerAddress uint16
	bit             uint8
	fromHighByte    bool
	length          uint8

	fieldType uint8
	byteOrder packet.ByteOrder
	name      string
}

// ModbusAddress sets modbus address for Field. Usage `host:port`
func (f *Field) ModbusAddress(address string) *Field {
	f.address = address
	return f
}

// UnitID sets unitID for Field
func (f *Field) UnitID(unitID uint8) *Field {
	f.unitID = unitID
	return f
}

// ByteOrder sets word and byte order for Field to be used when extracting values from response
func (f *Field) ByteOrder(byteOrder packet.ByteOrder) *Field {
	f.byteOrder = byteOrder
	return f
}

// Name sets name/identifier for Field to be used to uniquely identify value when extracting values from response
func (f *Field) Name(name string) *Field {
	f.name = name
	return f
}

// registerSize returns how many register/words does this field would take in modbus response
func (f *Field) registerSize() uint16 {
	switch f.fieldType {
	case fieldTypeFloat64, fieldTypeInt64, fieldTypeUint64:
		return 4
	case fieldTypeFloat32, fieldTypeInt32, fieldTypeUint32:
		return 2
	case fieldTypeString:
		if f.length%2 == 0 { // even
			return uint16(f.length) / 2
		}
		return (uint16(f.length) / 2) + 1 // odd
	default:
		return 1
	}
}

// Builder helps to group extractable field values of different types into modbus requests with minimal amount of separate requests produced
type Builder struct {
	fields Fields

	address string // [network://]host:port
	unitID  uint8
}

// NewRequestBuilder creates new instance of Builder with given defaults
func NewRequestBuilder(address string, unitID uint8) *Builder {
	return &Builder{
		address: address,
		unitID:  unitID,
		fields:  make(Fields, 0, 5),
	}
}

// Add adds field into Builder
func (b *Builder) Add(field *Field) *Builder {
	b.fields = append(b.fields, field)
	return b
}

// Bit add bit (0-15) field to Builder to be requested and extracted
func (b *Builder) Bit(registerAddress uint16, bit uint8) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeBit,

		registerAddress: registerAddress,
		bit:             bit,
	}
}

// Byte add byte field to Builder to be requested and extracted
func (b *Builder) Byte(registerAddress uint16, fromHighByte bool) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeByte,

		registerAddress: registerAddress,
		fromHighByte:    fromHighByte,
	}
}

// Uint8 add uint8 field to Builder to be requested and extracted
func (b *Builder) Uint8(registerAddress uint16, fromHighByte bool) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeUint8,

		registerAddress: registerAddress,
		fromHighByte:    fromHighByte,
	}
}

// Int8 add int8 field to Builder to be requested and extracted
func (b *Builder) Int8(registerAddress uint16, fromHighByte bool) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeInt8,

		registerAddress: registerAddress,
		fromHighByte:    fromHighByte,
	}
}

// Uint16 add uint16 field to Builder to be requested and extracted
func (b *Builder) Uint16(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeUint16,

		registerAddress: registerAddress,
	}
}

// Int16 add int16 field to Builder to be requested and extracted
func (b *Builder) Int16(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeInt16,

		registerAddress: registerAddress,
	}
}

// Uint32 add uint32 field to Builder to be requested and extracted
func (b *Builder) Uint32(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeUint32,

		registerAddress: registerAddress,
	}
}

// Int32 add int32 field to Builder to be requested and extracted
func (b *Builder) Int32(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeInt32,

		registerAddress: registerAddress,
	}
}

// Uint64 add uint64 field to Builder to be requested and extracted
func (b *Builder) Uint64(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeUint64,

		registerAddress: registerAddress,
	}
}

// Int64 add int64 field to Builder to be requested and extracted
func (b *Builder) Int64(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeInt64,

		registerAddress: registerAddress,
	}
}

// Float32 add float32 field to Builder to be requested and extracted
func (b *Builder) Float32(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeFloat32,

		registerAddress: registerAddress,
	}
}

// Float64 add float64 field to Builder to be requested and extracted
func (b *Builder) Float64(registerAddress uint16) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeFloat64,

		registerAddress: registerAddress,
	}
}

// String add string field to Builder to be requested and extracted
func (b *Builder) String(registerAddress uint16, length uint8) *Field {
	return &Field{
		address:   b.address,
		unitID:    b.unitID,
		fieldType: fieldTypeString,
		length:    length,

		registerAddress: registerAddress,
	}
}

// RegisterRequest helps to connect requested fields to responses
type RegisterRequest struct {
	startAddress uint16
	packet.Request
	fields Fields
}

// StartAddress returns start address of contained request
func (r RegisterRequest) StartAddress() uint16 {
	return r.startAddress
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
