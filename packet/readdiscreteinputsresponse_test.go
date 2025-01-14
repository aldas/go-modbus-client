package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadDiscreteInputsResponseTCP_Bytes(t *testing.T) {
	example := ReadDiscreteInputsResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
			UnitID: 1,
			// +1 function code
			InputsByteLength: 2,
			Data:             []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadDiscreteInputsResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadDiscreteInputsResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x2, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadDiscreteInputsResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.InputsByteLength = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x5, 0x10, 0x2, 0x2, 0x1, 0x2},
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

func TestParseReadDiscreteInputsResponseTCP(t *testing.T) {
	max124registers := make([]byte, 248)

	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadDiscreteInputsResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x01, 0x02, 0xCD, 0x6B},
			expect: &ReadDiscreteInputsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
					UnitID:           3,
					InputsByteLength: 2,
					Data:             []byte{0xcd, 0x6b},
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
		{
			name:  "ok, length is at the edge max byte/uint8 value",
			given: append([]byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x02, 248}, max124registers...),
			expect: &ReadDiscreteInputsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
					UnitID:           3,
					InputsByteLength: 248,
					Data:             max124registers,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadDiscreteInputsResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadDiscreteInputsResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadDiscreteInputsResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x1, 0x2, 0x2, 0x1, 0x2, 0x22, 0x22},
			expect: &ReadDiscreteInputsResponseRTU{
				ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
					UnitID:           1,
					InputsByteLength: 2,
					Data:             []byte{0x1, 0x2},
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
		{
			name: "ok, length is at the edge max byte/uint8 value",
			given: func() []byte {
				max124registers := make([]byte, 248)
				b := append([]byte{0x03, 0x02, 248}, max124registers...)
				return append(b, []byte{0xff, 0xff}...) // + CRC (invalid crc)
			}(),
			expect: &ReadDiscreteInputsResponseRTU{
				ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
					UnitID:           3,
					InputsByteLength: 248,
					Data:             make([]byte, 248),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadDiscreteInputsResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadDiscreteInputsResponseRTU_Bytes(t *testing.T) {
	example := ReadDiscreteInputsResponseRTU{
		ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
			UnitID: 1,
			// +1 function code
			InputsByteLength: 2,
			Data:             []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadDiscreteInputsResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadDiscreteInputsResponseRTU) {},
			expect: []byte{0x1, 0x2, 0x2, 0x0, 0x1, 0x78, 0x78},
		},
		{
			name: "ok2",
			given: func(r *ReadDiscreteInputsResponseRTU) {
				r.UnitID = 16
				r.InputsByteLength = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x2, 0x2, 0x1, 0x2, 0xc5, 0xea},
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

func TestReadDiscreteInputsResponse_FunctionCode(t *testing.T) {
	given := ReadDiscreteInputsResponse{}
	assert.Equal(t, uint8(2), given.FunctionCode())
}

func TestReadDiscreteInputsResponse_Bytes(t *testing.T) {
	example := ReadDiscreteInputsResponse{
		UnitID: 1,
		// +1 function code
		InputsByteLength: 2,
		Data:             []byte{0x0, 0x1},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadDiscreteInputsResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadDiscreteInputsResponse) {},
			expect: []byte{0x1, 0x2, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadDiscreteInputsResponse) {
				r.UnitID = 16
				r.InputsByteLength = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x2, 0x2, 0x1, 0x2},
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

func TestReadDiscreteInputsResponse_IsInputSet(t *testing.T) {
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
			given := ReadDiscreteInputsResponse{
				InputsByteLength: 2,
				Data:             []byte{0b00010010, 0b10000001},
			}
			result, err := given.IsInputSet(tc.whenStartAddress, tc.whenCoilAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadDiscreteInputsResponse_IsCoilSet(t *testing.T) {
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
			given := ReadDiscreteInputsResponse{
				InputsByteLength: 2,
				Data:             []byte{0b00010010, 0b10000001},
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
