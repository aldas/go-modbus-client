package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReadHoldingRegistersRequestTCP(t *testing.T) {
	expect := ReadHoldingRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
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
		expect           *ReadHoldingRegistersRequestTCP
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
			whenQuantity:     125 + 1,
			expect:           nil,
			expectError:      "quantity is out of range (1-125): 126",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadHoldingRegistersRequestTCP(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

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

func TestReadHoldingRegistersRequestTCP_Bytes(t *testing.T) {
	example := ReadHoldingRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadHoldingRegistersRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadHoldingRegistersRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x3, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadHoldingRegistersRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x10, 0x03, 0x00, 0x6B, 0x00, 0x03},
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

func TestReadHoldingRegistersRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name         string
		whenQuantity uint16
		expect       int
	}{
		{
			name:         "ok, 2 byte",
			whenQuantity: 1,
			expect:       9 + 2,
		},
		{
			name:         "ok, 4 bytes",
			whenQuantity: 2,
			expect:       9 + 4,
		},
		{
			name:         "ok, 250 bytes",
			whenQuantity: 125,
			expect:       9 + 250,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadHoldingRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewReadHoldingRegistersRequestRTU(t *testing.T) {
	expect := ReadHoldingRegistersRequestRTU{
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
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
		expect           *ReadHoldingRegistersRequestRTU
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
			whenQuantity:     125 + 1,
			expect:           nil,
			expectError:      "quantity is out of range (1-125): 126",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadHoldingRegistersRequestRTU(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadHoldingRegistersRequestRTU_Bytes(t *testing.T) {
	example := ReadHoldingRegistersRequestRTU{
		ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadHoldingRegistersRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadHoldingRegistersRequestRTU) {},
			expect: []byte{0x1, 0x3, 0x0, 0xc8, 0x0, 0xa, 0x44, 0x33},
		},
		{
			name: "ok2",
			given: func(r *ReadHoldingRegistersRequestRTU) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x03, 0x00, 0x6B, 0x00, 0x03, 0x77, 0x56},
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

func TestReadHoldingRegistersRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name         string
		whenQuantity uint16
		expect       int
	}{
		{
			name:         "ok, 2 byte",
			whenQuantity: 1,
			expect:       3 + 2 + 2,
		},
		{
			name:         "ok, 4 bytes",
			whenQuantity: 2,
			expect:       3 + 4 + 2,
		},
		{
			name:         "ok, 250 bytes",
			whenQuantity: 125,
			expect:       3 + 250 + 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadHoldingRegistersRequestRTU{
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestReadHoldingRegistersRequest_FunctionCode(t *testing.T) {
	given := ReadHoldingRegistersRequest{}
	assert.Equal(t, uint8(3), given.FunctionCode())
}

func TestReadHoldingRegistersRequest_Bytes(t *testing.T) {
	example := ReadHoldingRegistersRequest{
		UnitID:       1,
		StartAddress: 200,
		Quantity:     10,
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadHoldingRegistersRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadHoldingRegistersRequest) {},
			expect: []byte{0x1, 0x3, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadHoldingRegistersRequest) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x03, 0x00, 0x6B, 0x00, 0x03},
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

func TestParseReadHoldingRegistersRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadHoldingRegistersRequestTCP
		expectError string
	}{
		{
			name: "ok, parse ReadHoldingRegistersRequestTCP",
			when: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x6B, 0x00, 0x01},
			expect: &ReadHoldingRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x01,
					ProtocolID:    0,
				},
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       0x1,
					StartAddress: 0x6b,
					Quantity:     0x01,
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x07, 0x01, 0x03, 0x00, 0x6B, 0x00, 0x01},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x01},
			expect:      nil,
			expectError: "received function code in packet is not 0x03",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x6B, 0x00, 0x00},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x6B, 0x00, 0x7e},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadHoldingRegistersRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadHoldingRegistersRequestRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadHoldingRegistersRequestRTU
		expectError string
	}{
		{
			name: "ok, parse ReadHoldingRegistersRequestRTU",
			when: []byte{0x01, 0x03, 0x00, 0x6B, 0x00, 0x01, 0xFF, 0xFF},
			expect: &ReadHoldingRegistersRequestRTU{
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       0x1,
					StartAddress: 0x6b,
					Quantity:     0x01,
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x01, 0x03, 0x00, 0x6B, 0x00},
			expect:      nil,
			expectError: "invalid data length to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x00, 0x00, 0x6B, 0x00, 0x01, 0xFF, 0xFF},
			expect:      nil,
			expectError: "received function code in packet is not 0x03",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x01, 0x03, 0x00, 0x6B, 0x00, 0x00, 0xFF, 0xFF},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x01, 0x03, 0x00, 0x6B, 0x00, 0x7e, 0xFF, 0xFF},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadHoldingRegistersRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
