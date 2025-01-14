package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReadDiscreteInputsRequestTCP(t *testing.T) {
	expect := ReadDiscreteInputsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
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
		expect           *ReadDiscreteInputsRequestTCP
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
			packet, err := NewReadDiscreteInputsRequestTCP(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

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

func TestReadDiscreteInputsRequestTCP_Bytes(t *testing.T) {
	example := ReadDiscreteInputsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadDiscreteInputsRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadDiscreteInputsRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x2, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadDiscreteInputsRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x10, 0x02, 0x00, 0x6B, 0x00, 0x03},
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

func TestReadDiscreteInputsRequestTCP_ExpectedResponseLength(t *testing.T) {
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
			example := ReadDiscreteInputsRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewReadDiscreteInputsRequestRTU(t *testing.T) {
	expect := ReadDiscreteInputsRequestRTU{
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
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
		expect           *ReadDiscreteInputsRequestRTU
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
			packet, err := NewReadDiscreteInputsRequestRTU(tc.whenUnitID, tc.whenStartAddress, tc.whenQuantity)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadDiscreteInputsRequestRTU_Bytes(t *testing.T) {
	example := ReadDiscreteInputsRequestRTU{
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     10,
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadDiscreteInputsRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadDiscreteInputsRequestRTU) {},
			expect: []byte{0x1, 0x2, 0x0, 0xc8, 0x0, 0xa, 0x79, 0xf3},
		},
		{
			name: "ok2",
			given: func(r *ReadDiscreteInputsRequestRTU) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x03, 0x4a, 0x96},
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

func TestReadDiscreteInputsRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name         string
		whenQuantity uint16
		expect       int
	}{
		{
			name:         "ok, 1 byte",
			whenQuantity: 8,
			expect:       3 + 1 + 2,
		},
		{
			name:         "ok, 2 bytes",
			whenQuantity: 9,
			expect:       3 + 2 + 2,
		},
		{
			name:         "ok, 11 bytes",
			whenQuantity: 8*10 + 7,
			expect:       3 + 11 + 2,
		},
		{
			name:         "ok, 253 bytes",
			whenQuantity: 8 * 253,
			expect:       3 + 253 + 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadDiscreteInputsRequestRTU{
				ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
					UnitID:       1,
					StartAddress: 200,
					Quantity:     tc.whenQuantity,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestReadDiscreteInputsRequest_FunctionCode(t *testing.T) {
	given := ReadDiscreteInputsRequest{}
	assert.Equal(t, uint8(2), given.FunctionCode())
}

func TestReadDiscreteInputsRequest_Bytes(t *testing.T) {
	example := ReadDiscreteInputsRequest{
		UnitID:       1,
		StartAddress: 200,
		Quantity:     10,
	}

	var testCases = []struct {
		name   string
		given  func(r *ReadDiscreteInputsRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *ReadDiscreteInputsRequest) {},
			expect: []byte{0x1, 0x2, 0x0, 0xc8, 0x0, 0xa},
		},
		{
			name: "ok2",
			given: func(r *ReadDiscreteInputsRequest) {
				r.UnitID = 16
				r.StartAddress = 107
				r.Quantity = 3
			},
			expect: []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x03},
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

func TestParseReadDiscreteInputsRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadDiscreteInputsRequestTCP
		expectError string
	}{
		{
			name: "ok, parse ReadDiscreteInputsRequestTCP",
			when: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x02, 0x00, 0x6B, 0x00, 0x03},
			expect: &ReadDiscreteInputsRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x0102,
					ProtocolID:    0,
				},
				ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
					UnitID:       0x10,
					StartAddress: 0x6b,
					Quantity:     0x03,
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x07, 0x10, 0x02, 0x00, 0x6B, 0x00, 0x03},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expect:      nil,
			expectError: "received function code in packet is not 0x02",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x02, 0x00, 0x6B, 0x00, 0x00},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x02, 0x00, 0x6B, 0x00, 0x7e},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadDiscreteInputsRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadDiscreteInputsRequestRTU(t *testing.T) {
	example := ReadDiscreteInputsRequestRTU{
		ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
			UnitID:       0x10,
			StartAddress: 0x6b,
			Quantity:     0x03,
		},
	}
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadDiscreteInputsRequestRTU
		expectError string
	}{
		{
			name:   "ok, parse ReadDiscreteInputsRequestRTU, with crc bytes",
			when:   []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x03, 0xff, 0xff},
			expect: &example,
		},
		{
			name:   "ok, parse ReadDiscreteInputsRequestRTU, without crc bytes",
			when:   []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x03},
			expect: &example,
		},
		{
			name:        "nok, invalid data length to be valid packet",
			when:        []byte{0x10, 0x02, 0x00, 0x6B, 0x00},
			expect:      nil,
			expectError: "invalid data length to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x10, 0x00, 0x00, 0x6B, 0x00, 0x03, 0xff, 0xff},
			expect:      nil,
			expectError: "received function code in packet is not 0x02",
		},
		{
			name:        "nok, quantity can not be 0",
			when:        []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x0, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
		{
			name:        "nok, quantity can not be 126",
			when:        []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x7e, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid quantity. valid range 1..125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadDiscreteInputsRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
