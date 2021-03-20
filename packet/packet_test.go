package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCRC16(t *testing.T) {
	var testCases = []struct {
		name   string
		when   []byte
		expect uint16
	}{
		{
			name:   "ok",
			when:   []byte{0x01, 0x04, 0x02, 0xFF, 0xFF},
			expect: 0x80B8,
		},
		{
			name:   "ok",
			when:   []byte{0x01, 0x04, 0x02, 0xFF, 0xFF},
			expect: 0x80B8,
		},
		{
			name:   "ok",
			when:   []byte{0x11, 0x03, 0x00, 0x6B, 0x00, 0x03},
			expect: 0x8776,
		},
		{
			name:   "ok2",
			when:   []byte{0x03, 0x03, 0x02, 0xCD, 0x6B},
			expect: 0xFBD4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CRC16(tc.when)
			assert.Equal(t, tc.expect, result)
		})
	}
}
