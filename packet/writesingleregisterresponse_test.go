package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteSingleRegisterResponseTCP_Bytes(t *testing.T) {
	example := WriteSingleRegisterResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteSingleRegisterResponse: WriteSingleRegisterResponse{
			UnitID: 1,
			// +1 function code
			Address: 2,
			Data:    [2]byte{0x1, 0x2},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleRegisterResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleRegisterResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x6, 0x0, 0x2, 0x1, 0x2},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleRegisterResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.Address = 2
				r.Data = [2]byte{0x0, 0x0}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x6, 0x10, 0x6, 0x0, 0x2, 0x0, 0x0},
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

func TestParseWriteSingleRegisterResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteSingleRegisterResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x06, 0x3, 0x6, 0x0, 0x2, 0x1, 0x2},
			expect: &WriteSingleRegisterResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				WriteSingleRegisterResponse: WriteSingleRegisterResponse{
					UnitID:  3,
					Address: 2,
					Data:    [2]byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x3, 0x6, 0x0, 0x2, 0x1},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, PDU len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x04, 0x3, 0x6, 0x0, 0x2, 0x1, 0x2},
			expectError: "received data length does not match PDU len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteSingleRegisterResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteSingleRegisterResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteSingleRegisterResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x1, 0x6, 0x0, 0x2, 0x1, 0x2, 0x3b, 0x3e},
			expect: &WriteSingleRegisterResponseRTU{
				WriteSingleRegisterResponse: WriteSingleRegisterResponse{
					UnitID:  1,
					Address: 2,
					Data:    [2]byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x1, 0x6, 0x0, 0x2, 0x1, 0x2, 0x3b},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, too long",
			given:       []byte{0x1, 0x6, 0x0, 0x2, 0x1, 0x2, 0x3b, 0x3e, 0xff},
			expectError: "received data length too long to be valid packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteSingleRegisterResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteSingleRegisterResponseRTU_Bytes(t *testing.T) {
	example := WriteSingleRegisterResponseRTU{
		WriteSingleRegisterResponse: WriteSingleRegisterResponse{
			UnitID: 1,
			// +1 function code
			Address: 2,
			Data:    [2]byte{0x1, 0x2},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleRegisterResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleRegisterResponseRTU) {},
			expect: []byte{0x1, 0x6, 0x0, 0x2, 0x1, 0x2, 0x5b, 0xa8},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleRegisterResponseRTU) {
				r.UnitID = 16
				r.Address = 2
				r.Data = [2]byte{0x0, 0x0}
			},
			expect: []byte{0x10, 0x6, 0x0, 0x2, 0x0, 0x0, 0x4b, 0x2b},
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

func TestWriteSingleRegisterResponse_FunctionCode(t *testing.T) {
	given := WriteSingleRegisterResponse{}
	assert.Equal(t, uint8(6), given.FunctionCode())
}

func TestWriteSingleRegisterResponse_Bytes(t *testing.T) {
	example := WriteSingleRegisterResponse{
		UnitID: 1,
		// +1 function code
		Address: 2,
		Data:    [2]byte{0x1, 0x2},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleRegisterResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleRegisterResponse) {},
			expect: []byte{0x1, 0x6, 0x0, 0x2, 0x1, 0x2},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleRegisterResponse) {
				r.UnitID = 16
				r.Address = 2
				r.Data = [2]byte{0x0, 0x0}
			},
			expect: []byte{0x10, 0x6, 0x0, 0x2, 0x0, 0x0},
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

func TestWriteSingleRegisterResponse_AsRegisters(t *testing.T) {
	example := WriteSingleRegisterResponse{
		UnitID: 1,
		// +1 function code
		Address: 2,
		Data:    [2]byte{0x1, 0x2},
	}
	var testCases = []struct {
		name                    string
		given                   func(r *WriteSingleRegisterResponse)
		whenRequestStartAddress uint16
		expect                  *Registers
		expectError             string
	}{
		{
			name:                    "ok",
			given:                   func(r *WriteSingleRegisterResponse) {},
			whenRequestStartAddress: 1,
			expect: &Registers{
				defaultByteOrder: BigEndianHighWordFirst,
				startAddress:     1,
				endAddress:       2,
				data:             []byte{0x1, 0x2},
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
