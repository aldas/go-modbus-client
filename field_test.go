package modbus

import (
	"encoding/json"
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
	fByte := &Field{Type: FieldTypeByte, FromHighByte: false}
	fByteHigh := &Field{Type: FieldTypeByte, FromHighByte: true}
	fUint8 := &Field{Type: FieldTypeUint8, FromHighByte: false}
	fUint8High := &Field{Type: FieldTypeUint8, FromHighByte: true}
	fInt8 := &Field{Type: FieldTypeInt8, FromHighByte: false}
	fInt8High := &Field{Type: FieldTypeInt8, FromHighByte: true}
	fUint16 := &Field{Type: FieldTypeUint16, FromHighByte: false}
	fUint16High := &Field{Type: FieldTypeUint16, FromHighByte: true}
	fInt16 := &Field{Type: FieldTypeInt16, FromHighByte: false}
	fInt16High := &Field{Type: FieldTypeInt16, FromHighByte: true}
	fUint32 := &Field{Type: FieldTypeUint32, FromHighByte: false}
	fUint32High := &Field{Type: FieldTypeUint32, FromHighByte: true}

	var testCases = []struct {
		name      string
		given     *Field
		when      any
		expect    []byte
		expectErr string
	}{
		// ----------------- FieldTypeByte -----------------
		{name: "ok, byte from bool, false value",
			when: false, given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from bool, true value",
			when: true, given: fByte, expect: []byte{0x0, 1}},
		{name: "ok, byte from bool, from high byte, true value",
			when: true, given: fByteHigh, expect: []byte{1, 0x0}},

		{name: "ok, byte from uint8, empty value",
			when: uint8(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from uint8, non empty value",
			when: uint8(10), given: fByte, expect: []byte{0x0, 10}},
		{name: "ok, byte from uint8, from high byte, non empty value",
			when: uint8(10), given: fByteHigh, expect: []byte{10, 0x0}},
		{name: "ok, byte from byte, from high byte, non empty value",
			when: byte(10), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from int8, empty value",
			when: int8(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from int8, non empty value",
			when: int8(10), given: fByte, expect: []byte{0x0, 10}},
		{name: "ok, byte from int8, from high byte, non empty value",
			when: int8(10), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from uint16, empty value",
			when: uint16(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from uint16, non empty value",
			when: uint16(0xff0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from uint16, from high byte, non empty value",
			when: uint16(0xff0a), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from int16, empty value",
			when: int16(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from int16, non empty value",
			when: int16(0x1f0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from int16, from high byte, non empty value",
			when: int16(0x1f0a), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from uint32, empty value",
			when: uint32(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from uint32, non empty value",
			when: uint32(0xff0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from uint32, from high byte, non empty value",
			when: uint32(0xff0a), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from int32, empty value",
			when: int32(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from int32, non empty value",
			when: int32(0x1f0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from int32, from high byte, non empty value",
			when: int32(0x1f0a), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from uint64, empty value",
			when: uint64(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from uint64, non empty value",
			when: uint64(0xff0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from uint64, from high byte, non empty value",
			when: uint64(0xff0a), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from int64, empty value",
			when: int64(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from int64, non empty value",
			when: int64(300), given: fByte, expect: []byte{0x0, 0xff}}, // truncate to max uint8
		{name: "ok, byte from int64, from high byte, non empty value",
			when: int64(300), given: fByteHigh, expect: []byte{0xff, 0x0}}, // truncate to max uint8

		{name: "ok, byte from float32, empty value",
			when: float32(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from float32, non empty value",
			when: float32(0xff0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from float32, decimals, non empty value",
			when: float32(5.1234), given: fByte, expect: []byte{0x0, 5}}, // use only integer part
		{name: "ok, byte from float32, from high byte, non empty value",
			when: float32(0xff0a), given: fByteHigh, expect: []byte{10, 0x0}},

		{name: "ok, byte from float64, empty value",
			when: float64(0), given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from float64, non empty value",
			when: float64(0xff0a), given: fByte, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, byte from float64, decimals, non empty value",
			when: float64(5.1234), given: fByte, expect: []byte{0x0, 5}}, // use only integer part
		{name: "ok, byte from float64, from high byte, non empty value",
			when: float64(0xff0a), given: fByteHigh, expect: []byte{10, 0x0}},

		// nope = 0x6e 0x6f 0x70 0x65 (Little endian)
		{name: "ok, byte from string, empty value",
			when: "", given: fByte, expect: []byte{0x0, 0x0}},
		{name: "ok, byte from string, non empty value",
			when: "nope", given: fByte, expect: []byte{0x0, 0x6e}},
		{name: "ok, byte from string, from high byte, non empty value",
			when: "nope", given: fByteHigh, expect: []byte{0x6e, 0x0}},

		{name: "ok, byte from []byte, empty value",
			when: []byte{0x1, 0x2}, given: fByte, expect: []byte{0x0, 0x1}},
		{name: "ok, byte from []byte, non empty value",
			when: []byte{0x1, 0x2}, given: fByte, expect: []byte{0x0, 0x1}},
		{name: "ok, byte from []byte, from high byte, non empty value",
			when: []byte{0x1, 0x2}, given: fByteHigh, expect: []byte{0x1, 0x0}},

		// ----------------- FieldTypeUint8 -----------------
		{name: "ok, Uint8 from bool, false value",
			when: false, given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from bool, true value",
			when: true, given: fUint8, expect: []byte{0x0, 1}},
		{name: "ok, Uint8 from bool, from high byte, true value",
			when: true, given: fUint8High, expect: []byte{1, 0x0}},

		{name: "ok, Uint8 from uint8, empty value",
			when: uint8(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from uint8, non empty value",
			when: uint8(10), given: fUint8, expect: []byte{0x0, 10}},
		{name: "ok, Uint8 from uint8, from high byte, non empty value",
			when: uint8(10), given: fUint8High, expect: []byte{10, 0x0}},
		{name: "ok, Uint8 from byte, from high byte, non empty value",
			when: byte(10), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from int8, empty value",
			when: int8(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from int8, non empty value",
			when: int8(10), given: fUint8, expect: []byte{0x0, 10}},
		{name: "ok, Uint8 from int8, from high byte, non empty value",
			when: int8(10), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from uint16, empty value",
			when: uint16(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from uint16, non empty value",
			when: uint16(0xff0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from uint16, from high byte, non empty value",
			when: uint16(0xff0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from int16, empty value",
			when: int16(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from int16, non empty value",
			when: int16(0x1f0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from int16, from high byte, non empty value",
			when: int16(0x1f0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from uint32, empty value",
			when: uint32(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from uint32, non empty value",
			when: uint32(0xff0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from uint32, from high byte, non empty value",
			when: uint32(0xff0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from int32, empty value",
			when: int32(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from int32, non empty value",
			when: int32(0x1f0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from int32, from high byte, non empty value",
			when: int32(0x1f0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from uint64, empty value",
			when: uint64(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from uint64, non empty value",
			when: uint64(0xff0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from uint64, from high byte, non empty value",
			when: uint64(0xff0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from int64, empty value",
			when: int64(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from int64, non empty value",
			when: int64(0xff0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from int64, from high byte, non empty value",
			when: int64(0xff0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from float32, empty value",
			when: float32(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from float32, non empty value",
			when: float32(0xff0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from float32, decimals, non empty value",
			when: float32(5.1234), given: fUint8, expect: []byte{0x0, 5}}, // use only integer part
		{name: "ok, Uint8 from float32, from high byte, non empty value",
			when: float32(0xff0a), given: fUint8High, expect: []byte{10, 0x0}},

		{name: "ok, Uint8 from float64, empty value",
			when: float64(0), given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from float64, non empty value",
			when: float64(0xff0a), given: fUint8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Uint8 from float64, decimals, non empty value",
			when: float64(5.1234), given: fUint8, expect: []byte{0x0, 5}}, // use only integer part
		{name: "ok, Uint8 from float64, from high byte, non empty value",
			when: float64(0xff0a), given: fUint8High, expect: []byte{10, 0x0}},

		// nope = 0x6e 0x6f 0x70 0x65 (Little endian)
		{name: "ok, Uint8 from string, empty value",
			when: "", given: fUint8, expect: []byte{0x0, 0x0}},
		{name: "ok, Uint8 from string, non empty value",
			when: "nope", given: fUint8, expect: []byte{0x0, 0x6e}},
		{name: "ok, Uint8 from string, from high byte, non empty value",
			when: "nope", given: fUint8High, expect: []byte{0x6e, 0x0}},

		{name: "ok, Uint8 from []byte, empty value",
			when: []byte{0x1, 0x2}, given: fUint8, expect: []byte{0x0, 0x1}},
		{name: "ok, Uint8 from []byte, non empty value",
			when: []byte{0x1, 0x2}, given: fUint8, expect: []byte{0x0, 0x1}},
		{name: "ok, Uint8 from []byte, from high byte, non empty value",
			when: []byte{0x1, 0x2}, given: fUint8High, expect: []byte{0x1, 0x0}},

		// ----------------- FieldTypeInt8 -----------------
		{name: "ok, Int8 from bool, false value",
			when: false, given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from bool, true value",
			when: true, given: fInt8, expect: []byte{0x0, 1}},
		{name: "ok, Int8 from bool, from high byte, true value",
			when: true, given: fInt8High, expect: []byte{1, 0x0}},

		{name: "ok, Int8 from uint8, empty value",
			when: uint8(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from uint8, non empty value",
			when: uint8(10), given: fInt8, expect: []byte{0x0, 10}},
		{name: "ok, Int8 from uint8, from high byte, non empty value",
			when: uint8(10), given: fInt8High, expect: []byte{10, 0x0}},
		{name: "ok, Int8 from byte, from high byte, non empty value",
			when: byte(10), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from int8, empty value",
			when: int8(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from int8, non empty value",
			when: int8(10), given: fInt8, expect: []byte{0x0, 10}},
		{name: "ok, Int8 from int8, from high byte, non empty value",
			when: int8(10), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from uint16, empty value",
			when: uint16(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from uint16, non empty value",
			when: uint16(0xff0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from uint16, from high byte, non empty value",
			when: uint16(0xff0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from int16, empty value",
			when: int16(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from int16, non empty value",
			when: int16(0x1f0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from int16, from high byte, non empty value",
			when: int16(0x1f0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from uint32, empty value",
			when: uint32(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from uint32, non empty value",
			when: uint32(0xff0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from uint32, from high byte, non empty value",
			when: uint32(0xff0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from int32, empty value",
			when: int32(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from int32, non empty value",
			when: int32(0x1f0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from int32, from high byte, non empty value",
			when: int32(0x1f0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from uint64, empty value",
			when: uint64(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from uint64, non empty value",
			when: uint64(0xff0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from uint64, from high byte, non empty value",
			when: uint64(0xff0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from int64, empty value",
			when: int64(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from int64, non empty value",
			when: int64(0xff0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from int64, from high byte, non empty value",
			when: int64(0xff0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from float32, empty value",
			when: float32(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from float32, non empty value",
			when: float32(0xff0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from float32, decimals, non empty value",
			when: float32(5.1234), given: fInt8, expect: []byte{0x0, 5}}, // use only integer part
		{name: "ok, Int8 from float32, from high byte, non empty value",
			when: float32(0xff0a), given: fInt8High, expect: []byte{10, 0x0}},

		{name: "ok, Int8 from float64, empty value",
			when: float64(0), given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from float64, non empty value",
			when: float64(0xff0a), given: fInt8, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, Int8 from float64, decimals, non empty value",
			when: float64(5.1234), given: fInt8, expect: []byte{0x0, 5}}, // use only integer part
		{name: "ok, Int8 from float64, from high byte, non empty value",
			when: float64(0xff0a), given: fInt8High, expect: []byte{10, 0x0}},

		// nope = 0x6e 0x6f 0x70 0x65 (Little endian)
		{name: "ok, Int8 from string, empty value",
			when: "", given: fInt8, expect: []byte{0x0, 0x0}},
		{name: "ok, Int8 from string, non empty value",
			when: "nope", given: fInt8, expect: []byte{0x0, 0x6e}},
		{name: "ok, Int8 from string, from high byte, non empty value",
			when: "nope", given: fInt8High, expect: []byte{0x6e, 0x0}},

		{name: "ok, Int8 from []byte, empty value",
			when: []byte{0x1, 0x2}, given: fInt8, expect: []byte{0x0, 0x1}},
		{name: "ok, Int8 from []byte, non empty value",
			when: []byte{0x1, 0x2}, given: fInt8, expect: []byte{0x0, 0x1}},
		{name: "ok, Int8 from []byte, from high byte, non empty value",
			when: []byte{0x1, 0x2}, given: fInt8High, expect: []byte{0x1, 0x0}},

		// ----------------- FieldTypeUint16 -----------------
		{name: "ok, uint16 from bool, false value",
			when: false, given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from bool, true value",
			when: true, given: fUint16, expect: []byte{0x0, 1}},
		{name: "ok, uint16 from bool, from high byte, true value",
			when: true, given: fUint16High, expect: []byte{0x0, 1}}, // should not change anything

		{name: "ok, uint16 from uint8, empty value",
			when: uint8(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from uint8, non empty value",
			when: uint8(10), given: fUint16, expect: []byte{0x0, 10}},

		{name: "ok, uint16 from int8, empty value",
			when: int8(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from int8, non empty value",
			when: int8(10), given: fUint16, expect: []byte{0x0, 10}},

		{name: "ok, uint16 from uint16, empty value",
			when: uint16(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from uint16, non empty value",
			when: uint16(0xff0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, uint16 from int16, empty value",
			when: int16(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from int16, non empty value",
			when: int16(0x1f0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, uint16 from uint32, empty value",
			when: uint32(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from uint32, non empty value",
			when: uint32(0xff0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, uint16 from int32, empty value",
			when: int32(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from int32, non empty value",
			when: int32(0x1f0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, uint16 from uint64, empty value",
			when: uint64(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from uint64, non empty value",
			when: uint64(0xff0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, uint16 from int64, empty value",
			when: int64(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from int64, non empty value",
			when: int64(0xff0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, uint16 from float32, empty value",
			when: float32(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from float32, non empty value",
			when: float32(0xff0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, uint16 from float32, decimals, non empty value",
			when: float32(5.1234), given: fUint16, expect: []byte{0x0, 5}}, // use only integer part

		{name: "ok, uint16 from float64, empty value",
			when: float64(0), given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from float64, non empty value",
			when: float64(0xff0a), given: fUint16, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, uint16 from float64, decimals, non empty value",
			when: float64(5.1234), given: fUint16, expect: []byte{0x0, 5}}, // use only integer part

		// nope = 0x6e 0x6f 0x70 0x65 (Little endian)
		{name: "ok, uint16 from string, empty value",
			when: "", given: fUint16, expect: []byte{0x0, 0x0}},
		{name: "ok, uint16 from string, non empty value",
			when: "nope", given: fUint16, expect: []byte{0x0, 0x6e}},

		{name: "ok, uint16 from []byte, empty value",
			when: []byte{0x1, 0x2}, given: fUint16, expect: []byte{0x0, 0x1}},
		{name: "ok, uint16 from []byte, non empty value",
			when: []byte{0x1, 0x2}, given: fUint16, expect: []byte{0x0, 0x1}},

		// ----------------- FieldTypeInt16 -----------------
		{name: "ok, int16 from bool, false value",
			when: false, given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from bool, true value",
			when: true, given: fInt16, expect: []byte{0x0, 1}},
		{name: "ok, int16 from bool, from high byte, true value",
			when: true, given: fInt16High, expect: []byte{0x0, 1}}, // should not change anything

		{name: "ok, int16 from uint8, empty value",
			when: uint8(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from uint8, non empty value",
			when: uint8(10), given: fInt16, expect: []byte{0x0, 10}},

		{name: "ok, int16 from int8, empty value",
			when: int8(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from int8, non empty value",
			when: int8(10), given: fInt16, expect: []byte{0x0, 10}},

		{name: "ok, int16 from uint16, empty value",
			when: uint16(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from uint16, non empty value",
			when: uint16(0xff0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, int16 from int16, empty value",
			when: int16(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from int16, non empty value",
			when: int16(0x1f0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, int16 from uint32, empty value",
			when: uint32(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from uint32, non empty value",
			when: uint32(0xff0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, int16 from int32, empty value",
			when: int32(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from int32, non empty value",
			when: int32(0x1f0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, int16 from uint64, empty value",
			when: uint64(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from uint64, non empty value",
			when: uint64(0xff0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, int16 from int64, empty value",
			when: int64(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from int64, non empty value",
			when: int64(0xff0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte

		{name: "ok, int16 from float32, empty value",
			when: float32(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from float32, non empty value",
			when: float32(0xff0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, int16 from float32, decimals, non empty value",
			when: float32(5.1234), given: fInt16, expect: []byte{0x0, 5}}, // use only integer part

		{name: "ok, int16 from float64, empty value",
			when: float64(0), given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from float64, non empty value",
			when: float64(0xff0a), given: fInt16, expect: []byte{0x0, 10}}, // only first byte
		{name: "ok, int16 from float64, decimals, non empty value",
			when: float64(5.1234), given: fInt16, expect: []byte{0x0, 5}}, // use only integer part

		// nope = 0x6e 0x6f 0x70 0x65 (Little endian)
		{name: "ok, int16 from string, empty value",
			when: "", given: fInt16, expect: []byte{0x0, 0x0}},
		{name: "ok, int16 from string, non empty value",
			when: "nope", given: fInt16, expect: []byte{0x0, 0x6e}},

		{name: "ok, int16 from []byte, empty value",
			when: []byte{0x1, 0x2}, given: fInt16, expect: []byte{0x0, 0x1}},
		{name: "ok, int16 from []byte, non empty value",
			when: []byte{0x1, 0x2}, given: fInt16, expect: []byte{0x0, 0x1}},

		// ----------------- FieldTypeUint32 -----------------
		{name: "ok, uint32 from bool, false value",
			when: false, given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from bool, true value",
			when: true, given: fUint32, expect: []byte{0x0, 0x0, 0x0, 1}},
		{name: "ok, uint32 from bool, from high byte, true value",
			when: true, given: fUint32High, expect: []byte{0x0, 0x0, 0x0, 1}}, // should not change anything

		{name: "ok, uint32 from uint8, empty value",
			when: uint8(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from uint8, non empty value",
			when: uint8(10), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 10}},

		{name: "ok, uint32 from int8, empty value",
			when: int8(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from int8, non empty value",
			when: int8(10), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 10}},

		{name: "ok, uint32 from uint16, empty value",
			when: uint16(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from uint16, non empty value",
			when: uint16(0xff0a), given: fUint32, expect: []byte{0x0, 0x0, 0xff, 10}},

		{name: "ok, uint32 from int16, empty value",
			when: int16(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from int16, non empty value",
			when: int16(0x0102), given: fUint32, expect: []byte{0x0, 0x0, 0x1, 0x2}},

		{name: "ok, uint32 from uint32, empty value",
			when: uint32(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from uint32, non empty value",
			when: uint32(0x01020304), given: fUint32, expect: []byte{0x1, 0x2, 0x3, 0x4}},

		{name: "ok, uint32 from int32, empty value",
			when: int32(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from int32, non empty value",
			when: int32(0x01020304), given: fUint32, expect: []byte{0x1, 0x2, 0x3, 0x4}},

		{name: "ok, uint32 from uint64, empty value",
			when: uint64(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from uint64, non empty value",
			when: uint64(0x0102030405060708), given: fUint32, expect: []byte{1, 2, 3, 4, 5, 6, 7, 8}},

		{name: "ok, uint32 from int64, empty value",
			when: int64(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from int64, non empty value",
			when: int64(0x0102030405060708), given: fUint32, expect: []byte{1, 2, 3, 4, 5, 6, 7, 8}},

		{name: "ok, uint32 from float32, empty value",
			when: float32(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from float32, non empty value",
			when: float32(0x01020304), given: fUint32, expect: []byte{1, 2, 3, 4}},
		{name: "ok, uint32 from float32, decimals, non empty value",
			when: float32(5.1234), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 5}}, // use only integer part

		{name: "ok, uint32 from float64, empty value",
			when: float64(0), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from float64, non empty value",
			when: float64(0x0102030405060708), given: fUint32, expect: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
		{name: "ok, uint32 from float64, decimals, non empty value",
			when: float64(5.1234), given: fUint32, expect: []byte{0x0, 0x0, 0x0, 5}}, // use only integer part

		// nope = 0x6e 0x6f 0x70 0x65 (Little endian)
		{name: "ok, uint32 from string, empty value",
			when: "", given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from string, non empty value",
			when: "nope", given: fUint32, expect: []byte{0x6e, 0x6f, 0x70, 0x65}},

		{name: "ok, uint32 from []byte, empty value",
			when: []byte{}, given: fUint32, expect: []byte{0x0, 0x0, 0x0, 0x0}},
		{name: "ok, uint32 from []byte, non empty value",
			when: []byte{0x1, 0x2, 0x3, 0x4}, given: fUint32, expect: []byte{0x1, 0x2, 0x3, 0x4}},
	}

	// - bool
	// - uint32,int32
	// - uint64,int64
	// - float32,float64
	// - string (Note: raw utf8 bytes. If you need ASCII, convert the string before)
	// - []byte
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.given.MarshalBytes(tc.when, true)

			assert.Equal(t, tc.expect, b)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
