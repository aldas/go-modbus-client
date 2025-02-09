package modbus

import (
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBField_ServerAddress(t *testing.T) {
	given := &BField{}

	given.ServerAddress(":502")

	assert.Equal(t, ":502", given.Field.ServerAddress)
}

func TestBField_FunctionCode(t *testing.T) {
	given := &BField{}

	given.FunctionCode(0x2)

	assert.Equal(t, uint8(0x2), given.Field.FunctionCode)
}

func TestBField_Protocol(t *testing.T) {
	given := &BField{}

	given.Protocol(ProtocolTCP)

	assert.Equal(t, ProtocolTCP, given.Field.Protocol)
}

func TestBField_RequestInterval(t *testing.T) {
	given := &BField{}

	given.RequestInterval(1 * time.Second)

	assert.Equal(t, Duration(1*time.Second), given.Field.RequestInterval)
}

func TestBField_UnitID(t *testing.T) {
	given := &BField{}

	given.UnitID(1)

	assert.Equal(t, uint8(1), given.Field.UnitID)
}

func TestBField_ByteOrder(t *testing.T) {
	given := &BField{}

	given.ByteOrder(packet.BigEndian)

	assert.Equal(t, packet.BigEndian, given.Field.ByteOrder)
}

func TestBField_Name(t *testing.T) {
	given := &BField{}

	given.Name("fire_alarm_do")

	assert.Equal(t, "fire_alarm_do", given.Field.Name)
}

func TestBuilder_Add(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)
	b.Add(&BField{Field{ServerAddress: "test", UnitID: 1}})

	assert.Equal(t, "test", b.fields[0].ServerAddress)
	assert.Equal(t, uint8(1), b.fields[0].UnitID)
}

func TestBuilder_Bit(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Bit(256, 4).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeBit,
		Address:       256,
		Bit:           4,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Byte(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Byte(256, true).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeByte,
		Address:       256,
		FromHighByte:  true,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint8(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint8(256, true).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint8,
		Address:       256,
		FromHighByte:  true,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int8(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int8(256, true).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt8,
		Address:       256,
		FromHighByte:  true,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint16(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint16(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint16,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int16(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int16(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt16,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint32(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint32(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint32,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int32(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int32(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt32,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Uint64(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Uint64(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeUint64,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Int64(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Int64(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeInt64,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Float32(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Float32(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeFloat32,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Float64(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Float64(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeFloat64,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_String(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.String(256, 10).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeString,
		Address:       256,
		Length:        10,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Bytes(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Bytes(256, 10).Name("raw_bytes"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeRawBytes,
		Address:       256,
		Length:        10,
		Name:          "raw_bytes",
	}
	assert.Equal(t, expect, b.fields[0])
}

func TestBuilder_Coil(t *testing.T) {
	b := NewRequestBuilder(":5020", 2)

	b.Add(b.Coil(256).Name("fire_alarm_di"))

	expect := Field{
		ServerAddress: ":5020",
		UnitID:        2,
		Type:          FieldTypeCoil,
		Address:       256,
		Name:          "fire_alarm_di",
	}
	assert.Equal(t, expect, b.fields[0])
}
