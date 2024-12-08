package modbus

import (
	"encoding/json"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestField_registerSize(t *testing.T) {
	var testCases = []struct {
		name   string
		when   Field
		expect uint16
	}{
		{
			name:   "bit",
			when:   Field{Type: FieldTypeBit, Bit: 1},
			expect: 1,
		},
		{
			name:   "byte",
			when:   Field{Type: FieldTypeByte, FromHighByte: true},
			expect: 1,
		},
		{
			name:   "uint8",
			when:   Field{Type: FieldTypeUint8, FromHighByte: false},
			expect: 1,
		},
		{
			name:   "int8",
			when:   Field{Type: FieldTypeInt8, FromHighByte: true},
			expect: 1,
		},
		{
			name:   "uint16",
			when:   Field{Type: FieldTypeUint16},
			expect: 1,
		},
		{
			name:   "int16",
			when:   Field{Type: FieldTypeInt16},
			expect: 1,
		},
		{
			name:   "uint32",
			when:   Field{Type: FieldTypeUint32},
			expect: 2,
		},
		{
			name:   "int32",
			when:   Field{Type: FieldTypeInt32},
			expect: 2,
		},
		{
			name:   "uint64",
			when:   Field{Type: FieldTypeUint64},
			expect: 4,
		},
		{
			name:   "int64",
			when:   Field{Type: FieldTypeInt64},
			expect: 4,
		},
		{
			name:   "float32",
			when:   Field{Type: FieldTypeFloat32},
			expect: 2,
		},
		{
			name:   "float64",
			when:   Field{Type: FieldTypeFloat64},
			expect: 4,
		},
		{
			name:   "string odd size",
			when:   Field{Type: FieldTypeString, Length: 5},
			expect: 3,
		},
		{
			name:   "string even size",
			when:   Field{Type: FieldTypeString, Length: 6},
			expect: 3,
		},
		{
			name:   "string even size2",
			when:   Field{Type: FieldTypeString, Length: 4},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.when.registerSize())
		})
	}
}

func TestField_ExtractFrom(t *testing.T) {
	var testCases = []struct {
		name              string
		givenRegisterData []byte
		whenType          FieldType
		whenByteOrder     packet.ByteOrder
		whenInvalid       []byte
		expect            interface{}
		expectErr         string
	}{
		{
			name:              "bit",
			givenRegisterData: []byte{0x0, 0x0, 0b00010001, 0x0},
			whenType:          FieldTypeBit,
			expect:            true,
		},
		{
			name:              "byte",
			givenRegisterData: []byte{0x0, 0x0, 0x1, 0x0},
			whenType:          FieldTypeByte,
			expect:            byte(1),
		},
		{
			name:              "uint8",
			whenType:          FieldTypeUint8,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0x0},
			expect:            uint8(255),
		},
		{
			name:              "int8",
			whenType:          FieldTypeInt8,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0x0},
			expect:            int8(-1),
		},
		{
			name:              "uint16",
			whenType:          FieldTypeUint16,
			givenRegisterData: []byte{0x0, 0x0, 0x0, 0xFF},
			expect:            uint16(255),
		},
		{
			name:              "int16",
			whenType:          FieldTypeInt16,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0xFF},
			expect:            int16(-1),
		},
		{
			name:              "uint32",
			whenType:          FieldTypeUint32,
			givenRegisterData: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
			expect:            uint32(1),
		},
		{
			name:              "int32",
			whenType:          FieldTypeInt32,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0xFF, 0xFF, 0xFF},
			expect:            int32(-1),
		},
		{
			name:              "uint64",
			whenType:          FieldTypeUint64,
			givenRegisterData: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
			expect:            uint64(1),
		},
		{
			name:              "int64",
			whenType:          FieldTypeInt64,
			givenRegisterData: []byte{0x0, 0x0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expect:            int64(-1),
		},
		{
			name:              "float32",
			whenType:          FieldTypeFloat32,
			whenByteOrder:     packet.BigEndianLowWordFirst,
			givenRegisterData: []byte{0x0, 0x0, 0xcc, 0xcd, 0x3f, 0xec},
			expect:            float32(1.85),
		},
		{
			name:              "float64",
			whenType:          FieldTypeFloat64,
			whenByteOrder:     packet.BigEndianLowWordFirst,
			givenRegisterData: []byte{0x0, 0x0, 0x99, 0x9a, 0x99, 0x99, 0x99, 0x99, 0x3f, 0xfd},
			expect:            float64(1.85),
		},
		{
			name:          "string odd size",
			whenType:      FieldTypeString,
			whenByteOrder: packet.LittleEndian,
			givenRegisterData: []byte{
				0x0, 0x0, // register 0
				0x41, 0x42, // register 1 [A,B] -> LE -> [B,A]
				0x43, 0x44, // register 2 [C,D] -> LE -> [D,C]
			},
			expect: "BAD",
		},
		{
			name:          "raw bytes odd size",
			whenType:      FieldTypeRawBytes,
			whenByteOrder: packet.LittleEndian,
			givenRegisterData: []byte{
				0x0, 0x0, // register 0
				0x41, 0x42, // register 1 [A,B] -> LE -> [B,A]
				0x43, 0x44, // register 2 [C,D] -> LE -> [D,C]
			},
			expect: []byte{0x42, 0x41, 0x44}, // BAD
		},
		{
			name:              "nok, coil can not be extracted from registers",
			whenType:          FieldTypeCoil,
			givenRegisterData: []byte{0x0, 0x0, 0x53, 0x56, 0x43, 0x83},
			expectErr:         "extraction failure due unknown field type",
		},
		{
			name:              "nok, matches invalid",
			whenType:          FieldTypeUint16,
			whenInvalid:       []byte{0xff, 0xff},
			givenRegisterData: []byte{0x0, 0x0, 0xff, 0xff, 0x43, 0x83},
			expectErr:         "invalid value",
		},
		{
			name:      "nok, unknown type",
			whenType:  0,
			expect:    nil,
			expectErr: "extraction failure due unknown field type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := Field{
				Name:            "test",
				ServerAddress:   ":502",
				FunctionCode:    0,
				UnitID:          1,
				Protocol:        0,
				RequestInterval: 0,
				Address:         1,
				Type:            tc.whenType,
				Bit:             8,
				FromHighByte:    true,
				Length:          3,
				ByteOrder:       tc.whenByteOrder,
				Invalid:         tc.whenInvalid,
			}

			registers, _ := packet.NewRegisters(tc.givenRegisterData, 0)

			result, err := f.ExtractFrom(registers)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestField_Validate(t *testing.T) {
	example := Field{
		ServerAddress: ":502",
		UnitID:        1,
		Address:       100,
		Type:          FieldTypeString,
		Bit:           0,
		FromHighByte:  false,
		Length:        10,
		ByteOrder:     0,
		Name:          "fire_alarm_di",
	}
	var testCases = []struct {
		name      string
		given     func(f *Field)
		expectErr string
	}{
		{
			name:  "ok",
			given: func(f *Field) {},
		},
		{
			name:      "nok, server address is empty",
			given:     func(f *Field) { f.ServerAddress = "" },
			expectErr: "field server address can not be empty",
		},
		{
			name:      "nok, type is not set",
			given:     func(f *Field) { f.Type = 0 },
			expectErr: "field type must be set",
		},
		{
			name:      "nok, type is invalid value",
			given:     func(f *Field) { f.Type = 16 },
			expectErr: "field type has invalid value",
		},
		{
			name:      "nok, bit out of range",
			given:     func(f *Field) { f.Bit = 16 },
			expectErr: "field bit value must be in range (0-15)",
		},
		{
			name: "nok, string type must have length",
			given: func(f *Field) {
				f.Type = FieldTypeString
				f.Length = 0
			},
			expectErr: "field with type string must have length set",
		},
		{
			name: "nok, raw bytes type must have length",
			given: func(f *Field) {
				f.Type = FieldTypeRawBytes
				f.Length = 0
			},
			expectErr: "field with type bytes must have length set",
		},
		{
			name: "nok, coil invalid function code",
			given: func(f *Field) {
				f.Type = FieldTypeCoil
				f.FunctionCode = 3
			},
			expectErr: "field with type coil must have function code of 0,1,2",
		},
		{
			name: "nok, invalid protocol",
			given: func(f *Field) {
				f.Protocol = 3
			},
			expectErr: "field has invalid protocol type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := example

			tc.given(&f)

			err := f.Validate()
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldType_UnmarshalJSON(t *testing.T) {
	var testCases = []struct {
		name      string
		given     string
		expect    FieldType
		expectErr string
	}{
		{
			name:   "ok, case",
			given:  `"bIT"`,
			expect: FieldTypeBit,
		},
		{
			name:   "ok, all variants",
			given:  `"byte"`,
			expect: FieldTypeByte,
		},
		{
			name:      "nok, unknown type",
			given:     `"unknown"`,
			expect:    0,
			expectErr: `unknown field type value, given: 'unknown'`,
		},
		{
			name:      "nok, too short",
			given:     `""`,
			expect:    0,
			expectErr: `field type value too short, given: '""'`,
		},
		{
			name:      "nok, wrong start",
			given:     `unknown"`,
			expect:    0,
			expectErr: `field type value does not start with quote mark, given: 'unknown"'`,
		},
		{
			name:      "nok, wrong end",
			given:     `"unknown`,
			expect:    0,
			expectErr: `field type value does not end with quote mark, given: '"unknown'`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result FieldType
			err := result.UnmarshalJSON([]byte(tc.given))

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseFieldType(t *testing.T) {
	var testCases = []struct {
		name      string
		given     string
		expect    FieldType
		expectErr string
	}{
		{
			name:   "ok, case",
			given:  `bIT`,
			expect: FieldTypeBit,
		},
		{
			name:   "ok, byte",
			given:  `byte`,
			expect: FieldTypeByte,
		},
		{
			name:   "ok, uint8",
			given:  `uint8`,
			expect: FieldTypeUint8,
		},
		{
			name:   "ok, int8",
			given:  `int8`,
			expect: FieldTypeInt8,
		},
		{
			name:   "ok, uint16",
			given:  `uint16`,
			expect: FieldTypeUint16,
		},
		{
			name:   "ok, int16",
			given:  `int16`,
			expect: FieldTypeInt16,
		},
		{
			name:   "ok, uint32",
			given:  `uint32`,
			expect: FieldTypeUint32,
		},
		{
			name:   "ok, int32",
			given:  `int32`,
			expect: FieldTypeInt32,
		},
		{
			name:   "ok, uint64",
			given:  `uint64`,
			expect: FieldTypeUint64,
		},
		{
			name:   "ok, int64",
			given:  `int64`,
			expect: FieldTypeInt64,
		},
		{
			name:   "ok, float32",
			given:  `float32`,
			expect: FieldTypeFloat32,
		},
		{
			name:   "ok, float64",
			given:  `float64`,
			expect: FieldTypeFloat64,
		},
		{
			name:   "ok, string",
			given:  `string`,
			expect: FieldTypeString,
		},
		{
			name:   "ok, coil",
			given:  `coil`,
			expect: FieldTypeCoil,
		},
		{
			name:   "ok, bytes",
			given:  `bytes`,
			expect: FieldTypeRawBytes,
		},
		{
			name:      "nok, unknown type",
			given:     `"unknown"`,
			expect:    0,
			expectErr: `unknown field type value, given: '"unknown"'`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseFieldType(tc.given)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestField_CheckInvalid(t *testing.T) {
	var testCases = []struct {
		name              string
		when              Field
		givenData         []byte
		givenStartAddress uint16
		expectErr         string
	}{
		{
			name: "ok",
			when: Field{Address: 3, Type: FieldTypeInt16, Invalid: []byte{0xff, 0xff}},
			givenData: []byte{
				0xff, 0xff, // register 2
				0x0, 0x0, // register 3
				0xff, 0xff, // register 4
			},
			givenStartAddress: 2,
		},
		{
			name: "nok, invalid value",
			when: Field{Address: 4, Type: FieldTypeInt16, Invalid: []byte{0xff, 0xff}},
			givenData: []byte{
				0xff, 0xff, // register 2
				0x0, 0x0, // register 3
				0xff, 0xff, // register 4
			},
			givenStartAddress: 2,
			expectErr:         "invalid value",
		},
		{
			name: "nok, invalid value, multi register check",
			when: Field{Address: 3, Type: FieldTypeInt16, Invalid: []byte{0xff, 0xff, 0xff, 0xff}},
			givenData: []byte{
				0x0, 0x0, // register 2
				0xff, 0xff, // register 3
				0xff, 0xff, // register 4
			},
			givenStartAddress: 2,
			expectErr:         "invalid value",
		},
		{
			name: "nok, address is out of bounds",
			when: Field{Address: 0, Type: FieldTypeInt16, Invalid: []byte{0xff, 0xff, 0xff, 0xff}},
			givenData: []byte{
				0x0, 0x0, // register 2
				0xff, 0xff, // register 3
				0xff, 0xff, // register 4
			},
			givenStartAddress: 2,
			expectErr:         "address under startAddress bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registers, _ := packet.NewRegisters(tc.givenData, tc.givenStartAddress)
			err := tc.when.CheckInvalid(registers)

			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvalid_MarshalJSON(t *testing.T) {
	result, err := Invalid.MarshalJSON([]byte{0xCA, 0xFE})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"cafe"`), result)
}

func TestInvalid_UnmarshalJSON(t *testing.T) {
	var testCases = []struct {
		name      string
		given     string
		expect    []Invalid
		expectErr string
	}{
		{
			name:   "ok, case",
			given:  `["cafe", "BaBe"]`,
			expect: []Invalid{{0xca, 0xfe}, {0xba, 0xbe}},
		},
		{
			name:      "nok, too short",
			given:     `[11]`,
			expect:    []Invalid{Invalid(nil)},
			expectErr: `could not unmarshal Invalid, raw value too short`,
		},
		{
			name:      "nok, not quoted string",
			given:     `[111]`,
			expect:    []Invalid{Invalid(nil)},
			expectErr: `could not unmarshal Invalid, raw value does not seems to be string`,
		},
		{
			name:      "nok, not hex string",
			given:     `["g"]`,
			expect:    []Invalid{Invalid(nil)},
			expectErr: `could not unmarshal Invalid hex string, err: encoding/hex: invalid byte: U+0067 'g'`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result []Invalid
			err := json.Unmarshal([]byte(tc.given), &result)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldBytes(t *testing.T) {
	var testCases = []struct {
		name      string
		given     *Field
		when      any
		expect    []byte
		expectErr string
	}{
		{name: "ok, field type bit",
			when:   true,
			given:  &Field{Type: FieldTypeBit, Bit: 2},
			expect: []byte{0x0, 0x4},
		},
		{name: "ok, field type byte",
			when:   byte(10),
			given:  &Field{Type: FieldTypeByte},
			expect: []byte{0x0, 10},
		},
		{name: "ok, field type uint8, high byte",
			when:   byte(10),
			given:  &Field{Type: FieldTypeByte, FromHighByte: true},
			expect: []byte{10, 0x0},
		},
		{name: "ok, field type Int8",
			when:   uint8(1),
			given:  &Field{Type: FieldTypeInt8},
			expect: []byte{0x0, 1},
		},
		{name: "ok, field type uint16",
			when:   uint16(1),
			given:  &Field{Type: FieldTypeUint16},
			expect: []byte{0x0, 1},
		},
		{name: "ok, field type int16",
			when:   int16(1),
			given:  &Field{Type: FieldTypeInt16},
			expect: []byte{0x0, 1},
		},
		{name: "ok, field type uint32",
			when:   uint32(1),
			given:  &Field{Type: FieldTypeUint32},
			expect: []byte{0x0, 0x0, 0x0, 1},
		},
		{name: "ok, field type int32",
			when:   int32(1),
			given:  &Field{Type: FieldTypeInt32},
			expect: []byte{0x0, 0x0, 0x0, 1},
		},
		{name: "ok, field type uint64",
			when:   uint64(1),
			given:  &Field{Type: FieldTypeUint64},
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 1},
		},
		{name: "ok, field type int64",
			when:   int64(1),
			given:  &Field{Type: FieldTypeInt64},
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 1},
		},
		{name: "ok, field type float32",
			when:   float32(1),
			given:  &Field{Type: FieldTypeFloat32},
			expect: []byte{0x3f, 0x80, 0x0, 0x0},
		},
		{name: "ok, field type float64",
			when:   float64(1),
			given:  &Field{Type: FieldTypeFloat64},
			expect: []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{name: "ok, field type string",
			when:   "ABCD",
			given:  &Field{Type: FieldTypeString, Length: 2},
			expect: []byte{0x42, 0x41},
		},
		{name: "ok, field type string",
			when:   []byte("ABCD"),
			given:  &Field{Type: FieldTypeRawBytes, Length: 2},
			expect: []byte{0x42, 0x41},
		},
		{name: "ok, field type uint32, low word first",
			when:   uint32(1),
			given:  &Field{Type: FieldTypeUint32, ByteOrder: packet.BigEndianLowWordFirst},
			expect: []byte{0x0, 1, 0x0, 0x0},
		},
		{name: "nok, field type coil is unsupported",
			when:      true,
			given:     &Field{Type: FieldTypeCoil},
			expect:    nil,
			expectErr: "coil field type is unsupported for MarshalBytes",
		},
		{name: "nok, unsupported value type",
			when:      errors.New("nope"),
			given:     &Field{Type: FieldTypeUint32},
			expect:    nil,
			expectErr: "marshalFieldTypeUint32: can not marshal unsupported type",
		},
		{name: "nok, unsupported field type",
			when:      errors.New("nope"),
			given:     &Field{Type: 0},
			expect:    nil,
			expectErr: "unsupported field type for MarshalBytes",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.given.MarshalBytes(tc.when)

			assert.Equal(t, tc.expect, b)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
