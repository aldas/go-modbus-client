package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadServerIDResponseTCP_Bytes(t *testing.T) {
	example := ReadServerIDResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadServerIDResponse: ReadServerIDResponse{
			UnitID:         1,
			Status:         0xff,
			ServerID:       []byte{0x01},
			AdditionalData: []byte{0x03, 0x04},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadServerIDResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadServerIDResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x7, 0x1, 0x11, 0x1, 0x01, 0xFF, 0x03, 0x04},
		},
		{
			name: "ok2",
			given: func(r *ReadServerIDResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.Status = 0xFE
				r.ServerID = []byte{0x10, 0x11}
				r.AdditionalData = nil
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x6, 0x10, 0x11, 0x2, 0x10, 0x11, 0xfe},
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

func TestParseReadServerIDResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadServerIDResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x08, 0x03, 0x11, 0x02, 0x01, 0x02, 0xFF, 0x03, 0x04},
			expect: &ReadServerIDResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadServerIDResponse: ReadServerIDResponse{
					UnitID:         3,
					Status:         0xff,
					ServerID:       []byte{0x01, 0x02},
					AdditionalData: []byte{0x03, 0x04},
				},
			},
		},
		{
			name:  "ok, without additional data",
			given: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x08, 0x03, 0x11, 0x01, 0x01, 0xFF},
			expect: &ReadServerIDResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				ReadServerIDResponse: ReadServerIDResponse{
					UnitID:         3,
					Status:         0xff,
					ServerID:       []byte{0x01},
					AdditionalData: nil,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x08, 0x03, 0x11, 0x00, 0x01},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, too small",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x08, 0x03, 0x11, 0x00, 0x01, 0xFF},
			expectError: "server id length too small to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x08, 0x03, 0x11, 0x02, 0x01, 0xFF},
			expectError: "received data length too short to be valid packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadServerIDResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadServerIDResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *ReadServerIDResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0x11, 0x02, 0x01, 0x02, 0xff, 0x03, 0x04, 0xec, 0xd2},
			expect: &ReadServerIDResponseRTU{
				ReadServerIDResponse: ReadServerIDResponse{
					UnitID:         16,
					Status:         0xff,
					ServerID:       []byte{0x01, 0x02},
					AdditionalData: []byte{0x03, 0x04},
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0x11, 0x01, 0x01, 0xff, 0xCC},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x11, 0x00, 0x01, 0xff, 0xCC, 0xCC},
			expectError: "server id length too small to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x11, 0x02, 0x01, 0xff, 0xCC, 0xCC},
			expectError: "received data length too short to be valid packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseReadServerIDResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadServerIDResponseRTU_Bytes(t *testing.T) {
	example := ReadServerIDResponseRTU{
		ReadServerIDResponse: ReadServerIDResponse{
			UnitID:         1,
			Status:         0xff,
			ServerID:       []byte{0x01, 0x02},
			AdditionalData: []byte{0x03, 0x04},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadServerIDResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadServerIDResponseRTU) {},
			expect: []byte{0x1, 0x11, 0x02, 0x01, 0x02, 0xff, 0x03, 0x04, 0x8c, 0x5f},
		},
		{
			name: "ok2",
			given: func(r *ReadServerIDResponseRTU) {
				r.UnitID = 16
				r.Status = 0xFE
				r.ServerID = []byte{0x10, 0x11}
				r.AdditionalData = nil
			},
			expect: []byte{0x10, 0x11, 0x2, 0x10, 0x11, 0xfe, 0x73, 0x25},
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

func TestReadServerIDResponse_FunctionCode(t *testing.T) {
	given := ReadServerIDResponse{}
	assert.Equal(t, uint8(17), given.FunctionCode())
}

func TestReadServerIDResponse_Bytes(t *testing.T) {
	example := ReadServerIDResponse{
		UnitID:         1,
		Status:         0xff,
		ServerID:       []byte{0x01},
		AdditionalData: []byte{0x03, 0x04},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadServerIDResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadServerIDResponse) {},
			expect: []byte{0x1, 0x11, 0x1, 0x1, 0xff, 0x3, 0x4},
		},
		{
			name: "ok2",
			given: func(r *ReadServerIDResponse) {
				r.UnitID = 16
				r.Status = 0xFE
				r.ServerID = []byte{0x10, 0x11}
				r.AdditionalData = nil
			},
			expect: []byte{0x10, 0x11, 0x2, 0x10, 0x11, 0xFE},
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
