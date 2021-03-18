package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWriteSingleCoilRequestTCP(t *testing.T) {
	expect := WriteSingleCoilRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID:    1,
			Address:   200,
			CoilState: true,
		},
	}

	var testCases = []struct {
		name          string
		whenUnitID    uint8
		whenAddress   uint16
		whenCoilState bool
		expect        *WriteSingleCoilRequestTCP
		expectError   string
	}{
		{
			name:          "ok, state true",
			whenUnitID:    1,
			whenAddress:   200,
			whenCoilState: true,
			expect:        &expect,
			expectError:   "",
		},
		{
			name:          "ok, state false",
			whenUnitID:    1,
			whenAddress:   200,
			whenCoilState: false,
			expect: &WriteSingleCoilRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:    1,
					Address:   200,
					CoilState: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteSingleCoilRequestTCP(tc.whenUnitID, tc.whenAddress, tc.whenCoilState)

			expect := tc.expect
			if packet != nil {
				assert.NotEqual(t, uint16(0), packet.TransactionID)
				expect.TransactionID = packet.TransactionID
			}
			assert.Equal(t, expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteSingleCoilRequestTCP_Bytes(t *testing.T) {
	example := WriteSingleCoilRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID:    1,
			Address:   200,
			CoilState: true,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleCoilRequestTCP)
		expect []byte
	}{
		{
			name:   "ok, state true",
			given:  func(r *WriteSingleCoilRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x5, 0x0, 0xc8, 0xff, 0x0},
		},
		{
			name: "ok, state false",
			given: func(r *WriteSingleCoilRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.Address = 107
				r.CoilState = false
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x10, 0x05, 0x00, 0x6B, 0x00, 0x00},
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

func TestWriteSingleCoilRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name          string
		whenCoilState bool
		expect        int
	}{
		{
			name:          "ok",
			whenCoilState: true,
			expect:        11,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteSingleCoilRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:    1,
					Address:   200,
					CoilState: tc.whenCoilState,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewWriteSingleCoilRequestRTU(t *testing.T) {
	expect := WriteSingleCoilRequestRTU{
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID:    1,
			Address:   200,
			CoilState: true,
		},
	}

	var testCases = []struct {
		name          string
		whenUnitID    uint8
		whenAddress   uint16
		whenCoilState bool
		expect        *WriteSingleCoilRequestRTU
		expectError   string
	}{
		{
			name:          "ok, state true",
			whenUnitID:    1,
			whenAddress:   200,
			whenCoilState: true,
			expect:        &expect,
			expectError:   "",
		},
		{
			name:          "ok, state false",
			whenUnitID:    1,
			whenAddress:   200,
			whenCoilState: false,
			expect: &WriteSingleCoilRequestRTU{
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:    1,
					Address:   200,
					CoilState: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteSingleCoilRequestRTU(tc.whenUnitID, tc.whenAddress, tc.whenCoilState)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteSingleCoilRequestRTU_Bytes(t *testing.T) {
	example := WriteSingleCoilRequestRTU{
		WriteSingleCoilRequest: WriteSingleCoilRequest{
			UnitID:    1,
			Address:   200,
			CoilState: true,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleCoilRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleCoilRequestRTU) {},
			expect: []byte{0x1, 0x5, 0x0, 0xc8, 0xff, 0x0, 0xc4, 0xd},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleCoilRequestRTU) {
				r.UnitID = 16
				r.Address = 107
				r.CoilState = false
			},
			expect: []byte{0x10, 0x05, 0x00, 0x6B, 0x0, 0x0, 0x57, 0xbf},
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

func TestWriteSingleCoilRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name          string
		whenCoilState uint16
		expect        int
	}{
		{
			name:          "ok",
			whenCoilState: 8,
			expect:        6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteSingleCoilRequestRTU{
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:    1,
					Address:   200,
					CoilState: true,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestWriteSingleCoilRequest_FunctionCode(t *testing.T) {
	given := WriteSingleCoilRequest{}
	assert.Equal(t, uint8(5), given.FunctionCode())
}

func TestWriteSingleCoilRequest_Bytes(t *testing.T) {
	example := WriteSingleCoilRequest{
		UnitID:    1,
		Address:   200,
		CoilState: true,
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleCoilRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleCoilRequest) {},
			expect: []byte{0x1, 0x5, 0x0, 0xc8, 0xff, 0x0},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleCoilRequest) {
				r.UnitID = 16
				r.Address = 107
				r.CoilState = false
			},
			expect: []byte{0x10, 0x05, 0x00, 0x6B, 0x00, 0x00},
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
