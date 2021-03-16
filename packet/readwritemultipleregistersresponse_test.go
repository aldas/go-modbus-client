package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadWriteMultipleRegistersResponseTCP_Bytes(t *testing.T) {
	example := ReadWriteMultipleRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
			Length:        5,
		},
		ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
			UnitID: 1,
			// +1 function code
			RegisterByteLen: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadWriteMultipleRegistersResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadWriteMultipleRegistersResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x17, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadWriteMultipleRegistersResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x5, 0x10, 0x17, 0x2, 0x1, 0x2},
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

func TestParseReadWriteMultipleRegistersResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadWriteMultipleRegistersResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x17, 0x02, 0xCD, 0x6B},
			expect: &ReadWriteMultipleRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
					Length:        5,
				},
				ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
					UnitID:          3,
					RegisterByteLen: 2,
					Data:            []byte{0xcd, 0x6b},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x17, 0x02, 0xCD},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x17, 0x01, 0xCD, 0x6B},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadWriteMultipleRegistersResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadWriteMultipleRegistersResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadWriteMultipleRegistersResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0x17, 0x2, 0x1, 0x2, 0xe, 0xd3},
			expect: &ReadWriteMultipleRegistersResponseRTU{
				ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
					UnitID:          16,
					RegisterByteLen: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0x17, 0x2, 0x1, 0xe, 0xd3},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x17, 0x1, 0x1, 0x2, 0xe, 0xd3},
			expectError: "received data length does not match byte len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadWriteMultipleRegistersResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadWriteMultipleRegistersResponseRTU_Bytes(t *testing.T) {
	example := ReadWriteMultipleRegistersResponseRTU{
		ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
			UnitID: 1,
			// +1 function code
			RegisterByteLen: 2,
			Data:            []byte{0x0, 0x1},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadWriteMultipleRegistersResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadWriteMultipleRegistersResponseRTU) {},
			expect: []byte{0x1, 0x17, 0x2, 0x0, 0x1, 0xe7, 0x20},
		},
		{
			name: "ok2",
			given: func(r *ReadWriteMultipleRegistersResponseRTU) {
				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x17, 0x2, 0x1, 0x2, 0x1a, 0xd0},
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

func TestReadWriteMultipleRegistersResponse_FunctionCode(t *testing.T) {
	given := ReadWriteMultipleRegistersResponse{}
	assert.Equal(t, uint8(23), given.FunctionCode())
}

func TestReadWriteMultipleRegistersResponse_Bytes(t *testing.T) {
	example := ReadWriteMultipleRegistersResponse{
		UnitID: 1,
		// +1 function code
		RegisterByteLen: 2,
		Data:            []byte{0x0, 0x1},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadWriteMultipleRegistersResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadWriteMultipleRegistersResponse) {},
			expect: []byte{0x1, 0x17, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *ReadWriteMultipleRegistersResponse) {
				r.UnitID = 16
				r.RegisterByteLen = 2
				r.Data = []byte{0x1, 0x2}
			},
			expect: []byte{0x10, 0x17, 0x2, 0x1, 0x2},
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

func TestReadWriteMultipleRegistersResponse_AsRegisters(t *testing.T) {
	example := ReadWriteMultipleRegistersResponse{
		UnitID: 1,
		// +1 function code
		RegisterByteLen: 2,
		Data:            []byte{0x0, 0x1},
	}
	var testCases = []struct {
		name                    string
		given                   func(r *ReadWriteMultipleRegistersResponse)
		whenRequestStartAddress uint16
		expect                  *Registers
		expectError             string
	}{
		{
			name:                    "ok",
			given:                   func(r *ReadWriteMultipleRegistersResponse) {},
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
