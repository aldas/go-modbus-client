package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReadWriteMultipleRegistersRequestTCP(t *testing.T) {
	var testCases = []struct {
		name                  string
		whenUnitID            uint8
		whenReadStartAddress  uint16
		whenReadQuantity      uint16
		whenWriteStartAddress uint16
		whenWriteQuantity     uint16
		whenWriteData         []byte
		expect                *ReadWriteMultipleRegistersRequestTCP
		expectError           string
	}{
		{
			name:                  "ok, write 1 register",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      1,
			whenWriteStartAddress: 16,
			whenWriteQuantity:     1,
			whenWriteData:         []byte{0x1, 0x2},
			expect: &ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x1234, ProtocolID: 0},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      1,
					WriteStartAddress: 16,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2},
				},
			},
			expectError: "",
		},
		{
			name:                  "ok, write 2 registers",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      1,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         []byte{0x1, 0x2, 0x3, 0x4},
			expect: &ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x1234, ProtocolID: 0},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      1,
					WriteStartAddress: 100,
					WriteQuantity:     2,
					WriteData:         []byte{0x1, 0x2, 0x3, 0x4},
				},
			},
		},
		{
			name:                  "nok, read quantity too large",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      125,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         []byte{0x1, 0x2, 0x3, 0x4},
			expect:                nil,
			expectError:           "read registers count out of range (1-124): 125",
		},
		{
			name:                  "nok, write data not even number of bytes",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      2,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         []byte{0x1, 0x2, 0x3},
			expect:                nil,
			expectError:           "write data length must be even number of bytes",
		},
		{
			name:                  "nok, write data too long",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      2,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         make([]byte, 125*2),
			expect:                nil,
			expectError:           "write registers count out of range (1-124): 125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadWriteMultipleRegistersRequestTCP(
				tc.whenUnitID,
				tc.whenReadStartAddress,
				tc.whenReadQuantity,
				tc.whenWriteStartAddress,
				tc.whenWriteData,
			)

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

func TestReadWriteMultipleRegistersRequestTCP_Bytes(t *testing.T) {
	var testCases = []struct {
		name   string
		given  ReadWriteMultipleRegistersRequestTCP
		expect []byte
	}{
		{
			name: "ok, write 1 register",
			given: ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x1234, ProtocolID: 0},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      1,
					WriteStartAddress: 100,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2},
				},
			},
			expect: []byte{
				0x12, 0x34, 0x0, 0x0, 0x0, 0xd, // MBAP header
				0x1, 0x17, 0x0, 0xc8, 0x0, 0x1, 0x0, 0x64, 0x0, 0x1, 0x2, 0x1, 0x2,
			},
		},
		{
			name: "ok, write 2 registers",
			given: ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x1234, ProtocolID: 0},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      1,
					WriteStartAddress: 100,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2, 0x3, 0x4},
				},
			},
			expect: []byte{
				0x12, 0x34, 0x0, 0x0, 0x0, 0xf,
				0x1, 0x17, 0x0, 0xc8, 0x0, 0x1, 0x0, 0x64, 0x0, 0x1, 0x4, 0x1, 0x2, 0x3, 0x4,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.given.Bytes())
		})
	}
}

func TestReadWriteMultipleRegistersRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name          string
		whenCoilState bool
		expect        int
	}{
		{
			name:          "ok",
			whenCoilState: true,
			expect:        17 + 3*2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x1234, ProtocolID: 0},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      3,
					WriteStartAddress: 100,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2},
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewReadWriteMultipleRegistersRequestRTU(t *testing.T) {
	var testCases = []struct {
		name                  string
		whenUnitID            uint8
		whenReadStartAddress  uint16
		whenReadQuantity      uint16
		whenWriteStartAddress uint16
		whenWriteQuantity     uint16
		whenWriteData         []byte
		expect                *ReadWriteMultipleRegistersRequestRTU
		expectError           string
	}{
		{
			name:                  "ok, write 1 register",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      1,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     1,
			whenWriteData:         []byte{0x1, 0x2},
			expect: &ReadWriteMultipleRegistersRequestRTU{
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      1,
					WriteStartAddress: 100,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2},
				},
			},
			expectError: "",
		},
		{
			name:                  "ok, write 2 registers",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      3,
			whenWriteStartAddress: 16,
			whenWriteQuantity:     2,
			whenWriteData:         []byte{0x1, 0x2, 0x3, 0x4},
			expect: &ReadWriteMultipleRegistersRequestRTU{
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      3,
					WriteStartAddress: 16,
					WriteQuantity:     2,
					WriteData:         []byte{0x1, 0x2, 0x3, 0x4},
				},
			},
		},
		{
			name:                  "nok, read quantity too large",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      125,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         []byte{0x1, 0x2, 0x3, 0x4},
			expect:                nil,
			expectError:           "read registers count out of range (1-124): 125",
		},
		{
			name:                  "nok, write data not even number of bytes",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      2,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         []byte{0x1, 0x2, 0x3},
			expect:                nil,
			expectError:           "write data length must be even number of bytes",
		},
		{
			name:                  "nok, write data too long",
			whenUnitID:            1,
			whenReadStartAddress:  200,
			whenReadQuantity:      2,
			whenWriteStartAddress: 100,
			whenWriteQuantity:     2,
			whenWriteData:         make([]byte, 125*2),
			expect:                nil,
			expectError:           "write registers count out of range (1-124): 125",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewReadWriteMultipleRegistersRequestRTU(
				tc.whenUnitID,
				tc.whenReadStartAddress,
				tc.whenReadQuantity,
				tc.whenWriteStartAddress,
				tc.whenWriteData,
			)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadWriteMultipleRegistersRequestRTU_Bytes(t *testing.T) {
	var testCases = []struct {
		name   string
		given  ReadWriteMultipleRegistersRequestRTU
		expect []byte
	}{
		{
			name: "ok",
			given: ReadWriteMultipleRegistersRequestRTU{
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      2,
					WriteStartAddress: 100,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2},
				},
			},
			expect: []byte{0x1, 0x17, 0x0, 0xc8, 0x0, 0x2, 0x0, 0x64, 0x0, 0x1, 0x2, 0x1, 0x2, 0x18, 0x18},
		},
		{
			name: "ok2",
			given: ReadWriteMultipleRegistersRequestRTU{
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      1,
					WriteStartAddress: 100,
					WriteQuantity:     2,
					WriteData:         []byte{0x1, 0x2, 0x3, 0x4},
				},
			},
			expect: []byte{0x1, 0x17, 0x0, 0xc8, 0x0, 0x1, 0x0, 0x64, 0x0, 0x2, 0x4, 0x1, 0x2, 0x3, 0x4, 0x73, 0x5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.given.Bytes())
		})
	}
}

func TestReadWriteMultipleRegistersRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name          string
		whenCoilState uint16
		expect        int
	}{
		{
			name:          "ok",
			whenCoilState: 8,
			expect:        6 + 2*2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := ReadWriteMultipleRegistersRequestRTU{
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            1,
					ReadStartAddress:  200,
					ReadQuantity:      2,
					WriteStartAddress: 100,
					WriteQuantity:     1,
					WriteData:         []byte{0x1, 0x2},
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestReadWriteMultipleRegistersRequest_FunctionCode(t *testing.T) {
	given := ReadWriteMultipleRegistersRequest{}
	assert.Equal(t, uint8(23), given.FunctionCode())
}

func TestReadWriteMultipleRegistersRequest_Bytes(t *testing.T) {
	var testCases = []struct {
		name   string
		given  ReadWriteMultipleRegistersRequest
		expect []byte
	}{
		{
			name: "ok",
			given: ReadWriteMultipleRegistersRequest{
				UnitID:            2,
				ReadStartAddress:  200,
				ReadQuantity:      3,
				WriteStartAddress: 100,
				WriteQuantity:     1,
				WriteData:         []byte{0x1, 0x2},
			},
			expect: []byte{0x2, 0x17, 0x0, 0xc8, 0x0, 0x3, 0x0, 0x64, 0x0, 0x1, 0x2, 0x1, 0x2},
		},
		{
			name: "ok2",
			given: ReadWriteMultipleRegistersRequest{
				UnitID:            1,
				ReadStartAddress:  200,
				ReadQuantity:      2,
				WriteStartAddress: 100,
				WriteQuantity:     2,
				WriteData:         []byte{0x1, 0x2, 0x3, 0x4},
			},
			expect: []byte{0x1, 0x17, 0x0, 0xc8, 0x0, 0x2, 0x0, 0x64, 0x0, 0x2, 0x4, 0x1, 0x2, 0x3, 0x4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.given.Bytes())
		})
	}
}

func TestParseReadWriteMultipleRegistersRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadWriteMultipleRegistersRequestTCP
		expectError string
	}{
		{
			name: "ok, parse ReadWriteMultipleRegistersRequestTCP",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect: &ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x0138,
					ProtocolID:    0,
				},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            0x11,
					ReadStartAddress:  0x0410,
					ReadQuantity:      0x01,
					WriteStartAddress: 0x0112,
					WriteQuantity:     0x02,
					WriteData:         []byte{0x00, 0xc8, 0x00, 0x82},
				},
			},
		},
		{
			name: "nok, invalid header",
			when: []byte{
				0x01, 0x38, 0x00, 0x01, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "invalid protocol id",
		},
		{
			name: "nok, invalid function code",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0xff, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "received function code in packet is not 0x17",
		},
		{
			name: "nok, read quantity can not be 0",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x00, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "invalid read quantity. valid range 1..125",
		},
		{
			name: "nok, read quantity can not be 126",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x7e, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "invalid read quantity. valid range 1..125",
		},
		{
			name: "nok, write quantity can not be 0",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x00, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "invalid write quantity. valid range 1..121",
		},
		{
			name: "nok, write quantity can not be 122",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x7a, 0x04, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "invalid write quantity. valid range 1..121",
		},
		{
			name: "nok, invalid write byte count",
			when: []byte{
				0x01, 0x38, 0x00, 0x00, 0x00, 0x0f,
				0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x05, 0x00, 0xc8, 0x00, 0x82,
			},
			expect:      nil,
			expectError: "received data write bytes length does not match write data length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadWriteMultipleRegistersRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseReadWriteMultipleRegistersRequestRTU(t *testing.T) {
	example := &ReadWriteMultipleRegistersRequestRTU{
		ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
			UnitID:            0x11,
			ReadStartAddress:  0x0410,
			ReadQuantity:      0x01,
			WriteStartAddress: 0x0112,
			WriteQuantity:     0x02,
			WriteData:         []byte{0x00, 0xc8, 0x00, 0x82},
		},
	}
	var testCases = []struct {
		name        string
		when        []byte
		expect      *ReadWriteMultipleRegistersRequestRTU
		expectError string
	}{
		{
			name:   "ok, parse ReadWriteMultipleRegistersRequestRTU with crc",
			when:   []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect: example,
		},
		{
			name:   "ok, parse ReadWriteMultipleRegistersRequestRTU without crc",
			when:   []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82},
			expect: example,
		},
		{
			name:        "nok, too short",
			when:        []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04},
			expect:      nil,
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x11, 0x00, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect:      nil,
			expectError: "received function code in packet is not 0x17",
		},
		{
			name:        "nok, read quantity can not be 0",
			when:        []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x00, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid read quantity. valid range 1..125",
		},
		{
			name:        "nok, read quantity can not be 126",
			when:        []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x7e, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid read quantity. valid range 1..125",
		},
		{
			name:        "nok, write quantity can not be 0",
			when:        []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x00, 0x04, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid write quantity. valid range 1..121",
		},
		{
			name:        "nok, write quantity can not be 122",
			when:        []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x7a, 0x04, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid write quantity. valid range 1..121",
		},
		{
			name:        "nok, invalid write byte count",
			when:        []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x03, 0x00, 0xc8, 0x00, 0x82, 0xff, 0xff},
			expect:      nil,
			expectError: "received data write bytes length does not match write data length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseReadWriteMultipleRegistersRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
