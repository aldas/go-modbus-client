package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteSingleCoilResponseTCP_Bytes(t *testing.T) {
	example := WriteSingleCoilResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteSingleCoilResponse: WriteSingleCoilResponse{
			UnitID: 1,
			// +1 function code
			StartAddress: 2,
			CoilState:    true,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleCoilResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleCoilResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x5, 0x0, 0x2, 0xff, 0x0},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleCoilResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 2
				r.CoilState = false
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x6, 0x10, 0x5, 0x0, 0x2, 0x0, 0x0},
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

func TestParseWriteSingleCoilResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteSingleCoilResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x06, 0x3, 0x5, 0x0, 0x2, 0xff, 0x0},
			expect: &WriteSingleCoilResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				WriteSingleCoilResponse: WriteSingleCoilResponse{
					UnitID:       3,
					StartAddress: 2,
					CoilState:    true,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x3, 0x5, 0x0, 0x2, 0xff},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, PDU len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x3, 0x5, 0x0, 0x2, 0xff, 0x0},
			expectError: "received data length does not match PDU len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteSingleCoilResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteSingleCoilResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteSingleCoilResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x1, 0x5, 0x0, 0x2, 0xff, 0x0, 0x13, 0x9d},
			expect: &WriteSingleCoilResponseRTU{
				WriteSingleCoilResponse: WriteSingleCoilResponse{
					UnitID:       1,
					StartAddress: 2,
					CoilState:    true,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x1, 0x5, 0x0, 0x2, 0xff, 0x0, 0x13},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, too long",
			given:       []byte{0x1, 0x5, 0x0, 0x2, 0xff, 0x0, 0x13, 0x9d, 0xff},
			expectError: "received data length too long to be valid packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteSingleCoilResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteSingleCoilResponseRTU_Bytes(t *testing.T) {
	example := WriteSingleCoilResponseRTU{
		WriteSingleCoilResponse: WriteSingleCoilResponse{
			UnitID: 1,
			// +1 function code
			StartAddress: 2,
			CoilState:    true,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleCoilResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleCoilResponseRTU) {},
			expect: []byte{0x1, 0x5, 0x0, 0x2, 0xff, 0x0, 0x2d, 0xfa},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleCoilResponseRTU) {
				r.UnitID = 16
				r.StartAddress = 2
				r.CoilState = false
			},
			expect: []byte{0x10, 0x5, 0x0, 0x2, 0x0, 0x0, 0x6f, 0x4b},
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

func TestWriteSingleCoilResponse_FunctionCode(t *testing.T) {
	given := WriteSingleCoilResponse{}
	assert.Equal(t, uint8(5), given.FunctionCode())
}

func TestWriteSingleCoilResponse_Bytes(t *testing.T) {
	example := WriteSingleCoilResponse{
		UnitID: 1,
		// +1 function code
		StartAddress: 2,
		CoilState:    true,
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleCoilResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleCoilResponse) {},
			expect: []byte{0x1, 0x5, 0x0, 0x2, 0xff, 0x0},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleCoilResponse) {
				r.UnitID = 16
				r.StartAddress = 2
				r.CoilState = false
			},
			expect: []byte{0x10, 0x5, 0x0, 0x2, 0x0, 0x0},
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
