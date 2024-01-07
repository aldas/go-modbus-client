package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReadServerIDRequestTCP(t *testing.T) {
	expect := ReadServerIDRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 1,
		},
	}

	var testCases = []struct {
		name        string
		whenUnitID  uint8
		expect      *ReadServerIDRequestTCP
		expectError string
	}{
		{
			name:        "ok",
			whenUnitID:  1,
			expect:      &expect,
			expectError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadServerIDRequestTCP(tc.whenUnitID)

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

func TestReadServerIDRequestTCP_Bytes(t *testing.T) {
	example := ReadServerIDRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 1,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadServerIDRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadServerIDRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x2, 0x1, 0x11},
		},
		{
			name: "ok2",
			given: func(r *ReadServerIDRequestTCP) {
				r.TransactionID = 1
				r.UnitID = 16
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x10, 0x11},
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

func TestReadServerIDRequestTCP_ExpectedResponseLength(t *testing.T) {
	example := ReadServerIDRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 1,
		},
	}

	assert.Equal(t, 8, example.ExpectedResponseLength())
}

func TestNewReadServerIDRequestRTU(t *testing.T) {
	expect := ReadServerIDRequestRTU{
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 1,
		},
	}

	var testCases = []struct {
		name        string
		whenUnitID  uint8
		expect      *ReadServerIDRequestRTU
		expectError string
	}{
		{
			name:        "ok",
			whenUnitID:  1,
			expect:      &expect,
			expectError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadServerIDRequestRTU(tc.whenUnitID)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadServerIDRequestRTU_Bytes(t *testing.T) {
	example := ReadServerIDRequestRTU{
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 1,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadServerIDRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadServerIDRequestRTU) {},
			expect: []byte{0x1, 0x11, 0xc0, 0x2c},
		},
		{
			name: "ok2",
			given: func(r *ReadServerIDRequestRTU) {
				r.UnitID = 16
			},
			expect: []byte{0x10, 0x11, 0xcc, 0x7c},
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

func TestReadServerIDRequestRTU_ExpectedResponseLength(t *testing.T) {
	example := ReadServerIDRequestRTU{
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 1,
		},
	}

	assert.Equal(t, 2, example.ExpectedResponseLength())
}

func TestReadServerIDRequest_FunctionCode(t *testing.T) {
	given := ReadServerIDRequest{}
	assert.Equal(t, uint8(17), given.FunctionCode())
}

func TestReadServerIDRequest_Bytes(t *testing.T) {
	example := ReadServerIDRequest{
		UnitID: 1,
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadServerIDRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadServerIDRequest) {},
			expect: []byte{0x1, 0x11},
		},
		{
			name: "ok2",
			given: func(r *ReadServerIDRequest) {
				r.UnitID = 16
			},
			expect: []byte{0x10, 0x11},
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

func TestParseReadServerIDRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadServerIDRequestTCP
		expectError string
	}{
		{
			name: "ok, parse ReadServerIDRequestTCP",
			when: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x02, 0x10, 0x11},
			expect: &ReadServerIDRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x0102,
					ProtocolID:    0,
				},
				ReadServerIDRequest: ReadServerIDRequest{
					UnitID: 0x10,
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x03, 0x10, 0x11},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x02, 0x10, 0x12},
			expect:      nil,
			expectError: "received function code in packet is not 0x11",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadServerIDRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadServerIDRequestRTU(t *testing.T) {
	example := ReadServerIDRequestRTU{
		ReadServerIDRequest: ReadServerIDRequest{
			UnitID: 0x10,
		},
	}
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadServerIDRequestRTU
		expectError string
	}{
		{
			name:   "ok, parse ReadServerIDRequestRTU, with crc bytes",
			when:   []byte{0x10, 0x11, 0xff, 0xff},
			expect: &example,
		},
		{
			name:   "ok, parse ReadServerIDRequestRTU, without crc bytes",
			when:   []byte{0x10, 0x11},
			expect: &example,
		},
		{
			name:        "nok, invalid data length to be valid packet",
			when:        []byte{0x10, 0x11, 0xff},
			expect:      nil,
			expectError: "invalid data length to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x10, 0x12, 0xff, 0xff},
			expect:      nil,
			expectError: "received function code in packet is not 0x11",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadServerIDRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
