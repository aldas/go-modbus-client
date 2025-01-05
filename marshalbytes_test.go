package modbus

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
			name:      "nok, unknown type",
			given:     []byte{0x0, 0x0, 0x0, 0x0},
			when:      uint8(1),
			expect:    []byte{0x0, 0x0, 0x0, 0x0},
			expectErr: "unknown type given to marshalStringOrByteRegisters",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := marshalStringOrByteRegisters(tc.given, tc.when)

			assert.Equal(t, tc.expect, tc.given)
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
