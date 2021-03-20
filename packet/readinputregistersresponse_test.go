package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadInputRegistersResponseTCP_Bytes(t *testing.T) {
	example := ReadInputRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadInputRegistersResponse: ReadInputRegistersResponse{
			UnitID: 1,
			// +1 function code
			RegisterByteLen: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadInputRegistersResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadInputRegistersResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x4, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadInputRegistersResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x5, 0x10, 0x4, 0x2, 0x1, 0x2},
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

func TestParseReadInputRegistersResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadInputRegistersResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x04, 0x02, 0xCD, 0x6B},
			expect: &ReadInputRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadInputRegistersResponse: ReadInputRegistersResponse{
					UnitID:          3,
					RegisterByteLen: 2,
					Data:            []byte{0xcd, 0x6b},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x04, 0x02, 0xCD},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x04, 0x01, 0xCD, 0x6B},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadInputRegistersResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadInputRegistersResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadInputRegistersResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0x4, 0x2, 0x1, 0x2, 0xb9, 0xd2},
			expect: &ReadInputRegistersResponseRTU{
				ReadInputRegistersResponse: ReadInputRegistersResponse{
					UnitID:          16,
					RegisterByteLen: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0x4, 0x2, 0x1, 0x2, 0xb9},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x4, 0x1, 0x1, 0x2, 0xb9, 0xd2},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadInputRegistersResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadInputRegistersResponseRTU_Bytes(t *testing.T) {
	example := ReadInputRegistersResponseRTU{
		ReadInputRegistersResponse: ReadInputRegistersResponse{
			UnitID: 1,
			// +1 function code
			RegisterByteLen: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadInputRegistersResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadInputRegistersResponseRTU) {},
			expect: []byte{0x1, 0x4, 0x2, 0x0, 0x1, 0x78, 0xf0},
		},
		{
			name: "ok2",
			given: func(r *ReadInputRegistersResponseRTU) {
				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x4, 0x2, 0x1, 0x2, 0xc5, 0x62},
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

func TestReadInputRegistersResponse_FunctionCode(t *testing.T) {
	given := ReadInputRegistersResponse{}
	assert.Equal(t, uint8(4), given.FunctionCode())
}

func TestReadInputRegistersResponse_Bytes(t *testing.T) {
	example := ReadInputRegistersResponse{
		UnitID: 1,
		// +1 function code
		RegisterByteLen: 2,
		Data:            []byte{0x0, 0x1},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadInputRegistersResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadInputRegistersResponse) {},
			expect: []byte{0x1, 0x4, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadInputRegistersResponse) {
				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x4, 0x2, 0x1, 0x2},
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

func TestReadInputRegistersResponse_AsRegisters(t *testing.T) {
	example := ReadInputRegistersResponse{
		UnitID: 1,
		// +1 function code
		RegisterByteLen: 2,
		Data:            []byte{0x0, 0x1},
	}
	var testCases = []struct {
		name                    string
		given                   func(r *ReadInputRegistersResponse)
		whenRequestStartAddress uint16
		expect                  *Registers
		expectError             string
	}{
		{
			name:                    "ok",
			given:                   func(r *ReadInputRegistersResponse) {},
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
