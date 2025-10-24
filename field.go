package modbus

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	// FieldTypeRawBytes represents N registers contents as byte slice.
	FieldTypeRawBytes FieldType = 15

	maxFieldTypeValue = uint8(15)
)

// ErrInvalidValue is returned when extracted value for Field resulted invalid value (Field.Invalid).
var ErrInvalidValue = errors.New("invalid value")

// FieldType is enum type for data types that Field can represent
type FieldType uint8

// UnmarshalJSON converts raw bytes from JSON to FieldType
func (ft *FieldType) UnmarshalJSON(raw []byte) error {
	if len(raw) < 3 {
		return fmt.Errorf("field type value too short, given: '%s'", raw)
	}
	if raw[0] != '"' {
		return fmt.Errorf("field type value does not start with quote mark, given: '%s'", raw)
	}
	e := len(raw) - 1
	if raw[e] != '"' {
		return fmt.Errorf("field type value does not end with quote mark, given: '%s'", raw)
	}

	tmp, err := ParseFieldType(string(raw[1:e]))
	if err != nil {
		return err
	}
	*ft = tmp
	return nil
}

// ParseFieldType parses given string to FieldType
func ParseFieldType(raw string) (FieldType, error) {
	var ft FieldType = 0
	switch strings.ToLower(raw) {
	case `bit`:
		ft = FieldTypeBit
	case `byte`:
		ft = FieldTypeByte
	case `uint8`:
		ft = FieldTypeUint8
	case `int8`:
		ft = FieldTypeInt8
	case `uint16`:
		ft = FieldTypeUint16
	case `int16`:
		ft = FieldTypeInt16
	case `uint32`:
		ft = FieldTypeUint32
	case `int32`:
		ft = FieldTypeInt32
	case `uint64`:
		ft = FieldTypeUint64
	case `int64`:
		ft = FieldTypeInt64
	case `float32`:
		ft = FieldTypeFloat32
	case `float64`:
		ft = FieldTypeFloat64
	case `string`:
		ft = FieldTypeString
	case `bytes`:
		ft = FieldTypeRawBytes
	case `coil`:
		ft = FieldTypeCoil
	default:
		return ft, fmt.Errorf("unknown field type value, given: '%s'", raw)
	}
	return ft, nil
}

// Fields is slice of Field instances
type Fields []Field

// Field is distinct field be requested and extracted from response
// Tag `mapstructure` allows you to marshal https://github.com/spf13/viper supported configuration format to the Field
type Field struct {
	Name string `json:"name" mapstructure:"name"`

	// ServerAddress is Modbus server location as URL.
	// URL: `scheme://host:port` or file `/dev/ttyS0?BaudRate=4800`
	// Query parameters are used to pass information about network connection or how to split
	// fields into request batches.
	//
	// Splitter logic know following query parameters:
	// - `max_quantity_per_request` maximum quantity (uint16) that request can have. How many
	//		registers/coils requests will be limited to.
	// - `invalid_addr=70,100-120` - addresses or address range that splitter will avoid to include in request
	//		when creating batches.
	ServerAddress   string       `json:"server_address" mapstructure:"server_address"`
	FunctionCode    uint8        `json:"function_code" mapstructure:"function_code"`
	UnitID          uint8        `json:"unit_id" mapstructure:"unit_id"`
	Protocol        ProtocolType `json:"protocol" mapstructure:"protocol"`
	RequestInterval Duration     `json:"request_interval" mapstructure:"request_interval"`

	// Address of the register (first register of that data type) or discrete/coil address in modbus.
	// Addresses are 0-based.
	Address uint16    `json:"address" mapstructure:"address"`
	Type    FieldType `json:"type" mapstructure:"type"`

	// Only relevant to register function fields
	Bit uint8 `json:"bit" mapstructure:"bit"`

	// FromHighByte is for single byte data types stored in single register (e.g. byte,uint8,int8)
	//
	// In Modbus (which uses big-endian format), the most significant byte is
	// sent first and is therefore considered the 0th byte. The least significant byte
	// is sent second and is considered the 1st byte.
	//
	// Modbus register with value `0x1234`.
	//  - 0x12 is High Byte, 0th byte
	//  - 0x34 is Low byte, is the 1st byte
	FromHighByte bool `json:"from_high_byte" mapstructure:"from_high_byte"`

	// Length is length of string and raw bytes data types.
	Length uint8 `json:"length" mapstructure:"length"`

	ByteOrder packet.ByteOrder `json:"byte_order" mapstructure:"byte_order"`

	// Invalid that represents not existent value in modbus. Given value (presented in hex) when encountered is converted to ErrInvalidValue error.
	// for example your energy meter ac power is uint32 value of which `0xffffffff` should be treated as error/invalid value.
	//
	// Usually invalid value is largest unsigned or smallest signed value per data type. Example:
	// - uint8 	= 0xff (255)
	// - int8 	= 0x80 (-127)
	// - uint16 = 0xff, 0xff (65535)
	// - int16	= 0x80, 0x00 (-32768)
	// - uint32 = 0xff, 0xff, 0xff, 0xff (4294967295)
	// - int32	= 0x80, 0x0, 0x0, 0x0 (-2147483648)
	// - uint64 = 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff (18446744073709551615)
	// - int64	= 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0 (-9223372036854775808)
	// - float32 is same as uint32
	// - float64 is same as uint64
	// - bit, boolean - can not have invalid values
	Invalid Invalid `json:"invalid,omitempty" mapstructure:"invalid"`
}

// registerSize returns how many register/words does this field would take in modbus response
func (f *Field) registerSize() uint16 {
	switch f.Type {
	case FieldTypeFloat64, FieldTypeInt64, FieldTypeUint64:
		return 4
	case FieldTypeFloat32, FieldTypeInt32, FieldTypeUint32:
		return 2
	case FieldTypeString, FieldTypeRawBytes:
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
	switch f.Type {
	case FieldTypeCoil:
		fc := f.FunctionCode
		if !(fc == 0 || fc == packet.FunctionReadCoils || fc == packet.FunctionReadDiscreteInputs) {
			return errors.New("field with type coil must have function code of 0,1,2")
		}
	case FieldTypeString:
		if f.Length == 0 {
			return errors.New("field with type string must have length set")
		}
	case FieldTypeRawBytes:
		if f.Length == 0 {
			return errors.New("field with type bytes must have length set")
		}
	}

	switch f.Protocol {
	case ProtocolTCP, ProtocolRTU, protocolAny:
	default:
		return errors.New("field has invalid protocol type")
	}
	return nil
}

// ExtractFrom extracts field value from given registers data
func (f *Field) ExtractFrom(registers *packet.Registers) (interface{}, error) {
	if err := f.CheckInvalid(registers); err != nil {
		return nil, err
	}

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
	case FieldTypeRawBytes:
		return registers.BytesWithByteOrder(f.Address, f.Length, f.ByteOrder)
	}
	return nil, errors.New("extraction failure due unknown field type")
}

// MarshalBytes converts given value to suitable Modbus bytes (in big endian) that can be used as request data.
// This method will truncate/limit value to max/min value of field being marshalled.
// In case field type is Float32 or Float64 providing bigger int (e.g. uint64) conversion will have loss of precision
//
// Accepted types for given value are:
// - bool
// - uint8,int8,byte
// - uint16,int16
// - uint32,int32
// - uint64,int64
// - float32,float64
// - string (Note: raw utf8 bytes. If you need ASCII, convert the string before)
// - []byte
func (f *Field) MarshalBytes(value any) ([]byte, error) {
	registerSize := f.registerSize()
	sizeBytes := int(registerSize) * 2

	dst := make([]byte, sizeBytes)
	switch f.Type {
	case FieldTypeBit: // bit can originate only from single register
		if err := marshalFieldTypeBit(dst, value, f.Bit); err != nil {
			return nil, err
		}
	case FieldTypeString, FieldTypeRawBytes:
		if err := marshalFieldTypeStringOrBytes(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeByte, FieldTypeUint8: // byte is alias to uint8
		if err := marshalFieldTypeUint8(dst, value, f.FromHighByte); err != nil {
			return nil, err
		}
	case FieldTypeInt8:
		if err := marshalFieldTypeInt8(dst, value, f.FromHighByte); err != nil {
			return nil, err
		}
	case FieldTypeUint16:
		if err := marshalFieldTypeUint16(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeInt16:
		if err := marshalFieldTypeInt16(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeUint32:
		if err := marshalFieldTypeUint32(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeInt32:
		if err := marshalFieldTypeInt32(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeUint64:
		if err := marshalFieldTypeUint64(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeInt64:
		if err := marshalFieldTypeInt64(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeFloat32:
		if err := marshalFieldTypeFloat32(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeFloat64:
		if err := marshalFieldTypeFloat64(dst, value); err != nil {
			return nil, err
		}
	case FieldTypeCoil:
		return nil, errors.New("coil field type is unsupported for MarshalBytes")
	default:
		return nil, errors.New("unsupported field type for MarshalBytes")
	}

	if registerSize > 1 && f.ByteOrder&packet.LowWordFirst != 0 {
		if err := registersToLowWordFirst(dst); err != nil {
			return nil, err
		}
	}
	return dst, nil
}

// CheckInvalid compares Invalid value to bytes in fields registers. When raw data in response
// equal to Invalid the ErrInvalidValue error is returned. Nil return value means no problems occurred.
func (f *Field) CheckInvalid(registers *packet.Registers) error {
	if len(f.Invalid) == 0 {
		return nil
	}

	if f.Type == FieldTypeByte || f.Type == FieldTypeUint8 || f.Type == FieldTypeInt8 {
		regData, err := registers.Register(f.Address)
		if err != nil {
			return err
		}
		b := regData[1]
		if f.FromHighByte {
			b = regData[0]
		}
		if b == f.Invalid[0] {
			return ErrInvalidValue
		}
		return nil
	}

	ok, err := registers.IsEqualBytes(f.Address, uint8(f.registerSize()*2), f.Invalid)
	if err != nil {
		return err
	}
	if ok {
		return ErrInvalidValue
	}
	return nil
}

// Invalid that represents not existent value in modbus. Given value (presented in hex) when encountered is converted to ErrInvalidValue error.
// for example your energy meter ac power is uint32 value of which `0xffffffff` should be treated as error/invalid value.
type Invalid []byte

// MarshalJSON converts Invalid to JSON bytes
func (i Invalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(i))
}

// UnmarshalJSON converts raw bytes from JSON to Invalid
func (i *Invalid) UnmarshalJSON(b []byte) error {
	l := len(b)
	if l < 3 { // minimum is `"0"
		return errors.New("could not unmarshal Invalid, raw value too short")
	}
	if b[0] != '"' || b[l-1] != '"' {
		return errors.New("could not unmarshal Invalid, raw value does not seems to be string")
	}

	b, err := hex.DecodeString(string(b[1 : l-1]))
	if err != nil {
		return fmt.Errorf("could not unmarshal Invalid hex string, err: %w", err)
	}
	*i = b
	return nil
}
