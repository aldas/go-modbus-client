package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReadCoilsRequestTCP(t *testing.T) {
	expect := ReadCoilsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name             string
		whenUnitID       uint8
		whenStartAddress uint16
		whenQuantity     uint16
		expect           *ReadCoilsRequestTCP
		expectError      string
	}{
		{
			name:             "ok",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenQuantity:     10,
			expect:           &expect,
			expectError:      "",
		},
		{
			name:             "nok, quantity too big",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenQuantity:     2000 + 1,
			expect:           nil,
			expectError:      "quantity is out of range (1-2000): 2001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadCoilsRequestTCP(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

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

func TestReadCoilsRequestTCP_Bytes(t *testing.T) {
	example := ReadCoilsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadCoilsRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadCoilsRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadCoilsRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
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

func TestReadCoilsRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name         string
		whenQuantity uint16
		expect       int
	}{
		{
			name:         "ok, 1 byte",
			whenQuantity: 8,
			expect:       9 + 1,
		},
		{
			name:         "ok, 2 bytes",
			whenQuantity: 9,
			expect:       9 + 2,
		},
		{
			name:         "ok, 11 bytes",
			whenQuantity: 8*10 + 7,
			expect:       9 + 11,
		},
		{
			name:         "ok, 253 bytes",
			whenQuantity: 8 * 253,
			expect:       9 + 253,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadCoilsRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewReadCoilsRequestRTU(t *testing.T) {
	expect := ReadCoilsRequestRTU{
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name             string
		whenUnitID       uint8
		whenStartAddress uint16
		whenQuantity     uint16
		expect           *ReadCoilsRequestRTU
		expectError      string
	}{
		{
			name:             "ok",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenQuantity:     10,
			expect:           &expect,
			expectError:      "",
		},
		{
			name:             "nok, quantity too big",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenQuantity:     2000 + 1,
			expect:           nil,
			expectError:      "quantity is out of range (1-2000): 2001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadCoilsRequestRTU(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadCoilsRequestRTU_Bytes(t *testing.T) {
	example := ReadCoilsRequestRTU{
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadCoilsRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadCoilsRequestRTU) {},
			expect: []byte{0x1, 0x1, 0x0, 0xc8, 0x0, 0xa, 0x3d, 0xf3},
		},
		{
			name: "ok2",
			given: func(r *ReadCoilsRequestRTU) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03, 0xe, 0x96},
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

func TestReadCoilsRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name         string
		whenQuantity uint16
		expect       int
	}{
		{
			name:         "ok, 1 byte",
			whenQuantity: 8,
			expect:       4 + 1,
		},
		{
			name:         "ok, 2 bytes",
			whenQuantity: 9,
			expect:       4 + 2,
		},
		{
			name:         "ok, 11 bytes",
			whenQuantity: 8*10 + 7,
			expect:       4 + 11,
		},
		{
			name:         "ok, 253 bytes",
			whenQuantity: 8 * 253,
			expect:       4 + 253,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadCoilsRequestRTU{
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestReadCoilsRequest_FunctionCode(t *testing.T) {
	given := ReadCoilsRequest{}
	assert.Equal(t, uint8(1), given.FunctionCode())
}

func TestReadCoilsRequest_Bytes(t *testing.T) {
	example := ReadCoilsRequest{
		UnitID:       1,
		StartAddress: 200,
		Quantity:     10,
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadCoilsRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadCoilsRequest) {},
			expect: []byte{0x1, 0x1, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadCoilsRequest) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
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

func TestParseReadCoilsRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadCoilsRequestTCP
		expectError string
	}{
		{
			name: "ok, parse ReadCoilsRequestTCP",
			when: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expect: &ReadCoilsRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x0102,
					ProtocolID:    0,
				},
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       0x10,
					StartAddress: 0x6b,
					Quantity:     0x03,
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x07, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x02, 0x00, 0x6B, 0x00, 0x03},
			expect:      nil,
			expectError: "received function code in packet is not 0x01",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x00},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x7e},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadCoilsRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadCoilsRequestRTU(t *testing.T) {
	example := ReadCoilsRequestRTU{
		ReadCoilsRequest: ReadCoilsRequest{
			UnitID:       0x10,
			StartAddress: 0x6b,
			Quantity:     0x03,
		},
	}
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadCoilsRequestRTU
		expectError string
	}{
		{
			name:   "ok, parse ReadCoilsRequestRTU, with crc bytes",
			when:   []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03, 0xff, 0xff},
			expect: &example,
		},
		{
			name:   "ok, parse ReadCoilsRequestRTU, without crc bytes",
			when:   []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expect: &example,
		},
		{
			name:        "nok, invalid data length to be valid packet",
			when:        []byte{0x10, 0x01, 0x00, 0x6B, 0x00},
			expect:      nil,
			expectError: "invalid data length to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x10, 0x00, 0x00, 0x6B, 0x00, 0x03, 0xff, 0xff},
			expect:      nil,
			expectError: "received function code in packet is not 0x01",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x0, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x7e, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadCoilsRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
