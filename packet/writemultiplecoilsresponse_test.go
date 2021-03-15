package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteMultipleCoilsResponseTCP_Bytes(t *testing.T) {
	example := WriteMultipleCoilsResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
			Length:        6,
		},
		WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
			UnitID: 1,
			// +1 function code
			StartAddress: 2,
			CoilCount:    1,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleCoilsResponseTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleCoilsResponseTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0xf, 0x0, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleCoilsResponseTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 2
				r.CoilCount = 20
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x6, 0x10, 0xf, 0x0, 0x2, 0x0, 0x14},
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

func TestParseWriteMultipleCoilsResponseTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteMultipleCoilsResponseTCP
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0xf, 0x0, 0x2, 0x0, 0x1},
			expect: &WriteMultipleCoilsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
					Length:        6,
				},
				WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
					UnitID:       3,
					StartAddress: 2,
					CoilCount:    1,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0xf, 0x0, 0x2, 0x0},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0xf, 0x0, 0x2, 0x0, 0x1, 0xff},
			expectError: "received data length does not match PDU len in packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteMultipleCoilsResponseTCP(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteMultipleCoilsResponseRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		given       []byte
		expect      *WriteMultipleCoilsResponseRTU
		expectError string
	}{
		{
			name:  "ok",
			given: []byte{0x10, 0xf, 0x0, 0x2, 0x0, 0x1, 0xec, 0xd2},
			expect: &WriteMultipleCoilsResponseRTU{
				WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
					UnitID:       16,
					StartAddress: 2,
					CoilCount:    1,
				},
			},
		},
		{
			name:        "nok, too short",
			given:       []byte{0x10, 0xf, 0x0, 0x2, 0x0, 0x1, 0xec},
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, byte len does not match packet len",
			given:       []byte{0x10, 0xf, 0x0, 0x2, 0x0, 0x1, 0xec, 0xd2, 0xff},
			expectError: "received data length too long to be valid packet",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := ParseWriteMultipleCoilsResponseRTU(tc.given)

			assert.Equal(t, tc.expect, packet)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteMultipleCoilsResponseRTU_Bytes(t *testing.T) {
	example := WriteMultipleCoilsResponseRTU{
		WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
			UnitID: 1,
			// +1 function code
			StartAddress: 2,
			CoilCount:    1,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleCoilsResponseRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleCoilsResponseRTU) {},
			expect: []byte{0x1, 0xf, 0x0, 0x2, 0x0, 0x1, 0xc7, 0x56},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleCoilsResponseRTU) {
				r.UnitID = 16
				r.StartAddress = 2
				r.CoilCount = 2
			},
			expect: []byte{0x10, 0xf, 0x0, 0x2, 0x0, 0x2, 0x7, 0x66},
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

func TestWriteMultipleCoilsResponse_FunctionCode(t *testing.T) {
	given := WriteMultipleCoilsResponse{}
	assert.Equal(t, uint8(15), given.FunctionCode())
}

func TestWriteMultipleCoilsResponse_Bytes(t *testing.T) {
	example := WriteMultipleCoilsResponse{
		UnitID: 1,
		// +1 function code
		StartAddress: 2,
		CoilCount:    1,
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleCoilsResponse)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleCoilsResponse) {},
			expect: []byte{0x1, 0xf, 0x0, 0x2, 0x0, 0x1},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleCoilsResponse) {
				r.UnitID = 16
				r.StartAddress = 2
				r.CoilCount = 2
			},
			expect: []byte{0x10, 0xf, 0x0, 0x2, 0x0, 0x2},
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
