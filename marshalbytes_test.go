package modbus

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestSize(t *testing.T) {
	var testCases = []struct {
		name            string
		when            any
		expectSizeBytes int
		expectIsNumber  bool
		expectErr       string
	}{
		{name: "ok, bool", when: true, expectSizeBytes: 1, expectIsNumber: true},
		{name: "ok, uint8", when: uint8(1), expectSizeBytes: 1, expectIsNumber: true},
		{name: "ok, int8", when: int8(1), expectSizeBytes: 1, expectIsNumber: true},
		{name: "ok, uint16", when: uint16(1), expectSizeBytes: 2, expectIsNumber: true},
		{name: "ok, int16", when: int16(1), expectSizeBytes: 2, expectIsNumber: true},
		{name: "ok, uint32", when: uint32(1), expectSizeBytes: 4, expectIsNumber: true},
		{name: "ok, int32", when: int32(1), expectSizeBytes: 4, expectIsNumber: true},
		{name: "ok, uint64", when: uint64(1), expectSizeBytes: 8, expectIsNumber: true},
		{name: "ok, int64", when: int64(1), expectSizeBytes: 8, expectIsNumber: true},
		{name: "ok, float32", when: float32(1), expectSizeBytes: 4, expectIsNumber: true},
		{name: "ok, float64", when: float64(1), expectSizeBytes: 8, expectIsNumber: true},
		{name: "ok, string", when: "test", expectSizeBytes: 4, expectIsNumber: false},
		{name: "ok, string empty", when: "", expectSizeBytes: 0, expectIsNumber: false},
		{name: "ok, bytes", when: []byte{0x1}, expectSizeBytes: 1, expectIsNumber: false},
		{name: "ok, bytes empty", when: []byte{}, expectSizeBytes: 0, expectIsNumber: false},
		{name: "ok, int", when: int(1), expectSizeBytes: 8, expectIsNumber: true},   // will fail on 32bit arch
		{name: "ok, uint", when: uint(1), expectSizeBytes: 8, expectIsNumber: true}, // will fail on 32bit arch
		{name: "ok, not supported", when: errors.New("x"), expectErr: "can not marshal unsupported type"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sizeBytes, isNumber, err := size(tc.when)

			assert.Equal(t, tc.expectSizeBytes, sizeBytes)
			assert.Equal(t, tc.expectIsNumber, isNumber)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeInt8(t *testing.T) {
	var testCases = []struct {
		name         string
		given        []byte
		when         any
		whenHighByte bool
		expect       []byte
		expectErr    string
	}{
		{
			name:   "ok, bool(true) to int8",
			when:   true,
			expect: []byte{0x0, 0x1},
		},
		{
			name:         "ok, bool(true) to int8, high byte",
			when:         true,
			whenHighByte: true,
			expect:       []byte{0x1, 0x0},
		},
		{
			name: "ok, bool(false) to int8",
			when: false,
		},
		{
			name:   "ok, uint8 to int8",
			when:   uint8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint8(0) to int8",
			when: uint8(0),
		},
		{
			name:   "ok, int8 to int8",
			when:   int8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int8(0) to int8",
			when: int8(0),
		},
		{
			name:   "ok, int8(-1) negative to 0 to int8",
			when:   int8(-1),
			expect: []byte{0x0, 0xff},
		},
		{
			name:   "ok, int8(max neg) to int8",
			when:   int8(math.MinInt8),
			expect: []byte{0x0, 0x80},
		},
		{
			name:   "ok, uint16 to int8",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, uint16(0) to int8",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to int8",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, int16(0) to int8",
			when: int16(0),
		},
		{
			name:   "ok, int16(-1) negative to 0 to int8",
			when:   int16(-1),
			expect: []byte{0x0, 0xff},
		},
		{
			name:   "ok, int16(max neg) to int8",
			when:   int16(math.MinInt16),
			expect: []byte{0x0, 0x80},
		},
		{
			name:   "ok, uint32 to int8",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, uint32(0) to int8",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to int8",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, int32(0) to int8",
			when: int32(0),
		},
		{
			name:   "ok, int32(-1) negative to 0 to int8",
			when:   int32(-1),
			expect: []byte{0x0, 0xff},
		},
		{
			name:   "ok, int32(max neg) to int8",
			when:   int32(math.MinInt32),
			expect: []byte{0x0, 0x80},
		},
		{
			name:   "ok, uint64 to int8",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, uint64(0) to int8",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to int8",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, int64(0) to int8",
			when: int64(0),
		},
		{
			name:   "ok, int64(-1) negative to 0 to int8",
			when:   int64(-1),
			expect: []byte{0x0, 0xff},
		},
		{
			name:   "ok, int64(max neg) to int8",
			when:   int64(math.MinInt64),
			expect: []byte{0x0, 0x80},
		},
		{
			name:   "ok, float32 to int8",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, float32(0) to int8",
			when: float32(0),
		},
		{
			name:   "ok, float32(-1) negative to 0 to int8",
			when:   float32(-1),
			expect: []byte{0x0, 0xff},
		},
		{
			name:   "ok, float32(max neg) to int8",
			when:   float32(-math.MaxFloat32),
			expect: []byte{0x0, 0x80},
		},
		{
			name:   "ok, float64 to int8",
			when:   math.MaxFloat64,
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name: "ok, float64(0) to int8",
			when: float64(0),
		},
		{
			name:   "ok, float64(max neg) to int8",
			when:   float64(-math.MaxFloat64),
			expect: []byte{0x0, 0x80},
		},
		{
			name:   "ok, uint to int8",
			when:   uint(math.MaxUint),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name:   "ok, int to int8",
			when:   int(math.MaxInt),
			expect: []byte{0x0, 0x7f}, // limit to max int8
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0},
			when:      uint8(1),
			expect:    []byte{0x0},
			expectErr: "field type int8 requires at least 2 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeInt8: can not marshal unsupported type",
		},
		{
			name:      "nok, string is unsupported type",
			when:      []byte{},
			expectErr: "marshalFieldTypeInt8: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeInt8(dst, tc.when, tc.whenHighByte)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeUint8(t *testing.T) {
	var testCases = []struct {
		name         string
		given        []byte
		when         any
		whenHighByte bool
		expect       []byte
		expectErr    string
	}{
		{
			name:   "ok, bool(true) to uint8",
			when:   true,
			expect: []byte{0x0, 0x1},
		},
		{
			name:         "ok, bool(true) to uint8 high byte",
			when:         true,
			whenHighByte: true,
			expect:       []byte{0x1, 0x0},
		},
		{
			name: "ok, bool(false) to uint8",
			when: false,
		},
		{
			name:   "ok, uint8 to uint8",
			when:   uint8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint8(0) to uint8",
			when: uint8(0),
		},
		{
			name:   "ok, int8 to uint8",
			when:   int8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int8(0) to uint8",
			when: int8(0),
		},
		{
			name:   "ok, int8(-1) negative to 0 to uint8",
			when:   int8(-1),
			expect: []byte{0x0, 0xff},
		},
		{
			name:   "ok, uint16 to uint8",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, uint16(0) to uint8",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to uint8",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, int16(0) to uint8",
			when: int16(0),
		},
		{
			name: "ok, int16(-1) negative to 0 to uint8",
			when: int16(-1),
		},
		{
			name:   "ok, uint32 to uint8",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, uint32(0) to uint8",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to uint8",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, int32(0) to uint8",
			when: int32(0),
		},
		{
			name: "ok, int32(-1) negative to 0 to uint8",
			when: int32(-1),
		},
		{
			name:   "ok, uint64 to uint8",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, uint64(0) to uint8",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to uint8",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, int64(0) to uint8",
			when: int64(0),
		},
		{
			name: "ok, int64(-1) negative to 0 to uint8",
			when: int64(-1),
		},
		{
			name:   "ok, float32 to uint8",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, float32(0) to uint8",
			when: float32(0),
		},
		{
			name: "ok, float32(-1) negative to 0 to uint8",
			when: float32(-1),
		},
		{
			name:   "ok, float64 to uint8",
			when:   math.MaxFloat64,
			expect: []byte{0x0, 0xff}, // limit to max uint8
		},
		{
			name: "ok, float64(0) to uint8",
			when: float64(0),
		},
		{
			name:   "ok, uint to uint8",
			when:   uint(math.MaxUint),
			expect: []byte{0x0, 0xff}, // this will fail on 32bit arch
		},
		{
			name:   "ok, int to uint8",
			when:   int(math.MaxInt),
			expect: []byte{0x0, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0},
			when:      uint8(1),
			expect:    []byte{0x0},
			expectErr: "field type byte or uint8 requires at least 2 bytes",
		},
		{
			name:   "ok, []byte is supported type",
			when:   []byte{0x1, 0x2, 0x3},
			expect: []byte{0x0, 0x1},
		},
		{
			name:         "ok, []byte is supported type, to high byte",
			when:         []byte{0x1, 0x2, 0x3},
			whenHighByte: true,
			expect:       []byte{0x1, 0x0},
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeUint8: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeUint8(dst, tc.when, tc.whenHighByte)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeInt16(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to int16",
			when:   true,
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, bool(false) to int16",
			when: false,
		},
		{
			name:   "ok, uint8 to int16",
			when:   uint8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint8(0) to int16",
			when: uint8(0),
		},
		{
			name:   "ok, int8 to int16",
			when:   int8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int8(0) to int16",
			when: int8(0),
		},
		{
			name:   "ok, int8(-1) negative to 0 to int16",
			when:   int8(-1),
			expect: []byte{0xff, 0xff},
		},
		{
			name:   "ok, int8(max neg) to int16",
			when:   int8(math.MinInt8),
			expect: []byte{0xff, 0x80},
		},
		{
			name:   "ok, uint16 to int16",
			when:   uint16(0xFEED),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, uint16(0) to int16",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to int16",
			when:   int16(0x7EED),
			expect: []byte{0x7e, 0xed},
		},
		{
			name: "ok, int16(0) to int16",
			when: int16(0),
		},
		{
			name:   "ok, int16(-1) negative to 0 to int16",
			when:   int16(-1),
			expect: []byte{0xff, 0xff},
		},
		{
			name:   "ok, int16(max neg) to int16",
			when:   int16(math.MinInt16),
			expect: []byte{0x80, 0x0},
		},
		{
			name:   "ok, uint32 to int16",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, uint32(0) to int16",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to int16",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, int32(0) to int16",
			when: int32(0),
		},
		{
			name:   "ok, int32(-1) negative to 0 to int16",
			when:   int32(-1),
			expect: []byte{0xff, 0xff},
		},
		{
			name:   "ok, int32(max neg) to int16",
			when:   int32(math.MinInt32),
			expect: []byte{0x80, 0x0},
		},
		{
			name:   "ok, uint64 to int16",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, uint64(0) to int16",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to int16",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, int64(0) to int16",
			when: int64(0),
		},
		{
			name:   "ok, int64(-1) negative to 0 to int16",
			when:   int64(-1),
			expect: []byte{0xff, 0xff},
		},
		{
			name:   "ok, int64(max neg) to int16",
			when:   int64(math.MinInt64),
			expect: []byte{0x80, 0x0},
		},
		{
			name:   "ok, float32 to int16",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, float32(0) to int16",
			when: float32(0),
		},
		{
			name:   "ok, float32(-1) negative to 0 to int16",
			when:   float32(-1),
			expect: []byte{0xff, 0xff},
		},
		{
			name:   "ok, float32(max neg) to int16",
			when:   float32(-math.MaxFloat32),
			expect: []byte{0x80, 0x0},
		},
		{
			name:   "ok, float64 to int16",
			when:   math.MaxFloat64,
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, float64(0) to int16",
			when: float64(0),
		},
		{
			name:   "ok, float64(max neg) to int16",
			when:   float64(-math.MaxFloat64),
			expect: []byte{0x80, 0x0},
		},
		{
			name:   "ok, uint to int16",
			when:   uint(math.MaxUint),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name:   "ok, int to int16",
			when:   int(math.MaxInt),
			expect: []byte{0x7f, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0},
			when:      uint8(1),
			expect:    []byte{0x0},
			expectErr: "field type int16 requires at least 2 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeInt16: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expectErr: "marshalFieldTypeInt16: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeInt16(dst, tc.when)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeUint16(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to uint16",
			when:   true,
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, bool(false) to uint16",
			when: false,
		},
		{
			name:   "ok, uint8 to uint16",
			when:   uint8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint8(0) to uint16",
			when: uint8(0),
		},
		{
			name:   "ok, int8 to uint16",
			when:   int8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int8(0) to uint16",
			when: int8(0),
		},
		{
			name: "ok, int8(-1) negative to 0 to uint16",
			when: int8(-1),
		},
		{
			name:   "ok, uint16 to uint16",
			when:   uint16(0xFEED),
			expect: []byte{0xfe, 0xed},
		},
		{
			name: "ok, uint16(0) to uint16",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to uint16",
			when:   int16(0x7EED),
			expect: []byte{0x7e, 0xed},
		},
		{
			name: "ok, int16(0) to uint16",
			when: int16(0),
		},
		{
			name: "ok, int16(-1) negative to 0 to uint16",
			when: int16(-1),
		},
		{
			name:   "ok, uint32 to uint16",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0xff, 0xff}, // limit to max uint16
		},
		{
			name: "ok, uint32(0) to uint16",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to uint16",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, int32(0) to uint16",
			when: int32(0),
		},
		{
			name: "ok, int32(-1) negative to 0 to uint16",
			when: int32(-1),
		},
		{
			name:   "ok, uint64 to uint16",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0xff, 0xff}, // limit to max uint16
		},
		{
			name: "ok, uint64(0) to uint16",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to uint16",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x7f, 0xff}, // limit to max uint16
		},
		{
			name: "ok, int64(0) to uint16",
			when: int64(0),
		},
		{
			name: "ok, int64(-1) negative to 0 to uint16",
			when: int64(-1),
		},
		{
			name:   "ok, float32 to uint16",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x7f, 0xff}, // limit to max int16
		},
		{
			name: "ok, float32(0) to uint16",
			when: float32(0),
		},
		{
			name: "ok, float32(-1) negative to 0 to uint16",
			when: float32(-1),
		},
		{
			name:   "ok, float64 to uint16",
			when:   math.MaxFloat64,
			expect: []byte{0x7f, 0xff}, // limit to max int32
		},
		{
			name: "ok, float64(0) to uint16",
			when: float64(0),
		},
		{
			name:   "ok, uint to uint16",
			when:   uint(math.MaxUint),
			expect: []byte{0xff, 0xff}, // this will fail on 32bit arch
		},
		{
			name:   "ok, int to uint16",
			when:   int(math.MaxInt),
			expect: []byte{0x7f, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0},
			when:      uint8(1),
			expect:    []byte{0x0},
			expectErr: "field type uint16 requires at least 2 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeUint16: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expectErr: "marshalFieldTypeUint16: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeUint16(dst, tc.when)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeInt32(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to int32",
			when:   true,
			expect: []byte{0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, bool(false) to int32",
			when:   false,
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8 to int32",
			when:   uint8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, uint8(0) to int32",
			when:   uint8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8 to int32",
			when:   int8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x1},
		},
		{
			name: "ok, int8(0) to int32",
			when: int8(0),
		},
		{
			name:   "ok, int8(-1) negative to 0 to int32",
			when:   int8(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, int8(max neg) to int32",
			when:   int8(math.MinInt8),
			expect: []byte{0xff, 0xff, 0xff, 0x80},
		},
		{
			name:   "ok, uint16 to int32",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0x0, 0xfe, 0xed},
		},
		{
			name: "ok, uint16(0) to int32",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to int32",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0x0, 0x7e, 0xed},
		},
		{
			name: "ok, int16(0) to int32",
			when: int16(0),
		},
		{
			name:   "ok, int16(-1) negative to 0 to int32",
			when:   int16(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, int16(max neg) to int32",
			when:   int16(math.MinInt16),
			expect: []byte{0xff, 0xff, 0x80, 0x0},
		},
		{
			name:   "ok, uint32 to int32",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name: "ok, uint32(0) to int32",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to int32",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x7e, 0xed, 0xfe, 0xed},
		},
		{
			name: "ok, int32(0) to int32",
			when: int32(0),
		},
		{
			name:   "ok, int32(-1) negative to 0 to int32",
			when:   int32(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, int32(max neg) to int32",
			when:   int32(math.MinInt32),
			expect: []byte{0x80, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint64 to int32",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name: "ok, uint64(0) to int32",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to int32",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name: "ok, int64(0) to int32",
			when: int64(0),
		},
		{
			name:   "ok, int64(-1) negative to 0 to int32",
			when:   int64(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, int64(max neg) to int32",
			when:   int64(math.MinInt64),
			expect: []byte{0x80, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32 to int32",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x80, 0x0, 0x0, 0x0}, // limit to max int32
		},
		{
			name: "ok, float32(0) to int32",
			when: float32(0),
		},
		{
			name:   "ok, float32(-1) negative to 0 to int32",
			when:   float32(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, float32(max neg) to int32",
			when:   float32(-math.MaxFloat32),
			expect: []byte{0x80, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float64 to int32",
			when:   math.MaxFloat64,
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name: "ok, float64(0) to int32",
			when: float64(0),
		},
		{
			name:   "ok, float64(max neg) to int32",
			when:   float64(-math.MaxFloat64),
			expect: []byte{0x80, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint to int32",
			when:   uint(math.MaxUint),
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name:   "ok, int to int32",
			when:   int(math.MaxInt),
			expect: []byte{0x7f, 0xff, 0xff, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0},
			expectErr: "field type int32 requires at least 4 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeInt32: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expectErr: "marshalFieldTypeInt32: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0, 0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0, 0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeInt32(dst, tc.when)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeUint32(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to uint32",
			when:   true,
			expect: []byte{0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, bool(false) to uint32",
			when:   false,
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8 to uint32",
			when:   uint8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, uint8(0) to uint32",
			when:   uint8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8 to uint32",
			when:   int8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, int8(0) to uint32",
			when:   int8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8(-1) negative to 0 to uint32",
			when:   int8(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint16 to uint32",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0x0, 0xfe, 0xed},
		},
		{
			name:   "ok, uint16(0) to uint32",
			when:   uint16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16 to uint32",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0x0, 0x7e, 0xed},
		},
		{
			name:   "ok, int16(0) to uint32",
			when:   int16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16(-1) negative to 0 to uint32",
			when:   int16(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint32 to uint32",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0xfe, 0xed, 0xfe, 0xed},
		},
		{
			name:   "ok, uint32(0) to uint32",
			when:   uint32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32 to uint32",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x7e, 0xed, 0xfe, 0xed},
		},
		{
			name:   "ok, int32(0) to uint32",
			when:   int32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32(-1) negative to 0 to uint32",
			when:   int32(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint64 to uint32",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0xff, 0xff, 0xff, 0xff}, // limit to max uint32
		},
		{
			name:   "ok, uint64(0) to uint32",
			when:   uint64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64 to uint32",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name:   "ok, int64(0) to uint32",
			when:   int64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(-1) negative to 0 to uint32",
			when:   int64(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32 to uint32",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x80, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(0) to uint32",
			when:   float32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(-1) negative to 0 to uint32",
			when:   float32(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float64 to uint32",
			when:   math.MaxFloat64,
			expect: []byte{0x7f, 0xff, 0xff, 0xff}, // limit to max int32
		},
		{
			name:   "ok, float64(0) to uint32",
			when:   float64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint to uint32",
			when:   uint(math.MaxUint),
			expect: []byte{0xff, 0xff, 0xff, 0xff}, // this will fail on 32bit arch
		},
		{
			name:   "ok, int to uint32",
			when:   int(math.MaxInt),
			expect: []byte{0x7f, 0xff, 0xff, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0},
			expectErr: "field type byte or uint32 requires at least 4 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expect:    []byte{0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeUint32: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expect:    []byte{0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeUint32: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := []byte{0x0, 0x0, 0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeUint32(dst, tc.when)

			assert.Equal(t, tc.expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeInt64(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to int64",
			when:   true,
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, bool(false) to int64",
			when:   false,
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8 to int64",
			when:   uint8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, uint8(0) to int64",
			when:   uint8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8 to int64",
			when:   int8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, int8(0) to int64",
			when:   int8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8(-1) negative to 0 to int64",
			when:   int8(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, uint16 to int64",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xed},
		},
		{
			name:   "ok, uint16(0) to int64",
			when:   uint16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16 to int64",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7e, 0xed},
		},
		{
			name:   "ok, int16(0) to int64",
			when:   int16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16(-1) negative to 0 to int64",
			when:   int16(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, uint32 to int64",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0xfe, 0xed, 0xfe, 0xed},
		},
		{
			name:   "ok, uint32(0) to int64",
			when:   uint32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32 to int64",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x7e, 0xed, 0xfe, 0xed},
		},
		{
			name:   "ok, int32(0) to int64",
			when:   int32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32(-1) negative to 0 to int64",
			when:   int32(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, uint64 to int64",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		},
		{
			name:   "ok, uint64(0) to int64",
			when:   uint64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64 to int64",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x71, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		},
		{
			name:   "ok, int64(0) to int64",
			when:   int64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(-1) negative to 0 to int64",
			when:   int64(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, float32 to int64",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(0) to int64",
			when:   float32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(-1) negative to 0 to int64",
			when:   float32(-1),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, float64 to int64",
			when:   math.MaxFloat64,
			expect: []byte{0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float64(0) to int64",
			when:   float64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint to int64",
			when:   uint(math.MaxUint),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // this will fail on 32bit arch
		},
		{
			name:   "ok, int to int64",
			when:   int(math.MaxInt),
			expect: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "field type int64 requires at least 8 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeInt64: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeInt64: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeInt64(dst, tc.when)

			assert.Equal(t, tc.expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeUint64(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to uint64",
			when:   true,
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, bool(false) to uint64",
			when:   false,
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8 to uint64",
			when:   uint8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, uint8(0) to uint64",
			when:   uint8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8 to uint64",
			when:   int8(1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		},
		{
			name:   "ok, int8(0) to uint64",
			when:   int8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8(-1) negative to 0 to uint64",
			when:   int8(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint16 to uint64",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xed},
		},
		{
			name:   "ok, uint16(0) to uint64",
			when:   uint16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16 to uint64",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7e, 0xed},
		},
		{
			name:   "ok, int16(0) to uint64",
			when:   int16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16(-1) negative to 0 to uint64",
			when:   int16(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint32 to uint64",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0xfe, 0xed, 0xfe, 0xed},
		},
		{
			name:   "ok, uint32(0) to uint64",
			when:   uint32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32 to uint64",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x7e, 0xed, 0xfe, 0xed},
		},
		{
			name:   "ok, int32(0) to uint64",
			when:   int32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32(-1) negative to 0 to uint64",
			when:   int32(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint64 to uint64",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		},
		{
			name:   "ok, uint64(0) to uint64",
			when:   uint64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64 to uint64",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x71, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		},
		{
			name:   "ok, int64(0) to uint64",
			when:   int64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(-1) negative to 0 to uint64",
			when:   int64(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32 to uint64",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(0) to uint64",
			when:   float32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(-1) negative to 0 to uint64",
			when:   float32(-1),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float64 to uint64",
			when:   math.MaxFloat64,
			expect: []byte{0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float64(0) to uint64",
			when:   float64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint to uint64",
			when:   uint(math.MaxUint),
			expect: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // this will fail on 32bit arch
		},
		{
			name:   "ok, int to uint64",
			when:   int(math.MaxInt),
			expect: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "field type byte or uint64 requires at least 8 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeUint64: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeUint64: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeUint64(dst, tc.when)

			assert.Equal(t, tc.expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeFloat32(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to float32",
			when:   true,
			expect: []byte{0x3f, 0x80, 0x0, 0x0},
		},
		{
			name:   "ok, bool(false) to float32",
			when:   false,
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8 to float32",
			when:   uint8(1),
			expect: []byte{0x3f, 0x80, 0x0, 0x0},
		},
		{
			name:   "ok, uint8(0) to float32",
			when:   uint8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8 to float32",
			when:   int8(1),
			expect: []byte{0x3f, 0x80, 0x0, 0x0},
		},
		{
			name: "ok, int8(0) to float32",
			when: int8(0),
		},
		{
			name:   "ok, int8(-1) negative to 0 to float32",
			when:   int8(-1),
			expect: []byte{0xbf, 0x80, 0x0, 0x0},
		},
		{
			name:   "ok, int8(max neg) to float32",
			when:   int8(math.MinInt8),
			expect: []byte{0xc3, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint16 to float32",
			when:   uint16(0xFEED),
			expect: []byte{0x47, 0x7e, 0xed, 0x0},
		},
		{
			name: "ok, uint16(0) to float32",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to float32",
			when:   int16(0x7EED),
			expect: []byte{0x46, 0xfd, 0xda, 0x0},
		},
		{
			name: "ok, int16(0) to float32",
			when: int16(0),
		},
		{
			name:   "ok, int16(-1) negative to 0 to float32",
			when:   int16(-1),
			expect: []byte{0xbf, 0x80, 0x0, 0x0},
		},
		{
			name:   "ok, int16(max neg) to float32",
			when:   int16(math.MinInt16),
			expect: []byte{0xc7, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint32 to float32",
			when:   uint32(math.MaxUint32),
			expect: []byte{0x4f, 0x80, 0x0, 0x0},
		},
		{
			name: "ok, uint32(0) to float32",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to float32",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x4e, 0xfd, 0xdb, 0xfe},
		},
		{
			name: "ok, int32(0) to float32",
			when: int32(0),
		},
		{
			name:   "ok, int32(-1) negative to 0 to float32",
			when:   int32(-1),
			expect: []byte{0xbf, 0x80, 0x0, 0x0},
		},
		{
			name:   "ok, int32(max neg) to float32",
			when:   int32(math.MinInt32),
			expect: []byte{0xcf, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint64 to float32",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x5b, 0x81, 0x1, 0x82},
		},
		{
			name: "ok, uint64(0) to float32",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to float32",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0xdf, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(0) to float32",
			when:   int64(0),
			expect: []byte{0xdf, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(-1) negative to 0 to float32",
			when:   int64(-1),
			expect: []byte{0xdf, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(max neg) to float32",
			when:   int64(math.MinInt64),
			expect: []byte{0xdf, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32 to float32",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x7f, 0x7f, 0xff, 0xff},
		},
		{
			name: "ok, float32(0) to float32",
			when: float32(0),
		},
		{
			name:   "ok, float32(-1) negative to 0 to float32",
			when:   float32(-1),
			expect: []byte{0xbf, 0x80, 0x0, 0x0},
		},
		{
			name:   "ok, float32(max neg) to float32",
			when:   float32(-math.MaxFloat32),
			expect: []byte{0xff, 0x7f, 0xff, 0xff},
		},
		{
			name:   "ok, float64 to float32",
			when:   math.MaxFloat64,
			expect: []byte{0x7f, 0x7f, 0xff, 0xff},
		},
		{
			name: "ok, float64(0) to float32",
			when: float64(0),
		},
		{
			name:   "ok, float64(max neg) to float32",
			when:   float64(-math.MaxFloat64),
			expect: []byte{0xff, 0x7f, 0xff, 0xff},
		},
		{
			name:   "ok, uint to float32",
			when:   uint(math.MaxUint),
			expect: []byte{0x5f, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int to float32",
			when:   int(math.MaxInt),
			expect: []byte{0xdf, 0x0, 0x0, 0x0},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0},
			expectErr: "field type float32 requires at least 4 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeFloat32: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expectErr: "marshalFieldTypeFloat32: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0, 0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0, 0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeFloat32(dst, tc.when)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeFloat64(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, bool(true) to float64",
			when:   true,
			expect: []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, bool(false) to float64",
			when:   false,
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8 to float64",
			when:   uint8(1),
			expect: []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint8(0) to float64",
			when:   uint8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8 to float64",
			when:   int8(1),
			expect: []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8(0) to float64",
			when:   int8(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int8(-1) negative to 0 to float64",
			when:   int8(-1),
			expect: []byte{0xbf, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint16 to float64",
			when:   uint16(0xFEED),
			expect: []byte{0x40, 0xef, 0xdd, 0xa0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint16(0) to float64",
			when:   uint16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16 to float64",
			when:   int16(0x7EED),
			expect: []byte{0x40, 0xdf, 0xbb, 0x40, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16(0) to float64",
			when:   int16(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int16(-1) negative to 0 to float64",
			when:   int16(-1),
			expect: []byte{0xbf, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint32 to float64",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x41, 0xef, 0xdd, 0xbf, 0xdd, 0xa0, 0x0, 0x0},
		},
		{
			name:   "ok, uint32(0) to float64",
			when:   uint32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32 to float64",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x41, 0xdf, 0xbb, 0x7f, 0xbb, 0x40, 0x0, 0x0},
		},
		{
			name:   "ok, int32(0) to float64",
			when:   int32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int32(-1) negative to 0 to float64",
			when:   int32(-1),
			expect: []byte{0xbf, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint64 to float64",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x43, 0x70, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70},
		},
		{
			name:   "ok, uint64(0) to float64",
			when:   uint64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64 to float64",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0xc3, 0xe0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(0) to float64",
			when:   int64(0),
			expect: []byte{0xc3, 0xe0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, int64(-1) negative to 0 to float64",
			when:   int64(-1),
			expect: []byte{0xc3, 0xe0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32 to float64",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x47, 0xef, 0xff, 0xff, 0xe0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(0) to float64",
			when:   float32(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float32(-1) negative to 0 to float64",
			when:   float32(-1),
			expect: []byte{0xbf, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, float64 to float64",
			when:   math.MaxFloat64,
			expect: []byte{0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:   "ok, float64(0) to float64",
			when:   float64(0),
			expect: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, uint to float64",
			when:   uint(math.MaxUint),
			expect: []byte{0x43, 0xe0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, // this will fail on 32bit arch
		},
		{
			name:   "ok, int to float64",
			when:   int(math.MaxInt),
			expect: []byte{0xc3, 0xe0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "field type float64 requires at least 8 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeFloat64: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x0},
			expect:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			expectErr: "marshalFieldTypeFloat64: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeFloat64(dst, tc.when)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalStringOrByteRegisters(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, string, smaller than dst",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   "ABC",
			expect: []byte{0x0, 0x43, 0x42, 0x41},
		},
		{
			name:   "ok, string, longer than dst",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   "ABCDE",
			expect: []byte{0x44, 0x43, 0x42, 0x41},
		},
		{
			name:   "ok, string, equal size",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   "ABCD",
			expect: []byte{0x44, 0x43, 0x42, 0x41},
		},
		{
			name:   "ok, byte slice, smaller than dst",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   []byte("ABC"),
			expect: []byte{0x0, 0x43, 0x42, 0x41},
		},
		{
			name:   "ok, byte slice, longer than dst",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   []byte("ABCDE"),
			expect: []byte{0x44, 0x43, 0x42, 0x41},
		},
		{
			name:   "ok, byte slice, equal size",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   []byte("ABCD"),
			expect: []byte{0x44, 0x43, 0x42, 0x41},
		},
		{
			name:   "ok, string empty",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   "",
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:   "ok, []byte empty",
			given:  []byte{0x0, 0x0, 0x0, 0x0},
			when:   []byte{},
			expect: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:      "nok, unknown type",
			given:     []byte{0x0, 0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0, 0x0},
			expectErr: "can not marshal number type to field with string or bytes type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := marshalFieldTypeStringOrBytes(tc.given, tc.when)

			assert.Equal(t, tc.expect, tc.given)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarshalFieldTypeBit(t *testing.T) {
	var testCases = []struct {
		name      string
		given     []byte
		when      any
		whenBit   uint8
		expect    []byte
		expectErr string
	}{
		{
			name:    "ok, bool(true), low byte to bit",
			when:    true,
			whenBit: 1, // low byte
			expect:  []byte{0x0, 0x2},
		},
		{
			name:    "ok, bool(true), high byte to bit",
			when:    true,
			whenBit: 8, // high byte
			expect:  []byte{0x1, 0x0},
		},
		{
			name: "ok, bool(false) to bit",
			when: false,
		},
		{
			name:   "ok, uint8 to bit",
			when:   uint8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint8(0) to bit",
			when: uint8(0),
		},
		{
			name:   "ok, int8 to bit",
			when:   int8(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int8(0) to bit",
			when: int8(0),
		},
		{
			name:   "ok, int8(-1) negative to 0 to bit",
			when:   int8(-1),
			expect: []byte{0x0, 0x1},
		},
		{
			name:   "ok, uint16 to bit",
			when:   uint16(0xFEED),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint16(0) to bit",
			when: uint16(0),
		},
		{
			name:   "ok, int16 to bit",
			when:   int16(0x7EED),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int16(0) to bit",
			when: int16(0),
		},
		{
			name:   "ok, int16(-1) negative to 0 to bit",
			when:   int16(-1),
			expect: []byte{0x0, 0x1},
		},
		{
			name:   "ok, uint32 to bit",
			when:   uint32(0xFEEDFEED),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint32(0) to bit",
			when: uint32(0),
		},
		{
			name:   "ok, int32 to bit",
			when:   int32(0x7EEDFEED),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int32(0) to bit",
			when: int32(0),
		},
		{
			name:   "ok, int32(-1) negative to 0 to bit",
			when:   int32(-1),
			expect: []byte{0x0, 0x1},
		},
		{
			name:   "ok, uint64 to bit",
			when:   uint64(0x0102_0304_0506_0708),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint64(0) to bit",
			when: uint64(0),
		},
		{
			name:   "ok, int64 to bit",
			when:   int64(0x7102_0304_0506_0708),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int64(0) to bit",
			when: int64(0),
		},
		{
			name:   "ok, int64(-1) negative to 0 to bit",
			when:   int64(-1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, int(0) to bit",
			when: int(0),
		},
		{
			name:   "ok, int(-1) negative to 0 to bit",
			when:   int(-1),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, uint(0) to bit",
			when: uint(0),
		},
		{
			name:   "ok, uint(1) to bit",
			when:   uint(1),
			expect: []byte{0x0, 0x1},
		},
		{
			name:   "ok, float32 to bit",
			when:   float32(math.MaxFloat32),
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, float32(0) to bit",
			when: float32(0),
		},
		{
			name:   "ok, float32(-1) negative to 0 to bit",
			when:   float32(-1),
			expect: []byte{0x0, 0x1},
		},
		{
			name:   "ok, float64 to bit",
			when:   math.MaxFloat64,
			expect: []byte{0x0, 0x1},
		},
		{
			name: "ok, float64(0) to bit",
			when: float64(0),
		},
		{
			name:   "ok, uint to bit",
			when:   uint(math.MaxUint),
			expect: []byte{0x0, 0x1},
		},
		{
			name:   "ok, int to bit",
			when:   int(math.MaxInt),
			expect: []byte{0x0, 0x1},
		},
		{
			name:      "nok, too short dst",
			given:     []byte{0x0},
			when:      uint8(1),
			expect:    []byte{0x0},
			expectErr: "field type bit requires at least 2 bytes",
		},
		{
			name:      "nok, string is unsupported type",
			when:      "nope",
			expectErr: "marshalFieldTypeFloat64: can not marshal unsupported type",
		},
		{
			name:      "nok, []byte is unsupported type",
			when:      []byte{0x1, 0x2, 0x3},
			expectErr: "marshalFieldTypeFloat64: can not marshal unsupported type",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := []byte{0x0, 0x0}
			if tc.expect != nil {
				expect = tc.expect
			}
			dst := []byte{0x0, 0x0}
			if tc.given != nil {
				dst = tc.given
			}
			err := marshalFieldTypeBit(dst, tc.when, tc.whenBit)

			assert.Equal(t, expect, dst)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistersToLowWordFirst(t *testing.T) {
	var testCases = []struct {
		name      string
		when      []byte
		expect    []byte
		expectErr string
	}{
		{
			name:   "ok, 1 register, do nothing",
			when:   []byte{0x1, 0x2},
			expect: []byte{0x1, 0x2},
		},
		{
			name:   "ok, 2 registers",
			when:   []byte{0x44, 0x43, 0x42, 0x41},
			expect: []byte{0x42, 0x41, 0x44, 0x43},
		},
		{
			name:   "ok, 3 registers",
			when:   []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
			expect: []byte{0x5, 0x6, 0x3, 0x4, 0x1, 0x2},
		},
		{
			name:   "ok, 4 registers",
			when:   []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
			expect: []byte{0x7, 0x8, 0x5, 0x6, 0x3, 0x4, 0x1, 0x2},
		},
		{
			name:      "ok, size is odd number of bytes for target",
			when:      []byte{0x1, 0x2, 0x3, 0x4, 0x5},
			expect:    []byte{0x1, 0x2, 0x3, 0x4, 0x5},
			expectErr: "registersToLowWordFirst: target size must be even bytes",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := registersToLowWordFirst(tc.when)

			assert.Equal(t, tc.expect, tc.when)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
