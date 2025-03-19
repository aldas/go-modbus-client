package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorResponseTCP_Error(t *testing.T) {
	var testCases = []struct {
		given  ErrorResponseTCP
		expect string
	}{
		{
			given:  ErrorResponseTCP{Code: 1},
			expect: "Illegal function",
		},
		{
			given:  ErrorResponseTCP{Code: 2},
			expect: "Illegal data address",
		},
		{
			given:  ErrorResponseTCP{Code: 3},
			expect: "Illegal data value",
		},
		{
			given:  ErrorResponseTCP{Code: 4},
			expect: "Server failure",
		},
		{
			given:  ErrorResponseTCP{Code: 5},
			expect: "Acknowledge",
		},
		{
			given:  ErrorResponseTCP{Code: 6},
			expect: "Server busy",
		},
		{
			given:  ErrorResponseTCP{Code: 8},
			expect: "Memory parity error",
		},
		{
			given:  ErrorResponseTCP{Code: 10},
			expect: "Gateway path unavailable",
		},
		{
			given:  ErrorResponseTCP{Code: 11},
			expect: "Gateway targeted device failed to respond",
		},
		{
			given:  ErrorResponseTCP{Code: 12},
			expect: "Unknown error code: 12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			assert.EqualError(t, tc.given, tc.expect)
		})
	}
}

func TestErrorResponseRTU_Error(t *testing.T) {
	var testCases = []struct {
		given  ErrorResponseRTU
		expect string
	}{
		{
			given:  ErrorResponseRTU{Code: 1},
			expect: "Illegal function",
		},
		{
			given:  ErrorResponseRTU{Code: 2},
			expect: "Illegal data address",
		},
		{
			given:  ErrorResponseRTU{Code: 3},
			expect: "Illegal data value",
		},
		{
			given:  ErrorResponseRTU{Code: 4},
			expect: "Server failure",
		},
		{
			given:  ErrorResponseRTU{Code: 5},
			expect: "Acknowledge",
		},
		{
			given:  ErrorResponseRTU{Code: 6},
			expect: "Server busy",
		},
		{
			given:  ErrorResponseRTU{Code: 8},
			expect: "Memory parity error",
		},
		{
			given:  ErrorResponseRTU{Code: 10},
			expect: "Gateway path unavailable",
		},
		{
			given:  ErrorResponseRTU{Code: 11},
			expect: "Gateway targeted device failed to respond",
		},
		{
			given:  ErrorResponseRTU{Code: 12},
			expect: "Unknown error code: 12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			assert.EqualError(t, tc.given, tc.expect)
		})
	}
}

func TestErrorResponseTCP_FunctionCode(t *testing.T) {
	given := ErrorResponseTCP{Function: 1}
	assert.Equal(t, uint8(1), given.FunctionCode())
}

func TestErrorResponseRTU_FunctionCode(t *testing.T) {
	given := ErrorResponseRTU{Function: 1}
	assert.Equal(t, uint8(1), given.FunctionCode())
}

func TestErrorResponseTCP_ModbusErrorCode(t *testing.T) {
	given := ErrorResponseTCP{Function: 1, Code: ErrIllegalDataAddress}
	assert.Equal(t, ErrIllegalDataAddress, given.ErrorCode())
}

func TestErrorResponseRTU_ModbusErrorCode(t *testing.T) {
	given := ErrorResponseRTU{Function: 1, Code: ErrIllegalDataAddress}
	assert.Equal(t, ErrIllegalDataAddress, given.ErrorCode())
}

func TestErrorResponseTCP_Bytes(t *testing.T) {
	var testCases = []struct {
		name   string
		given  ErrorResponseTCP
		expect []byte
	}{
		{
			name:   "ok",
			given:  ErrorResponseTCP{TransactionID: 1245, UnitID: 1, Function: 2, Code: 3},
			expect: []byte{0x4, 0xdd, 0x0, 0x0, 0x0, 0x3, 0x1, 0x82, 0x3},
		},
		{
			name:   "ok2",
			given:  ErrorResponseTCP{TransactionID: 55943, UnitID: 0, Function: 1, Code: 3},
			expect: []byte{0xda, 0x87, 0x00, 0x00, 0x00, 0x03, 0x00, 0x81, 0x03},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.given.Bytes()
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestErrorResponseRTU_Bytes(t *testing.T) {
	var testCases = []struct {
		name   string
		given  ErrorResponseRTU
		expect []byte
	}{
		{
			name:   "ok",
			given:  ErrorResponseRTU{UnitID: 1, Function: 2, Code: 3},
			expect: []byte{0x1, 0x82, 0x3, 0x0, 0xa1},
		},
		{
			name:   "ok2",
			given:  ErrorResponseRTU{UnitID: 1, Function: 1, Code: 1},
			expect: []byte{0x01, 0x81, 0x01, 0x81, 0x90},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.given.Bytes()
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestAsTCPErrorPacket(t *testing.T) {
	var testCases = []struct {
		name   string
		when   []byte
		expect *ErrorResponseTCP
	}{
		{
			name:   "ok",
			when:   []byte{0x4, 0xdd, 0x0, 0x0, 0x0, 0x3, 0x1, 0x82, 0x3},
			expect: &ErrorResponseTCP{TransactionID: 1245, UnitID: 1, Function: 2, Code: 3},
		},
		{
			name:   "ok2",
			when:   []byte{0xda, 0x87, 0x00, 0x00, 0x00, 0x03, 0x00, 0x81, 0x03},
			expect: &ErrorResponseTCP{TransactionID: 55943, UnitID: 0, Function: 1, Code: 3},
		},
		{
			name:   "ok, too short to be error response",
			when:   []byte{0xda, 0x87, 0x00, 0x00, 0x00, 0x03, 0x00},
			expect: nil,
		},
		{
			name:   "ok, function code is not high enough",
			when:   []byte{0xda, 0x87, 0x00, 0x00, 0x00, 0x03, 0x00, 0x01, 0x03},
			expect: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := AsTCPErrorPacket(tc.when)

			if tc.expect == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expect, result)
			}
		})
	}
}

func TestAsRTUErrorPacket(t *testing.T) {
	var testCases = []struct {
		name   string
		when   []byte
		expect *ErrorResponseRTU
	}{
		{
			name:   "ok",
			when:   []byte{0x1, 0x82, 0x3, 0xa1, 0x0},
			expect: &ErrorResponseRTU{UnitID: 1, Function: 2, Code: 3},
		},
		{
			name:   "ok2",
			when:   []byte{0x01, 0x81, 0x01, 0x90, 0x81},
			expect: &ErrorResponseRTU{UnitID: 1, Function: 1, Code: 1},
		},
		{
			name:   "ok, too short to be error response",
			when:   []byte{0x01, 0x81, 0x01, 0x90},
			expect: nil,
		},
		{
			name:   "ok, function code is not high enough",
			when:   []byte{0x01, 0x01, 0x01, 0x01, 0x81},
			expect: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := AsRTUErrorPacket(tc.when)

			if tc.expect == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expect, result)
			}
		})
	}
}
