package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadCoilsResponseTCP_Bytes(t *testing.T) {
	example := ReadCoilsResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadCoilsResponse: ReadCoilsResponse{
			UnitID: 1,
			// +1 function code
			CoilsByteLength: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadCoilsResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadCoilsResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadCoilsResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.CoilsByteLength = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x5, 0x10, 0x1, 0x2, 0x1, 0x2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			given := example
			tc.given(&given)

			assert.Equal(t, tc.expect, given.Bytes())
		})
	}
}

func TestParseReadCoilsResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadCoilsResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x01, 0x02, 0xCD, 0x6B},
			expect: &ReadCoilsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadCoilsResponse: ReadCoilsResponse{
					UnitID:          3,
					CoilsByteLength: 2,
					Data:            []byte{0xcd, 0x6b},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x01, 0x02},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x01, 0x01, 0xCD, 0x6B},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadCoilsResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadCoilsResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadCoilsResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xec, 0xd2},
			expect: &ReadCoilsResponseRTU{
				ReadCoilsResponse: ReadCoilsResponse{
					UnitID:          16,
					CoilsByteLength: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0x1, 0x2, 0xec, 0xd2},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x1, 0x1, 0x1, 0x2, 0xec, 0xd2},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadCoilsResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadCoilsResponseRTU_Bytes(t *testing.T) {
	example := ReadCoilsResponseRTU{
		ReadCoilsResponse: ReadCoilsResponse{
			UnitID: 1,
			// +1 function code
			CoilsByteLength: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadCoilsResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadCoilsResponseRTU) {},
			expect: []byte{0x1, 0x1, 0x2, 0x0, 0x1, 0x78, 0x3c},
		},
		{
			name: "ok2",
			given: func(r *ReadCoilsResponseRTU) {
				r.UnitID = 16
				r.CoilsByteLength = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			given := example
			tc.given(&given)

			assert.Equal(t, tc.expect, given.Bytes())
		})
	}
}

func TestReadCoilsResponse_FunctionCode(t *testing.T) {
	given := ReadCoilsResponse{}
	assert.Equal(t, uint8(1), given.FunctionCode())
}

func TestReadCoilsResponse_Bytes(t *testing.T) {
	example := ReadCoilsResponse{
		UnitID: 1,
		// +1 function code
		CoilsByteLength: 2,
		Data:            []byte{0x0, 0x1},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadCoilsResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadCoilsResponse) {},
			expect: []byte{0x1, 0x1, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadCoilsResponse) {
				r.UnitID = 16
				r.CoilsByteLength = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x1, 0x2, 0x1, 0x2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			given := example
			tc.given(&given)

			assert.Equal(t, tc.expect, given.Bytes())
		})
	}
}

func TestReadCoilsResponse_IsCoilSet(t *testing.T) {
	var testCases = []struct {
		name             string
		whenStartAddress uint16
		whenCoilAddress  uint16
		expect           bool
		expectError      string
	}{
		{
			name:             "ok, first byte, second bit",
			whenStartAddress: 0, whenCoilAddress: 1, expect: true,
		},
		{
			name:             "ok, first byte, second bit, start 1",
			whenStartAddress: 1, whenCoilAddress: 2, expect: true},
		{
			name:             "ok, first byte, second bit, start 100",
			whenStartAddress: 100, whenCoilAddress: 101, expect: true,
		},
		{
			name:             "ok, first byte, third bit",
			whenStartAddress: 0, whenCoilAddress: 2, expect: false,
		},
		{
			name:             "ok, second byte, first bit",
			whenStartAddress: 0, whenCoilAddress: 8, expect: true,
		},
		{
			name:             "ok, second byte, last bit",
			whenStartAddress: 0, whenCoilAddress: 15, expect: true,
		},
		{
			name:             "ok, bit out of bounds",
			whenStartAddress: 0, whenCoilAddress: 16, expect: false,
			expectError: "bit value more than data contains bits",
		},
		{
			name:             "ok, bit before start coil",
			whenStartAddress: 10, whenCoilAddress: 9, expect: false,
			expectError: "bit can not be before startBit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			given := ReadCoilsResponse{
				CoilsByteLength: 2,
				Data:            []byte{0b10000001, 0b00010010},
			}
			result, err := given.IsCoilSet(tc.whenStartAddress, tc.whenCoilAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
