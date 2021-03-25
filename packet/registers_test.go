package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisters_NewRegisters(t *testing.T) {
	var testCases = []struct {
		name             string
		whenData         []byte
		whenStartAddress uint16
		expect           *Registers
		expectError      string
	}{
		{
			name:             "ok",
			whenData:         []byte{0x1, 0x2},
			whenStartAddress: 1,
			expect: &Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     1,
				endAddress:       2,
				data:             []byte{0x1, 0x2},
			},
		},
		{
			name:             "nok, odd len",
			whenData:         []byte{0x1, 0x2, 0x1},
			whenStartAddress: 1,
			expect:           nil,
			expectError:      "data length must be odd number of bytes as 1 register is 2 bytes",
		},
		{
			name:             "nok, too short len",
			whenData:         []byte{0x1},
			whenStartAddress: 1,
			expect:           nil,
			expectError:      "data length at least 2 bytes as 1 register is 2 bytes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRegisters(tc.whenData, tc.whenStartAddress)
			assert.Equal(t, tc.expect, r)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_WithByteOrder(t *testing.T) {
	r, err := NewRegisters([]byte{0x0, 0x0, 0x0, 0x1}, 0)
	assert.NoError(t, err)
	assert.Equal(t, BigEndianHighWordFirst, r.defaultByteOrder)

	r = r.WithByteOrder(BigEndian)
	assert.Equal(t, BigEndian, r.defaultByteOrder)

	r.WithByteOrder(LittleEndian)
	assert.Equal(t, LittleEndian, r.defaultByteOrder)
}

func TestRegisters_Register(t *testing.T) {
	var testCases = []struct {
		name        string
		whenAddress uint16
		expectError string
		expect      []byte
	}{
		{
			name:        "ok",
			whenAddress: 2,
			expect:      []byte{0x0, 0x1},
		},
		{
			name:        "nok, address out of bound",
			whenAddress: 3,
			expect:      nil,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := NewRegisters([]byte{0x0, 0x2, 0x0, 0x1}, 1)

			result, err := r.Register(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)

				result[0] = 0xFF // should not change original slice
				assert.Equal(t, []byte{0x0, 0x2, 0x0, 0x1}, r.data)
			}
		})
	}
}

func TestRegisters_DoubleRegister(t *testing.T) {
	var testCases = []struct {
		name          string
		whenAddress   uint16
		whenByteOrder ByteOrder
		expectError   string
		expect        []byte
	}{
		{
			name:        "ok (default is high word first)",
			whenAddress: 2,
			expect:      []byte{0x0, 0x2, 0x0, 0x3},
		},
		{
			name:          "ok, low word first",
			whenAddress:   2,
			whenByteOrder: LowWordFirst,
			expect:        []byte{0x0, 0x3, 0x0, 0x2},
		},
		{
			name:        "nok, address out of bound",
			whenAddress: 5,
			expect:      nil,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := NewRegisters([]byte{0x0, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0, 0x4}, 1)

			result, err := r.DoubleRegister(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)

				result[0] = 0xFF // should not change original slice
				assert.Equal(t, []byte{0x0, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0, 0x4}, r.data)
			}
		})
	}
}

func TestRegisters_QuadRegister(t *testing.T) {
	var testCases = []struct {
		name          string
		whenAddress   uint16
		whenByteOrder ByteOrder
		expectError   string
		expect        []byte
	}{
		{
			name:        "ok (default is high word first)",
			whenAddress: 2,
			expect:      []byte{0x0, 0x2, 0x0, 0x3, 0x0, 0x4, 0x0, 0x5},
		},
		{
			name:          "ok, low word first",
			whenAddress:   2,
			whenByteOrder: LowWordFirst,
			expect:        []byte{0x0, 0x5, 0x0, 0x4, 0x0, 0x3, 0x0, 0x2},
		},
		{
			name:        "nok, address out of bound",
			whenAddress: 6,
			expect:      nil,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := NewRegisters([]byte{0x0, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0, 0x4, 0x0, 0x5, 0x0, 0x6, 0x0, 0x7, 0x0, 0x8}, 1)

			result, err := r.QuadRegister(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)

				result[0] = 0xFF // should not change original slice
				assert.Equal(t, []byte{0x0, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0, 0x4, 0x0, 0x5, 0x0, 0x6, 0x0, 0x7, 0x0, 0x8}, r.data)
			}
		})
	}
}

func TestRegisters_Bit(t *testing.T) {
	var testCases = []struct {
		name        string
		whenAddress uint16
		whenBit     uint8
		expect      bool
		expectError string
	}{
		{
			name:        "ok, first byte, first bit",
			whenAddress: 1, whenBit: 0, expect: false,
		},
		{
			name:        "ok, first byte, second bit",
			whenAddress: 1, whenBit: 1, expect: true,
		},
		{
			name:        "ok, second byte, first bit",
			whenAddress: 1, whenBit: 8, expect: true,
		},
		{
			name:        "ok, second byte, last bit",
			whenAddress: 1, whenBit: 15, expect: true,
		},
		{
			name:        "ok, bit out of bounds",
			whenAddress: 1, whenBit: 16, expect: false,
			expectError: "bit value more than register (16bit) contains",
		},
		{
			name:        "ok, address over data bounds",
			whenAddress: 10, whenBit: 9, expect: false,
			expectError: "address over startAddress+quantity bounds",
		},
		{
			name:        "ok, address before data bounds",
			whenAddress: 0, whenBit: 1, expect: false,
			expectError: "address under startAddress bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   2,
				data:         []byte{0b10000001, 0b00010010},
			}

			result, err := r.Bit(tc.whenAddress, tc.whenBit)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Byte(t *testing.T) {
	var testCases = []struct {
		name             string
		whenAddress      uint16
		whenFromHighByte bool
		expect           uint8
		expectError      string
	}{
		{
			name:             "ok, high byte",
			whenAddress:      2,
			whenFromHighByte: true,
			expect:           0xCA,
		},
		{
			name:             "ok, low byte",
			whenAddress:      2,
			whenFromHighByte: false,
			expect:           0xFE,
		},
		{
			name:             "ok, low byte, first register",
			whenAddress:      1,
			whenFromHighByte: false,
			expect:           0x1,
		},
		{
			name:             "ok, high byte, first register",
			whenAddress:      1,
			whenFromHighByte: true,
			expect:           255,
		},
		{
			name:             "ok, low byte, last register",
			whenAddress:      3,
			whenFromHighByte: true,
			expect:           0x3,
		},
		{
			name:             "nok, address before start",
			whenAddress:      0,
			whenFromHighByte: false,
			expect:           0,
			expectError:      "address under startAddress bounds",
		},
		{
			name:             "nok, address over end",
			whenAddress:      4,
			whenFromHighByte: false,
			expect:           0,
			expectError:      "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   4,
				data:         []byte{0xff, 0x1, 0xCA, 0xFE, 0x3, 0x0},
			}

			result, err := r.Byte(tc.whenAddress, tc.whenFromHighByte)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Int8(t *testing.T) {
	var testCases = []struct {
		name             string
		whenAddress      uint16
		whenFromHighByte bool
		expect           int8
		expectError      string
	}{
		{
			name:             "ok, high byte",
			whenAddress:      2,
			whenFromHighByte: true,
			expect:           -127,
		},
		{
			name:             "ok, low byte",
			whenAddress:      2,
			whenFromHighByte: false,
			expect:           65,
		},
		{
			name:             "ok, low byte, first register",
			whenAddress:      1,
			whenFromHighByte: false,
			expect:           1,
		},
		{
			name:             "ok, high byte, first register",
			whenAddress:      1,
			whenFromHighByte: true,
			expect:           -1,
		},
		{
			name:             "ok, low byte, last register",
			whenAddress:      3,
			whenFromHighByte: true,
			expect:           3,
		},
		{
			name:             "nok, address before start",
			whenAddress:      0,
			whenFromHighByte: false,
			expect:           0,
			expectError:      "address under startAddress bounds",
		},
		{
			name:             "nok, address over end",
			whenAddress:      4,
			whenFromHighByte: false,
			expect:           0,
			expectError:      "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   4,
				data:         []byte{0xff, 0x1, 0b10000001, 0b01000001, 0x3, 0x0},
			}

			result, err := r.Int8(tc.whenAddress, tc.whenFromHighByte)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Uint16(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               uint16
		expectError          string
	}{
		{
			name:        "ok, second register",
			whenAddress: 2,
			expect:      32767, // 0x7fff
		},
		{
			name:                 "ok, second register as LE",
			whenAddress:          2,
			whenDefaultByteOrder: LittleEndian,
			expect:               0xff7f,
		},
		{
			name:        "ok, first register",
			whenAddress: 1,
			expect:      65535,
		},
		{
			name:        "ok, last register",
			whenAddress: 3,
			expect:      1,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 4,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     1,
				endAddress:       4,
				data:             []byte{0xff, 0xff, 0x7f, 0xff, 0x0, 0x1},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Uint16(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Int16(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               int16
		expectError          string
	}{
		{
			name:        "ok, second register",
			whenAddress: 2,
			expect:      32767, // 0x7fff
		},
		{
			name:                 "ok, second register as LE",
			whenAddress:          2,
			whenDefaultByteOrder: LittleEndian,
			expect:               -129, // 0xff7f
		},
		{
			name:        "ok, first register",
			whenAddress: 1,
			expect:      -1,
		},
		{
			name:        "ok, last register",
			whenAddress: 3,
			expect:      1,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 4,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   4,
				data:         []byte{0xff, 0xff, 0x7f, 0xff, 0x0, 0x1},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Int16(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Uint32(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               uint32
		expectError          string
	}{
		{
			name:        "ok, second register",
			whenAddress: 2,
			expect:      2147483647, // 0x7fffffff
		},
		{
			name:                 "ok, second register LE",
			whenAddress:          2,
			whenDefaultByteOrder: LittleEndian,
			expect:               4294967167, // 0xffffff7f
		},
		{
			name:        "ok, first register",
			whenAddress: 1,
			expect:      0xffff7fff,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 3,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   4,
				data:         []byte{0xff, 0xff, 0x7f, 0xff, 0xff, 0xff},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Uint32(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Uint32WithByteOrder(t *testing.T) {
	var testCases = []struct {
		name          string
		givenBytes    []byte
		whenAddress   uint16
		whenByteOrder ByteOrder
		expect        uint32
		expectError   string
	}{
		{
			name:          "ok, BE = 1",
			givenBytes:    []byte{0x0, 0x0, 0x0, 0x1},
			whenByteOrder: BigEndian,
			expect:        1,
		},
		{
			name:          "ok, LE = 1",
			givenBytes:    []byte{0x1, 0x0, 0x0, 0x0},
			whenByteOrder: LittleEndian,
			expect:        1,
		},
		{
			name:          "ok, BE high word",
			givenBytes:    []byte{0x7f, 0xff, 0xff, 0xff},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        2147483647,
		},
		{
			name:          "ok, BE low word",
			givenBytes:    []byte{0x7f, 0xff, 0xff, 0xff},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        0xffff7fff,
		},
		{
			name:          "ok, BE low word = 1",
			givenBytes:    []byte{0x0, 0x1, 0x0, 0x0},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE high word = 1",
			givenBytes:    []byte{0x0, 0x0, 0x0, 0x1},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE low word = -2147483648",
			givenBytes:    []byte{0x0, 0x0, 0x80, 0x0},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        2147483648,
		},
		{
			name:          "ok, BE high word = -2147483648",
			givenBytes:    []byte{0x80, 0x0, 0x0, 0x0},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        2147483648,
		},
		{
			name:          "ok, useDefaultByteOrder = BE high word = -2147483648",
			givenBytes:    []byte{0x80, 0x0, 0x0, 0x0},
			whenByteOrder: useDefaultByteOrder,
			expect:        2147483648,
		},
		{
			name:          "ok, LE low word = 1",
			givenBytes:    []byte{0x0, 0x0, 0x1, 0x0},
			whenByteOrder: LittleEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, LE high word = 1",
			givenBytes:    []byte{0x1, 0x0, 0x0, 0x0},
			whenByteOrder: LittleEndianHighWordFirst,
			expect:        1,
		},
		{
			name:        "nok, address over end",
			whenAddress: 2,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 0,
				endAddress:   2,
				data:         tc.givenBytes,
			}
			result, err := r.Uint32WithByteOrder(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Int32(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               int32
		expectError          string
	}{
		{
			name:        "ok, second register",
			whenAddress: 2,
			expect:      2147483647, // 0x7fffffff
		},
		{
			name:                 "ok, second register LE",
			whenAddress:          2,
			whenDefaultByteOrder: LittleEndian,
			expect:               -129, // 0xffffff7f
		},
		{
			name:        "ok, first register",
			whenAddress: 1,
			expect:      -32769,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 3,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   4,
				data:         []byte{0xff, 0xff, 0x7f, 0xff, 0xff, 0xff},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Int32(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Int32WithByteOrder(t *testing.T) {
	var testCases = []struct {
		name          string
		givenBytes    []byte
		whenAddress   uint16
		whenByteOrder ByteOrder
		expect        int32
		expectError   string
	}{
		{
			name:          "ok, BE = 1",
			givenBytes:    []byte{0x0, 0x0, 0x0, 0x1},
			whenByteOrder: BigEndian,
			expect:        1,
		},
		{
			name:          "ok, LE = 1",
			givenBytes:    []byte{0x1, 0x0, 0x0, 0x0},
			whenByteOrder: LittleEndian,
			expect:        1,
		},
		{
			name:          "ok, BE high word",
			givenBytes:    []byte{0x7f, 0xff, 0xff, 0xff},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        2147483647,
		},
		{
			name:          "ok, BE low word",
			givenBytes:    []byte{0x7f, 0xff, 0xff, 0xff},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        -32769,
		},
		{
			name:          "ok, BE low word = 1",
			givenBytes:    []byte{0x0, 0x1, 0x0, 0x0},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE high word = 1",
			givenBytes:    []byte{0x0, 0x0, 0x0, 0x1},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE low word = -2147483648",
			givenBytes:    []byte{0x0, 0x0, 0x80, 0x0},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        -2147483648,
		},
		{
			name:          "ok, BE high word = -2147483648",
			givenBytes:    []byte{0x80, 0x0, 0x0, 0x0},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        -2147483648,
		},
		{
			name:          "ok, useDefaultByteOrder = BE high word = -2147483648",
			givenBytes:    []byte{0x80, 0x0, 0x0, 0x0},
			whenByteOrder: useDefaultByteOrder,
			expect:        -2147483648,
		},
		{
			name:          "ok, LE low word = 1",
			givenBytes:    []byte{0x0, 0x0, 0x1, 0x0},
			whenByteOrder: LittleEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, LE high word = 1",
			givenBytes:    []byte{0x1, 0x0, 0x0, 0x0},
			whenByteOrder: LittleEndianHighWordFirst,
			expect:        1,
		},
		{
			name:        "nok, address over end",
			whenAddress: 2,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 0,
				endAddress:   2,
				data:         tc.givenBytes,
			}
			result, err := r.Int32WithByteOrder(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Uint64(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               uint64
		expectError          string
	}{
		{
			name:        "ok, first register",
			whenAddress: 1,
			expect:      1,
		},
		{
			name:                 "ok, first register LE",
			whenAddress:          1,
			whenDefaultByteOrder: LittleEndian,
			expect:               0x100000000000000,
		},
		{
			name:        "ok, offset register",
			whenAddress: 5,
			expect:      2147483647,
		},
		{
			name:        "ok, 72623859790382856",
			given:       &Registers{startAddress: 1, endAddress: 5, data: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}},
			whenAddress: 1,
			expect:      72623859790382856,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 10,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   9,
				data:         []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x7F, 0xFF, 0xFF, 0xFF},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Uint64(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Uint64WithByteOrder(t *testing.T) {
	var testCases = []struct {
		name          string
		givenBytes    []byte
		whenAddress   uint16
		whenByteOrder ByteOrder
		expect        uint64
		expectError   string
	}{
		{
			name:          "ok, useDefaultByteOrder = BE = BE high word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			whenByteOrder: useDefaultByteOrder,
			expect:        1,
		},
		{
			name:          "ok, BE = BE high word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			whenByteOrder: BigEndian,
			expect:        1,
		},
		{
			name:          "ok, BE high word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE low word = 1",
			givenBytes:    []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, LE = LE high word = 1",
			givenBytes:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: LittleEndian,
			expect:        1,
		},
		{
			name:          "ok, LE high word = 1",
			givenBytes:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: LittleEndianHighWordFirst,
			expect:        1,
		},
		{
			name:          "ok, LE low word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
			whenByteOrder: LittleEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE high word 2147483647",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x7F, 0xFF, 0xFF, 0xFF},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        2147483647,
		},
		{
			name:          "ok, BE high word 72623859790382856",
			givenBytes:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        72623859790382856,
		},
		{
			name:        "nok, address over end",
			whenAddress: 10,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     0,
				endAddress:       9,
				data:             tc.givenBytes,
			}
			result, err := r.Uint64WithByteOrder(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Int64(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               int64
		expectError          string
	}{
		{
			name:        "ok, first register",
			whenAddress: 1,
			expect:      1,
		},
		{
			name:                 "ok, first register LE",
			whenAddress:          1,
			whenDefaultByteOrder: LittleEndian,
			expect:               72057594037927936, // 0x100000000000000
		},
		{
			name:        "ok, offset register",
			whenAddress: 5,
			expect:      9223372036854775807,
		},
		{
			name:        "ok, -1",
			given:       &Registers{startAddress: 1, endAddress: 5, data: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
			whenAddress: 1,
			expect:      -1,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 10,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   9,
				data:         []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Int64(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Int64WithByteOrder(t *testing.T) {
	var testCases = []struct {
		name          string
		givenBytes    []byte
		whenAddress   uint16
		whenByteOrder ByteOrder
		expect        int64
		expectError   string
	}{
		{
			name:          "ok, useDefaultByteOrder = BE = BE high word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			whenByteOrder: useDefaultByteOrder,
			expect:        1,
		},
		{
			name:          "ok, BE = BE high word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			whenByteOrder: BigEndian,
			expect:        1,
		},
		{
			name:          "ok, BE = BE high word = -9223372036854775808",
			givenBytes:    []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: BigEndian,
			expect:        -9223372036854775808,
		},
		{
			name:          "ok, BE high word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE low word = 1",
			givenBytes:    []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, LE = LE high word = 1",
			givenBytes:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: LittleEndian,
			expect:        1,
		},
		{
			name:          "ok, LE = LE high word = -9223372036854775808",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
			whenByteOrder: LittleEndian,
			expect:        -9223372036854775808,
		},
		{
			name:          "ok, LE high word = 1",
			givenBytes:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			whenByteOrder: LittleEndianHighWordFirst,
			expect:        1,
		},
		{
			name:          "ok, LE low word = 1",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
			whenByteOrder: LittleEndianLowWordFirst,
			expect:        1,
		},
		{
			name:          "ok, BE high word 2147483647",
			givenBytes:    []byte{0x00, 0x00, 0x00, 0x00, 0x7F, 0xFF, 0xFF, 0xFF},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        2147483647,
		},
		{
			name:          "ok, BE high word 72623859790382856",
			givenBytes:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        72623859790382856,
		},
		{
			name:        "nok, address over end",
			whenAddress: 10,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     0,
				endAddress:       9,
				data:             tc.givenBytes,
			}
			result, err := r.Int64WithByteOrder(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Float32(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               float32
		expectError          string
	}{
		{
			name:        "ok, second register",
			whenAddress: 1,
			expect:      1.85, // 0x3f, 0xec, 0xcc, 0xcd
		},
		{
			name:                 "ok, second register LE",
			given:                &Registers{startAddress: 1, endAddress: 5, data: []byte{0xcd, 0xcc, 0xec, 0x3f}},
			whenAddress:          1,
			whenDefaultByteOrder: LittleEndian,
			expect:               1.85,
		},
		{
			name:        "ok, offset register",
			whenAddress: 3,
			expect:      0.66666666666,
		},
		{
			name:        "ok, 0",
			given:       &Registers{startAddress: 1, endAddress: 5, data: []byte{0x00, 0x00, 0x00, 0x00}},
			whenAddress: 1,
			expect:      0,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 5,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   5,
				data:         []byte{0x3f, 0xec, 0xcc, 0xcd, 0x3f, 0x2a, 0xaa, 0xab},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Float32(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Float32WithByteOrder(t *testing.T) {
	var testCases = []struct {
		name          string
		givenBytes    []byte
		whenAddress   uint16
		whenByteOrder ByteOrder
		expect        float32
		expectError   string
	}{
		{
			name:          "ok, useDefaultByteOrder = BE = BE high word = 1.85",
			givenBytes:    []byte{0x3f, 0xec, 0xcc, 0xcd},
			whenByteOrder: useDefaultByteOrder,
			expect:        1.85,
		},
		{
			name:          "ok, BE = BE high word = 1.85",
			givenBytes:    []byte{0x3f, 0xec, 0xcc, 0xcd},
			whenByteOrder: BigEndian,
			expect:        1.85,
		},
		{
			name:          "ok, BE high word = 1.85",
			givenBytes:    []byte{0x3f, 0xec, 0xcc, 0xcd},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        1.85,
		},
		{
			name:          "ok, BE low word = 1.85",
			givenBytes:    []byte{0xcc, 0xcd, 0x3f, 0xec},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        1.85,
		},
		{
			name:          "ok, LE = LE high word = 1.85",
			givenBytes:    []byte{0xcd, 0xcc, 0xec, 0x3f},
			whenByteOrder: LittleEndian,
			expect:        1.85,
		},
		{
			name:          "ok, LE high word = 1.85",
			givenBytes:    []byte{0xcd, 0xcc, 0xec, 0x3f},
			whenByteOrder: LittleEndianHighWordFirst,
			expect:        1.85,
		},
		{
			name:          "ok, LE low word = 1.85",
			givenBytes:    []byte{0xec, 0x3f, 0xcd, 0xcc},
			whenByteOrder: LittleEndianLowWordFirst,
			expect:        1.85,
		},
		{
			name:        "nok, address over end",
			whenAddress: 5,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     0,
				endAddress:       5,
				data:             tc.givenBytes,
			}
			result, err := r.Float32WithByteOrder(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Float64(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                *Registers
		whenAddress          uint16
		whenDefaultByteOrder ByteOrder
		expect               float64
		expectError          string
	}{
		{
			name:        "ok, second register",
			whenAddress: 1,
			expect:      1.85, // 0x3f, 0xfd, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a
		},
		{
			name:                 "ok, second register LE",
			whenAddress:          1,
			given:                &Registers{startAddress: 1, endAddress: 5, data: []byte{0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0xfd, 0x3f}},
			whenDefaultByteOrder: LittleEndian,
			expect:               1.85,
		},
		{
			name:        "ok, offset register",
			whenAddress: 5,
			expect:      0.66666666666, // 0x3f, 0xe5, 0x55, 0x55, 0x55, 0x54, 0x6a, 0xc5
		},
		{
			name:        "ok, 0",
			given:       &Registers{startAddress: 1, endAddress: 5, data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			whenAddress: 1,
			expect:      0,
		},
		{
			name:        "nok, address before start",
			whenAddress: 0,
			expect:      0,
			expectError: "address under startAddress bounds",
		},
		{
			name:        "nok, address over end",
			whenAddress: 6, // 6 + 4 = 10
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				startAddress: 1,
				endAddress:   9,
				data:         []byte{0x3f, 0xfd, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a, 0x3f, 0xe5, 0x55, 0x55, 0x55, 0x54, 0x6a, 0xc5},
			}
			if tc.given != nil {
				r = *tc.given
			}
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.Float64(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_Float64WithByteOrder(t *testing.T) {
	var testCases = []struct {
		name          string
		givenBytes    []byte
		whenAddress   uint16
		whenByteOrder ByteOrder
		expect        float64
		expectError   string
	}{
		{
			name:          "ok, useDefaultByteOrder = BE = BE high word = 1.85",
			givenBytes:    []byte{0x3f, 0xfd, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a},
			whenByteOrder: useDefaultByteOrder,
			expect:        1.85,
		},
		{
			name:          "ok, BE = BE high word = 1.85",
			givenBytes:    []byte{0x3f, 0xfd, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a},
			whenByteOrder: BigEndian,
			expect:        1.85,
		},
		{
			name:          "ok, BE high word = 1.85",
			givenBytes:    []byte{0x3f, 0xfd, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a},
			whenByteOrder: BigEndianHighWordFirst,
			expect:        1.85,
		},
		{
			name:          "ok, BE low word = 1.85",
			givenBytes:    []byte{0x99, 0x9a, 0x99, 0x99, 0x99, 0x99, 0x3f, 0xfd},
			whenByteOrder: BigEndianLowWordFirst,
			expect:        1.85,
		},
		{
			name:          "ok, LE = LE high word = 1.85",
			givenBytes:    []byte{0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0xfd, 0x3f},
			whenByteOrder: LittleEndian,
			expect:        1.85,
		},
		{
			name:          "ok, LE high word = 1.85",
			givenBytes:    []byte{0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0xfd, 0x3f},
			whenByteOrder: LittleEndianHighWordFirst,
			expect:        1.85,
		},
		{
			name:          "ok, LE low word = 1.85",
			givenBytes:    []byte{0xfd, 0x3f, 0x99, 0x99, 0x99, 0x99, 0x9a, 0x99},
			whenByteOrder: LittleEndianLowWordFirst,
			expect:        1.85,
		},
		{
			name:        "nok, address over end",
			whenAddress: 10,
			expect:      0,
			expectError: "address over startAddress+quantity bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     0,
				endAddress:       9,
				data:             tc.givenBytes,
			}
			result, err := r.Float64WithByteOrder(tc.whenAddress, tc.whenByteOrder)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisters_string(t *testing.T) {
	var testCases = []struct {
		name                 string
		given                Registers
		whenDefaultByteOrder ByteOrder
		address              uint16
		length               uint8
		expected             string
		expectedErr          string
	}{
		{
			name:     "BigEndian: string, string is in the middle of data",
			given:    Registers{data: []byte{0x0, 0x0, 0x56, 0x53, 0x0, 0x43, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x83, 0x83}},
			address:  1,
			length:   10, // 10 bytes = 5 registers
			expected: "SVC",
		},
		{
			name:     "BigEndian: string, string is in the end of data, odd length",
			given:    Registers{data: []byte{0x0, 0x0, 0x56, 0x53, 0x83, 0x43}},
			address:  1,
			length:   3, // 3 bytes = 2 registers
			expected: "SVC",
		},
		{
			name:                 "LittleEndian: string, string is in the end of data, odd length",
			given:                Registers{data: []byte{0x0, 0x0, 0x53, 0x56, 0x43, 0x83}},
			whenDefaultByteOrder: LittleEndian,
			address:              1,
			length:               3, // 3 bytes = 2 registers
			expected:             "SVC",
		},
		{
			name:     "BigEndian: string, string is in the end of data",
			given:    Registers{data: []byte{0x0, 0x0, 0x56, 0x53, 0x43, 0x43}},
			address:  1,
			length:   2, // 2 bytes = 1 registers
			expected: "SV",
		},
		{
			name:        "BigEndian: address before start",
			given:       Registers{startAddress: 2, data: []byte{0x0, 0x0, 0x56, 0x53, 0x43, 0x43}},
			address:     1,
			length:      2,
			expected:    "",
			expectedErr: "address under startAddress bounds",
		},
		{
			name:        "BigEndian: length over data bounds",
			given:       Registers{startAddress: 1, data: []byte{0x0, 0x0, 0x56, 0x53, 0x43, 0x43}},
			address:     1,
			length:      7,
			expected:    "",
			expectedErr: "address over data bounds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.given
			r.defaultByteOrder = BigEndianHighWordFirst
			if tc.whenDefaultByteOrder != 0 {
				r.WithByteOrder(tc.whenDefaultByteOrder)
			}

			result, err := r.String(tc.address, tc.length)

			if err != nil || tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
