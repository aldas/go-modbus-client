package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToCoilsBytes(t *testing.T) {
	var testCases = []struct {
		name      string
		whenCoils []bool
		expect    []byte
	}{
		{
			name:      "ok, 7 coils = 1 byte slice",
			whenCoils: []bool{true, false, true, false, true, false, true},
			expect:    []byte{0b01010101}, // 85
		},
		{
			name: "ok, 20 coils = 3 byte slice",
			whenCoils: []bool{
				true, false, true, false, true, false, true, false, // 8
				false, false, false, false, false, false, false, false, // 16
				true, false, false, true, // 20
			},
			expect: []byte{0b01010101, 0b0, 0b1001}, // 85, 0, 9
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CoilsToBytes(tc.whenCoils)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestNewWriteMultipleCoilsRequestTCP(t *testing.T) {
	expect := WriteMultipleCoilsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			CoilCount:    7,
			Data:         []byte{0b01010101},
		},
	}

	var testCases = []struct {
		name             string
		whenUnitID       uint8
		whenStartAddress uint16
		whenCoils        []bool
		expect           *WriteMultipleCoilsRequestTCP
		expectError      string
	}{
		{
			name:             "ok",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenCoils:        []bool{true, false, true, false, true, false, true},
			expect:           &expect,
			expectError:      "",
		},
		{
			name:             "nok, quantity too small",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenCoils:        []bool{},
			expect:           nil,
			expectError:      "coils count is out of range (1-1968): 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteMultipleCoilsRequestTCP(tc.whenUnitID, tc.whenStartAddress, tc.whenCoils)

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

func TestWriteMultipleCoilsRequestTCP_Bytes(t *testing.T) {
	example := WriteMultipleCoilsRequestTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			CoilCount:    7,
			Data:         []byte{0b01010101},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleCoilsRequestTCP)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleCoilsRequestTCP) {},
			expect: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x8, 0x1, 0xf, 0x0, 0xc8, 0x0, 0x7, 0x1, 0x55},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleCoilsRequestTCP) {
				r.TransactionID = 1

				r.UnitID = 16
				r.StartAddress = 107
				r.CoilCount = 20
				r.Data = []byte{0b01010101, 0b0, 0b1001}
			},
			expect: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0xa, 0x10, 0xf, 0x0, 0x6b, 0x0, 0x14, 0x3, 0x55, 0x0, 0x9},
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

func TestWriteMultipleCoilsRequestTCP_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name     string
		whenData []byte
		expect   int
	}{
		{
			name:     "ok, 1 byte",
			whenData: []byte{0b01010101},
			expect:   12,
		},
		{
			name:     "ok, 3 bytes",
			whenData: []byte{0b01010101, 0b0, 0b1001},
			expect:   12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteMultipleCoilsRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       1,
					StartAddress: 200,
					Data:         tc.whenData,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestNewWriteMultipleCoilsRequestRTU(t *testing.T) {
	expect := WriteMultipleCoilsRequestRTU{
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			CoilCount:    7,
			Data:         []byte{0b01010101},
		},
	}

	var testCases = []struct {
		name             string
		whenUnitID       uint8
		whenStartAddress uint16
		whenCoils        []bool
		expect           *WriteMultipleCoilsRequestRTU
		expectError      string
	}{
		{
			name:             "ok",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenCoils:        []bool{true, false, true, false, true, false, true},
			expect:           &expect,
			expectError:      "",
		},
		{
			name:             "nok, quantity too small",
			whenUnitID:       1,
			whenStartAddress: 200,
			whenCoils:        []bool{},
			expect:           nil,
			expectError:      "coils count is out of range (1-1968): 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := NewWriteMultipleCoilsRequestRTU(tc.whenUnitID, tc.whenStartAddress, tc.whenCoils)

			assert.Equal(t, tc.expect, packet)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteMultipleCoilsRequestRTU_Bytes(t *testing.T) {
	example := WriteMultipleCoilsRequestRTU{
		WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			CoilCount:    7,
			Data:         []byte{0b01010101},
		},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleCoilsRequestRTU)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleCoilsRequestRTU) {},
			expect: []byte{0x1, 0xf, 0x0, 0xc8, 0x0, 0x7, 0x1, 0x55, 0xef, 0x79},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleCoilsRequestRTU) {
				r.UnitID = 16
				r.StartAddress = 107
				r.CoilCount = 20
				r.Data = []byte{0b01010101, 0b0, 0b1001}
			},
			expect: []byte{0x10, 0xf, 0x0, 0x6b, 0x0, 0x14, 0x3, 0x55, 0x0, 0x9, 0xa, 0xf5},
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

func TestWriteMultipleCoilsRequestRTU_ExpectedResponseLength(t *testing.T) {
	var testCases = []struct {
		name     string
		whenData []byte
		expect   int
	}{
		{
			name:     "ok, 1 byte",
			whenData: []byte{0b01010101},
			expect:   8,
		},
		{
			name:     "ok, 3 bytes",
			whenData: []byte{0b01010101, 0b0, 0b1001},
			expect:   8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			example := WriteMultipleCoilsRequestRTU{
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       1,
					StartAddress: 200,
					Data:         tc.whenData,
				},
			}

			assert.Equal(t, tc.expect, example.ExpectedResponseLength())
		})
	}
}

func TestWriteMultipleCoilsRequest_FunctionCode(t *testing.T) {
	given := WriteMultipleCoilsRequest{}
	assert.Equal(t, uint8(15), given.FunctionCode())
}

func TestWriteMultipleCoilsRequest_Bytes(t *testing.T) {
	example := WriteMultipleCoilsRequest{
		UnitID:       1,
		StartAddress: 200,
		CoilCount:    7,
		Data:         []byte{0b01010101},
	}

	var testCases = []struct {
		name   string
		given  func(r *WriteMultipleCoilsRequest)
		expect []byte
	}{
		{
			name:   "ok",
			given:  func(r *WriteMultipleCoilsRequest) {},
			expect: []byte{0x1, 0xf, 0x0, 0xc8, 0x0, 0x7, 0x1, 0x55},
		},
		{
			name: "ok2",
			given: func(r *WriteMultipleCoilsRequest) {
				r.UnitID = 16
				r.StartAddress = 107
				r.CoilCount = 20
				r.Data = []byte{0b01010101, 0b0, 0b1001}
			},
			expect: []byte{0x10, 0xf, 0x0, 0x6b, 0x0, 0x14, 0x3, 0x55, 0x0, 0x9},
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

func TestParseWriteMultipleCoilsRequestTCP(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *WriteMultipleCoilsRequestTCP
		expectError string
	}{
		{
			name: "ok, parse WriteMultipleCoilsRequestTCP",
			when: []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x08, 0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05},
			expect: &WriteMultipleCoilsRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x0138,
					ProtocolID:    0,
				},
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       0x11,
					StartAddress: 0x0410,
					CoilCount:    0x03,
					Data:         []byte{0x05},
				},
			},
		},
		{
			name:        "nok, invalid header",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x09, 0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05},
			expect:      nil,
			expectError: "packet length does not match length in header",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x08, 0x11, 0x01, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05},
			expect:      nil,
			expectError: "received function code in packet is not 0x0f",
		},
		{
			name:        "nok, coils count can not be 0",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x08, 0x11, 0x0F, 0x04, 0x10, 0x00, 0x00, 0x01, 0x05},
			expect:      nil,
			expectError: "invalid coils count. valid range 1..1968",
		},
		{
			name:        "nok, coils count can not be 1969",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x08, 0x11, 0x0F, 0x04, 0x10, 0x07, 0xb1, 0x01, 0x05},
			expect:      nil,
			expectError: "invalid coils count. valid range 1..1968",
		},
		{
			name:        "nok, invalid coils byte count",
			when:        []byte{0x01, 0x38, 0x00, 0x00, 0x00, 0x08, 0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x02, 0x05},
			expect:      nil,
			expectError: "received data coils bytes length does not match write data length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWriteMultipleCoilsRequestTCP(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseWriteMultipleCoilsRequestRTU(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      *WriteMultipleCoilsRequestRTU
		expectError string
	}{
		{
			name: "ok, parse WriteMultipleCoilsRequestRTU with crc",
			when: []byte{0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05, 0xFF, 0xFF},
			expect: &WriteMultipleCoilsRequestRTU{
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       0x11,
					StartAddress: 0x0410,
					CoilCount:    0x03,
					Data:         []byte{0x05},
				},
			},
		},
		{
			name: "ok, parse WriteMultipleCoilsRequestRTU without crc",
			when: []byte{0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05},
			expect: &WriteMultipleCoilsRequestRTU{
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       0x11,
					StartAddress: 0x0410,
					CoilCount:    0x03,
					Data:         []byte{0x05},
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x11, 0x0F, 0x04, 0x10, 0x00, 0x03},
			expect:      nil,
			expectError: "received data length too short to be valid packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x11, 0x00, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05, 0xFF, 0xFF},
			expect:      nil,
			expectError: "received function code in packet is not 0x0f",
		},
		{
			name:        "nok, coils count can not be 0",
			when:        []byte{0x11, 0x0F, 0x04, 0x10, 0x00, 0x00, 0x01, 0x05, 0xFF, 0xFF},
			expect:      nil,
			expectError: "invalid coils count. valid range 1..1968",
		},
		{
			name:        "nok, coils count can not be 1969",
			when:        []byte{0x11, 0x0F, 0x04, 0x10, 0x0f, 0xb1, 0x01, 0x05, 0xFF, 0xFF},
			expect:      nil,
			expectError: "invalid coils count. valid range 1..1968",
		},
		{
			name:        "nok, invalid coils byte count",
			when:        []byte{0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x02, 0x05, 0xFF, 0xFF},
			expect:      nil,
			expectError: "received data coils bytes length does not match write data length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWriteMultipleCoilsRequestRTU(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
