package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWriteSingleRegisterRequestTCP(t *testing.T) {
	expect := WriteSingleRegisterRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID:  1,
			Address: 200,
			Data:    [2]byte{0x1, 0x2},
		},
	}

	var testCases = []struct {
		name        string
		whenUnitID  uint8
		whenAddress uint16
		whenData    []byte
		expect      *WriteSingleRegisterRequestTCP
		expectError string
	}{
		{
			name:        "ok, state true",
			whenUnitID:  1,
			whenAddress: 200,
			whenData:    []byte{0x1, 0x2},
			expect:      &expect,
			expectError: "",
		},
		{
			name:        "ok, state false",
			whenUnitID:  1,
			whenAddress: 200,
			whenData:    []byte{0x0, 0x0},
			expect: &WriteSingleRegisterRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  1,
					Address: 200,
					Data:    [2]byte{0x0, 0x0},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteSingleRegisterRequestTCP(tc.whenUnitID, tc.whenAddress, tc.whenData)

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

func TestWriteSingleRegisterRequestTCP_Bytes(t *testing.T) {
	example := WriteSingleRegisterRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID:  1,
			Address: 200,
			Data:    [2]byte{0x1, 0x2},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleRegisterRequestTCP)
		expect []byte
	}{
		{
			name:   "ok, state true",
			given:  func(r *WriteSingleRegisterRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x6, 0x0, 0xc8, 0x1, 0x2},
		},
		{
			name: "ok, state false",
			given: func(r *WriteSingleRegisterRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.Address = 107
				r.Data = [2]byte{0x0, 0x0}
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x10, 0x06, 0x00, 0x6B, 0x00, 0x00},
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

func TestWriteSingleRegisterRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name          string
		whenCoilState bool
		expect        int
	}{
		{
			name:          "ok",
			whenCoilState: true,
			expect:        12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteSingleRegisterRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  1,
					Address: 200,
					Data:    [2]byte{0x1, 0x2},
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewWriteSingleRegisterRequestRTU(t *testing.T) {
	expect := WriteSingleRegisterRequestRTU{
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID:  1,
			Address: 200,
			Data:    [2]byte{0x1, 0x2},
		},
	}

	var testCases = []struct {
		name        string
		whenUnitID  uint8
		whenAddress uint16
		whenData    []byte
		expect      *WriteSingleRegisterRequestRTU
		expectError string
	}{
		{
			name:        "ok, state true",
			whenUnitID:  1,
			whenAddress: 200,
			whenData:    []byte{0x1, 0x2},
			expect:      &expect,
			expectError: "",
		},
		{
			name:        "ok, state false",
			whenUnitID:  1,
			whenAddress: 200,
			whenData:    []byte{0x0, 0x0},
			expect: &WriteSingleRegisterRequestRTU{
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  1,
					Address: 200,
					Data:    [2]byte{0x0, 0x0},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteSingleRegisterRequestRTU(tc.whenUnitID, tc.whenAddress, tc.whenData)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteSingleRegisterRequestRTU_Bytes(t *testing.T) {
	example := WriteSingleRegisterRequestRTU{
		WriteSingleRegisterRequest: WriteSingleRegisterRequest{
			UnitID:  1,
			Address: 200,
			Data:    [2]byte{0x1, 0x2},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleRegisterRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleRegisterRequestRTU) {},
			expect: []byte{0x1, 0x6, 0x0, 0xc8, 0x1, 0x2, 0x88, 0x65},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleRegisterRequestRTU) {
				r.UnitID = 16
				r.Address = 107
				r.Data = [2]byte{0x0, 0x0}
			},
			expect: []byte{0x10, 0x06, 0x00, 0x6B, 0x0, 0x0, 0xfb, 0x57},
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

func TestWriteSingleRegisterRequestRTU_ExpectedResponseLength(t *testing.T) {
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
			example := WriteSingleRegisterRequestRTU{
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  1,
					Address: 200,
					Data:    [2]byte{0x1, 0x2},
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestWriteSingleRegisterRequest_FunctionCode(t *testing.T) {
	given := WriteSingleRegisterRequest{}
	assert.Equal(t, uint8(6), given.FunctionCode())
}

func TestWriteSingleRegisterRequest_Bytes(t *testing.T) {
	example := WriteSingleRegisterRequest{
		UnitID:  1,
		Address: 200,
		Data:    [2]byte{0x1, 0x2},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteSingleRegisterRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteSingleRegisterRequest) {},
			expect: []byte{0x1, 0x6, 0x0, 0xc8, 0x1, 0x2},
		},
		{
			name: "ok2",
			given: func(r *WriteSingleRegisterRequest) {
				r.UnitID = 16
				r.Address = 107
				r.Data = [2]byte{0x0, 0x0}
			},
			expect: []byte{0x10, 0x06, 0x00, 0x6B, 0x00, 0x00},
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

func TestParseWriteSingleRegisterRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *WriteSingleRegisterRequestTCP
		expectError string
	}{
		{
			name: "ok, parse WriteSingleRegisterRequestTCP",
			when: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x11, 0x06, 0x00, 0x6B, 0x01, 0x02},
			expect: &WriteSingleRegisterRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x01,
					ProtocolID:    0,
				},
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  0x11,
					Address: 0x6b,
					Data:    [2]byte{0x01, 0x02},
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x07, 0x11, 0x06, 0x00, 0x6B, 0x01, 0x02},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x11, 0x01, 0x00, 0x6B, 0x01, 0x02},
			expect:      nil,
			expectError: "received function code in packet is not 0x06",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWriteSingleRegisterRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteSingleRegisterRequestRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *WriteSingleRegisterRequestRTU
		expectError string
	}{
		{
			name: "ok, parse WriteSingleRegisterRequestRTU with crc",
			when: []byte{0x11, 0x06, 0x00, 0x6B, 0x01, 0x02, 0xff, 0xff},
			expect: &WriteSingleRegisterRequestRTU{
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  0x11,
					Address: 0x6b,
					Data:    [2]byte{0x01, 0x02},
				},
			},
		},
		{
			name: "ok, parse WriteSingleRegisterRequestRTU without crc",
			when: []byte{0x11, 0x06, 0x00, 0x6B, 0x01, 0x02},
			expect: &WriteSingleRegisterRequestRTU{
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  0x11,
					Address: 0x6b,
					Data:    [2]byte{0x01, 0x02},
				},
			},
		},
		{
			name:        "nok, invalid length",
			when:        []byte{0x11, 0x06, 0x00, 0x6B, 0x01, 0x02, 0xff},
			expect:      nil,
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x11, 0x00, 0x00, 0x6B, 0x01, 0x02, 0xff, 0xff},
			expect:      nil,
			expectError: "received function code in packet is not 0x06",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWriteSingleRegisterRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
