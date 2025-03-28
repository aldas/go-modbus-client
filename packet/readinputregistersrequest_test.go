package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReadInputRegistersRequestTCP(t *testing.T) {
	expect := ReadInputRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadInputRegistersRequest: ReadInputRegistersRequest{
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
		expect           *ReadInputRegistersRequestTCP
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
			packet, err := NewReadInputRegistersRequestTCP(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

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

func TestReadInputRegistersRequestTCP_Bytes(t *testing.T) {
	example := ReadInputRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadInputRegistersRequest: ReadInputRegistersRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadInputRegistersRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadInputRegistersRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x4, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadInputRegistersRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x10, 0x04, 0x00, 0x6B, 0x00, 0x03},
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

func TestReadInputRegistersRequestTCP_ExpectedResponseLength(t *testing.T) {
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
			example := ReadInputRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewReadInputRegistersRequestRTU(t *testing.T) {
	expect := ReadInputRegistersRequestRTU{
		ReadInputRegistersRequest: ReadInputRegistersRequest{
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
		expect           *ReadInputRegistersRequestRTU
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
			packet, err := NewReadInputRegistersRequestRTU(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadInputRegistersRequestRTU_Bytes(t *testing.T) {
	example := ReadInputRegistersRequestRTU{
		ReadInputRegistersRequest: ReadInputRegistersRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadInputRegistersRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadInputRegistersRequestRTU) {},
			expect: []byte{0x1, 0x4, 0x0, 0xc8, 0x0, 0xa, 0xf1, 0xf3},
		},
		{
			name: "ok2",
			given: func(r *ReadInputRegistersRequestRTU) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x04, 0x00, 0x6B, 0x00, 0x03, 0xc2, 0x96},
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

func TestReadInputRegistersRequestRTU_ExpectedResponseLength(t *testing.T) {
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
			example := ReadInputRegistersRequestRTU{
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestReadInputRegistersRequest_FunctionCode(t *testing.T) {
	given := ReadInputRegistersRequest{}
	assert.Equal(t, uint8(4), given.FunctionCode())
}

func TestReadInputRegistersRequest_Bytes(t *testing.T) {
	example := ReadInputRegistersRequest{
		UnitID:       1,
		StartAddress: 200,
		Quantity:     10,
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadInputRegistersRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadInputRegistersRequest) {},
			expect: []byte{0x1, 0x4, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadInputRegistersRequest) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x04, 0x00, 0x6B, 0x00, 0x03},
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

func TestParseReadInputRegistersRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadInputRegistersRequestTCP
		expectError string
	}{
		{
			name: "ok, parse ReadInputRegistersRequestTCP",
			when: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x6B, 0x00, 0x01},
			expect: &ReadInputRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x01,
					ProtocolID:    0,
				},
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       0x1,
					StartAddress: 0x6b,
					Quantity:     0x01,
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x07, 0x01, 0x04, 0x00, 0x6B, 0x00, 0x01},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x6B, 0x00, 0x01},
			expect:      nil,
			expectError: "received function code in packet is not 0x04",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x6B, 0x00, 0x00},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x6B, 0x00, 0x7e},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadInputRegistersRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadInputRegistersRequestRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadInputRegistersRequestRTU
		expectError string
	}{
		{
			name: "ok, parse ReadHoldingRegistersRequestRTU",
			when: []byte{0x01, 0x04, 0x00, 0x6B, 0x00, 0x01, 0xFF, 0xFF},
			expect: &ReadInputRegistersRequestRTU{
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       0x1,
					StartAddress: 0x6b,
					Quantity:     0x01,
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x01, 0x04, 0x00, 0x6B, 0x00},
			expect:      nil,
			expectError: "invalid data length to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x00, 0x00, 0x6B, 0x00, 0x01, 0xFF, 0xFF},
			expect:      nil,
			expectError: "received function code in packet is not 0x04",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x01, 0x04, 0x00, 0x6B, 0x00, 0x00, 0xFF, 0xFF},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x01, 0x04, 0x00, 0x6B, 0x00, 0x7e, 0xFF, 0xFF},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadInputRegistersRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
