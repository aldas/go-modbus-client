package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWriteMultipleRegistersRequestTCP(t *testing.T) {
	expect := WriteMultipleRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
			UnitID:        1,
			StartAddress:  200,
			RegisterCount: 1,
			Data:          []byte{0x01, 0x02},
		},
	}

	var testCases = []struct {
		name             string
		whenUnitID       uint8
		whenStartAddress uint16
		whenData         []byte
		expect           *WriteMultipleRegistersRequestTCP
		expectError      string
	}{
		{
			name:             "ok",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenData:         []byte{0x01, 0x02},
			expect:           &expect,
			expectError:      "",
		},
		{
			name:             "nok, registers count too small",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenData:         []byte{},
			expect:           nil,
			expectError:      "registers count out of range (1-124): 0",
		},
		{
			name:             "nok, registers data not even number of bytes",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenData:         []byte{0x1},
			expect:           nil,
			expectError:      "data length must be even number of bytes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteMultipleRegistersRequestTCP(tc.whenUnitID, tc.whenStartAddress, tc.whenData)

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

func TestWriteMultipleRegistersRequestTCP_Bytes(t *testing.T) {
	example := WriteMultipleRegistersRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
			UnitID:        1,
			StartAddress:  200,
			RegisterCount: 1,
			Data:          []byte{0x01, 0x02},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleRegistersRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleRegistersRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x9, 0x1, 0x10, 0x0, 0xc8, 0x0, 0x1, 0x2, 0x1, 0x2},
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

func TestWriteMultipleRegistersRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name     string
		whenData []byte
		expect   int
	}{
		{
			name:     "ok, 2 byte",
			whenData: []byte{0x1, 0x2},
			expect:   12,
		},
		{
			name:     "ok, 4 bytes",
			whenData: []byte{0x1, 0x2, 0x3, 0x4},
			expect:   12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:       1,
					StartAddress: 200,
					Data:         tc.whenData,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewWriteMultipleRegistersRequestRTU(t *testing.T) {
	expect := WriteMultipleRegistersRequestRTU{
		WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
			UnitID:        1,
			StartAddress:  200,
			RegisterCount: 1,
			Data:          []byte{0x01, 0x02},
		},
	}

	var testCases = []struct {
		name             string
		whenUnitID       uint8
		whenStartAddress uint16
		whenData         []byte
		expect           *WriteMultipleRegistersRequestRTU
		expectError      string
	}{
		{
			name:             "ok",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenData:         []byte{0x01, 0x02},
			expect:           &expect,
			expectError:      "",
		},
		{
			name:             "nok, registers count too small",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenData:         []byte{},
			expect:           nil,
			expectError:      "registers count out of range (1-124): 0",
		},
		{
			name:             "nok, registers data not even number of bytes",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenData:         []byte{0x1},
			expect:           nil,
			expectError:      "data length must be even number of bytes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteMultipleRegistersRequestRTU(tc.whenUnitID, tc.whenStartAddress, tc.whenData)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteMultipleRegistersRequestRTU_Bytes(t *testing.T) {
	example := WriteMultipleRegistersRequestRTU{
		WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
			UnitID:        1,
			StartAddress:  200,
			RegisterCount: 1,
			Data:          []byte{0x01, 0x02},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleRegistersRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleRegistersRequestRTU) {},
			expect: []byte{0x1, 0x10, 0x0, 0xc8, 0x0, 0x1, 0x2, 0x1, 0x2, 0x36, 0x49},
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

func TestWriteMultipleRegistersRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name     string
		whenData []byte
		expect   int
	}{
		{
			name:     "ok, 1 byte",
			whenData: []byte{0x1, 0x2},
			expect:   8,
		},
		{
			name:     "ok, 4 bytes",
			whenData: []byte{0x1, 0x2, 0x3, 0x4},
			expect:   8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteMultipleRegistersRequestRTU{
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:       1,
					StartAddress: 200,
					Data:         tc.whenData,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestWriteMultipleRegistersRequest_FunctionCode(t *testing.T) {
	given := WriteMultipleRegistersRequest{}
	assert.Equal(t, uint8(16), given.FunctionCode())
}

func TestWriteMultipleRegistersRequest_Bytes(t *testing.T) {
	example := WriteMultipleRegistersRequest{
		UnitID:        1,
		StartAddress:  200,
		RegisterCount: 1,
		Data:          []byte{0x01, 0x02},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleRegistersRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleRegistersRequest) {},
			expect: []byte{0x1, 0x10, 0x0, 0xc8, 0x0, 0x1, 0x2, 0x1, 0x2},
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

func TestParseWriteMultipleRegistersRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *WriteMultipleRegistersRequestTCP
		expectError string
	}{
		{
			name: "ok, parse WriteMultipleRegistersRequestTCP",
			when: []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x0d, 0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect: &WriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x0138,
					ProtocolID:    0,
				},
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:        0x11,
					StartAddress:  0x0410,
					RegisterCount: 0x03,
					Data:          []byte{0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x0e, 0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x0d, 0x11, 0x01, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect:      nil,
			expectError: "received function code in packet is not 0x10",
		},
		{
			name:        "nok, register count can not be 0",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x0d, 0x11, 0x10, 0x04, 0x10, 0x00, 0x00, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect:      nil,
			expectError: "invalid register count. valid range 1..123",
		},
		{
			name:        "nok, register count can not be 126",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x0d, 0x11, 0x10, 0x04, 0x10, 0x00, 0x7c, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect:      nil,
			expectError: "invalid register count. valid range 1..123",
		},
		{
			name:        "nok, invalid register byte count",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x0d, 0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x05, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect:      nil,
			expectError: "received data register bytes length does not match write data length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWriteMultipleRegistersRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteMultipleRegistersRequestRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *WriteMultipleRegistersRequestRTU
		expectError string
	}{
		{
			name: "ok, parse WriteMultipleRegistersRequestRTU with crc",
			when: []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01, 0xff, 0xff},
			expect: &WriteMultipleRegistersRequestRTU{
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:        0x11,
					StartAddress:  0x0410,
					RegisterCount: 0x03,
					Data:          []byte{0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
				},
			},
		},
		{
			name: "ok, parse WriteMultipleRegistersRequestRTU without crc",
			when: []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
			expect: &WriteMultipleRegistersRequestRTU{
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:        0x11,
					StartAddress:  0x0410,
					RegisterCount: 0x03,
					Data:          []byte{0x00, 0xC8, 0x00, 0x82, 0x87, 0x01},
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x06},
			expect:      nil,
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x11, 0x00, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01, 0xff, 0xff},
			expect:      nil,
			expectError: "received function code in packet is not 0x10",
		},
		{
			name:        "nok, register count can not be 0",
			when:        []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x00, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid register count. valid range 1..123",
		},
		{
			name:        "nok, register count can not be 123",
			when:        []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x7c, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01, 0xff, 0xff},
			expect:      nil,
			expectError: "invalid register count. valid range 1..123",
		},
		{
			name:        "nok, invalid register byte count",
			when:        []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x05, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01, 0xff, 0xff},
			expect:      nil,
			expectError: "received data register bytes length does not match write data length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWriteMultipleRegistersRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
