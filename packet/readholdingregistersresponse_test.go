package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadHoldingRegistersResponseTCP_Bytes(t *testing.T) {
	example := ReadHoldingRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
			UnitID: 1,
			// +1 function code
			RegisterByteLen: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadHoldingRegistersResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadHoldingRegistersResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x3, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadHoldingRegistersResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x5, 0x10, 0x3, 0x2, 0x1, 0x2},
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

func TestParseReadHoldingRegistersResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadHoldingRegistersResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x03, 0x02, 0xCD, 0x6B},
			expect: &ReadHoldingRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
					UnitID:          3,
					RegisterByteLen: 2,
					Data:            []byte{0xcd, 0x6b},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x03, 0x02, 0xCD},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x03, 0x01, 0xCD, 0x6B},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadHoldingRegistersResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadHoldingRegistersResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadHoldingRegistersResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0x3, 0x2, 0x1, 0x2, 0xe, 0xd3},
			expect: &ReadHoldingRegistersResponseRTU{
				ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
					UnitID:          16,
					RegisterByteLen: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0x3, 0x2, 0x1, 0xe, 0xd3},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x3, 0x1, 0x1, 0x2, 0xe, 0xd3},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadHoldingRegistersResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadHoldingRegistersResponseRTU_Bytes(t *testing.T) {
	example := ReadHoldingRegistersResponseRTU{
		ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
			UnitID: 1,
			// +1 function code
			RegisterByteLen: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadHoldingRegistersResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadHoldingRegistersResponseRTU) {},
			expect: []byte{0x1, 0x3, 0x2, 0x0, 0x1, 0x84, 0x79},
		},
		{
			name: "ok2",
			given: func(r *ReadHoldingRegistersResponseRTU) {
				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x3, 0x2, 0x1, 0x2, 0x16, 0xc4},
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

func TestReadHoldingRegistersResponse_FunctionCode(t *testing.T) {
	given := ReadHoldingRegistersResponse{}
	assert.Equal(t, uint8(3), given.FunctionCode())
}

func TestReadHoldingRegistersResponse_Bytes(t *testing.T) {
	example := ReadHoldingRegistersResponse{
		UnitID: 1,
		// +1 function code
		RegisterByteLen: 2,
		Data:            []byte{0x0, 0x1},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadHoldingRegistersResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadHoldingRegistersResponse) {},
			expect: []byte{0x1, 0x3, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadHoldingRegistersResponse) {
				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x3, 0x2, 0x1, 0x2},
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

func TestReadHoldingRegistersResponse_AsRegisters(t *testing.T) {
	example := ReadHoldingRegistersResponse{
		UnitID: 1,
		// +1 function code
		RegisterByteLen: 2,
		Data:            []byte{0x0, 0x1},
	}
	var testCases = []struct {
		name                    string
		given                   func(r *ReadHoldingRegistersResponse)
		whenRequestStartAddress uint16
		expect                  *Registers
		expectError             string
	}{
		{
			name:                    "ok",
			given:                   func(r *ReadHoldingRegistersResponse) {},
			whenRequestStartAddress: 1,
			expect: &Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     1,
				endAddress:       2,
				data:             []byte{0x0, 0x1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			given := example
			if tc.given != nil {
				tc.given(&given)
			}

			regs, err := given.AsRegisters(tc.whenRequestStartAddress)

			assert.Equal(t, tc.expect, regs)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
