package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteMultipleRegistersResponseTCP_Bytes(t *testing.T) {
	example := WriteMultipleRegistersResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
			UnitID: 1,
			// +1 function code
			StartAddress:  2,
			RegisterCount: 1,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleRegistersResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleRegistersResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x10, 0x0, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleRegistersResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 2
				r.RegisterCount = 20
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x6, 0x10, 0x10, 0x0, 0x2, 0x0, 0x14},
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

func TestParseWriteMultipleRegistersResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteMultipleRegistersResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0x10, 0x0, 0x2, 0x0, 0x1},
			expect: &WriteMultipleRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
				},
				WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
					UnitID:        3,
					StartAddress:  2,
					RegisterCount: 1,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0x10, 0x0, 0x2, 0x0},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0x10, 0x0, 0x2, 0x0, 0x1, 0xff},
			expectError: "received data length does not match PDU len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteMultipleRegistersResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteMultipleRegistersResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteMultipleRegistersResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0x10, 0x0, 0x2, 0x0, 0x1, 0xec, 0xd2},
			expect: &WriteMultipleRegistersResponseRTU{
				WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
					UnitID:        16,
					StartAddress:  2,
					RegisterCount: 1,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0x10, 0x0, 0x2, 0x0, 0x1, 0xec},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0x10, 0x0, 0x2, 0x0, 0x1, 0xec, 0xd2, 0xff},
			expectError: "received data length too long to be valid packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteMultipleRegistersResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteMultipleRegistersResponseRTU_Bytes(t *testing.T) {
	example := WriteMultipleRegistersResponseRTU{
		WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
			UnitID: 1,
			// +1 function code
			StartAddress:  2,
			RegisterCount: 1,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleRegistersResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleRegistersResponseRTU) {},
			expect: []byte{0x1, 0x10, 0x0, 0x2, 0x0, 0x1, 0xa0, 0x9},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleRegistersResponseRTU) {
				r.UnitID = 16
				r.StartAddress = 2
				r.RegisterCount = 2
			},
			expect: []byte{0x10, 0x10, 0x0, 0x2, 0x0, 0x2, 0xe3, 0x49},
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

func TestWriteMultipleRegistersResponse_FunctionCode(t *testing.T) {
	given := WriteMultipleRegistersResponse{}
	assert.Equal(t, uint8(16), given.FunctionCode())
}

func TestWriteMultipleRegistersResponse_Bytes(t *testing.T) {
	example := WriteMultipleRegistersResponse{
		UnitID: 1,
		// +1 function code
		StartAddress:  2,
		RegisterCount: 1,
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleRegistersResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleRegistersResponse) {},
			expect: []byte{0x1, 0x10, 0x0, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleRegistersResponse) {
				r.UnitID = 16
				r.StartAddress = 2
				r.RegisterCount = 2
			},
			expect: []byte{0x10, 0x10, 0x0, 0x2, 0x0, 0x2},
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
