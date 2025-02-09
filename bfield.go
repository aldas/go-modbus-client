package modbus

import (
	"github.com/aldas/go-modbus-client/packet"
	"time"
)

// BField is distinct field be requested and extracted from response
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
type BField struct {
	Field
	// note: this struct exists solely to provide fluent methods to set fields
}

// ServerAddress sets modbus server address for Field. Usage `[network://]host:port`
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) ServerAddress(serverAddress string) *BField {
	f.Field.ServerAddress = serverAddress
	return f
}

// FunctionCode sets FunctionCode for Field
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) FunctionCode(functionCode uint8) *BField {
	f.Field.FunctionCode = functionCode
	return f
}

// RequestInterval sets RequestInterval for Field
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) RequestInterval(requestInterval time.Duration) *BField {
	f.Field.RequestInterval = Duration(requestInterval)
	return f
}

// UnitID sets UnitID for Field
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) UnitID(unitID uint8) *BField {
	f.Field.UnitID = unitID
	return f
}

// Protocol sets Protocol for Field
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) Protocol(protocol ProtocolType) *BField {
	f.Field.Protocol = protocol
	return f
}

// ByteOrder sets word and byte order for Field to be used when extracting values from response
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) ByteOrder(byteOrder packet.ByteOrder) *BField {
	f.Field.ByteOrder = byteOrder
	return f
}

// Name sets name/identifier for Field to be used to uniquely identify value when extracting values from response
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (f *BField) Name(name string) *BField {
	f.Field.Name = name
	return f
}

// Add adds field into Builder
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
func (b *Builder) Add(field *BField) *Builder {
	b.fields = append(b.fields, field.Field)
	return b
}

// Bit add bit (0-15) field to Builder to be requested and extracted
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
//
// Deprecated: use `modbus.Field` struct and methods `Builder.AddAll(fields Fields)` and `Builder.AddField(field Field)`
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
