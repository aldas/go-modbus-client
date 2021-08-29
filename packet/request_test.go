package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTCPRequest(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      interface{}
		expectError string
	}{
		{
			name: "ok, FunctionReadCoils",
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
			name: "ok, FunctionReadDiscreteInputs",
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
			name: "ok, FunctionReadHoldingRegisters",
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
			name: "ok, FunctionReadInputRegisters",
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
			name: "ok, FunctionWriteSingleCoil",
			when: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x11, 0x05, 0x00, 0x6B, 0xFF, 0x00},
			expect: &WriteSingleCoilRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x01,
					ProtocolID:    0,
				},
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:    0x11,
					Address:   0x6b,
					CoilState: true,
				},
			},
		},
		{
			name: "ok, FunctionWriteSingleRegister",
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
			name: "ok, FunctionWriteMultipleCoils",
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
			name: "ok, FunctionWriteMultipleRegisters",
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
			name: "ok, FunctionReadWriteMultipleRegisters",
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
			name: "ok, FunctionReadWriteMultipleRegisters",
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
			name:        "nok, too short",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01},
			expect:      nil,
			expectError: "data is too short to be a Modbus TCP packet",
		},
		{
			name:        "nok, unknown function code",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x00, 0x00, 0x6B, 0x00, 0x03},
			expect:      nil,
			expectError: "unknown function code parsed: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseTCPRequest(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseRTURequestWithCRC(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      interface{}
		expectError string
	}{
		{
			name: "ok, parse ReadCoilsRequestTCP, with crc bytes",
			when: []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03, 0x0e, 0x96},
			expect: &ReadCoilsRequestRTU{
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       0x10,
					StartAddress: 0x6b,
					Quantity:     0x03,
				},
			},
		},
		{
			name: "ok, parse ReadDiscreteInputsRequestRTU, with crc bytes",
			when: []byte{0x10, 0x02, 0x00, 0x6B, 0x00, 0x03, 0x4a, 0x96},
			expect: &ReadDiscreteInputsRequestRTU{
				ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
					UnitID:       0x10,
					StartAddress: 0x6b,
					Quantity:     0x03,
				},
			},
		},
		{
			name: "ok, parse ReadHoldingRegistersRequestRTU with crc bytes",
			when: []byte{0x01, 0x03, 0x00, 0x6B, 0x00, 0x01, 0xf5, 0xd6},
			expect: &ReadHoldingRegistersRequestRTU{
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       0x1,
					StartAddress: 0x6b,
					Quantity:     0x01,
				},
			},
		},
		{
			name: "ok, parse ReadHoldingRegistersRequestRTU with crc bytes",
			when: []byte{0x01, 0x04, 0x00, 0x6B, 0x00, 0x01, 0x40, 0x16},
			expect: &ReadInputRegistersRequestRTU{
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       0x1,
					StartAddress: 0x6b,
					Quantity:     0x01,
				},
			},
		},
		{
			name: "ok, parse WriteSingleCoilRequestRTU with crc bytes",
			when: []byte{0x11, 0x05, 0x00, 0x6B, 0xFF, 0x00, 0xff, 0x76},
			expect: &WriteSingleCoilRequestRTU{
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:    0x11,
					Address:   0x6b,
					CoilState: true,
				},
			},
		},
		{
			name: "ok, parse WriteSingleRegisterRequestRTU with crc bytes",
			when: []byte{0x11, 0x06, 0x00, 0x6B, 0x01, 0x02, 0x7a, 0xd7},
			expect: &WriteSingleRegisterRequestRTU{
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  0x11,
					Address: 0x6b,
					Data:    [2]byte{0x01, 0x02},
				},
			},
		},
		{
			name: "ok, parse WriteMultipleCoilsRequestRTU with crc bytes",
			when: []byte{0x11, 0x0F, 0x04, 0x10, 0x00, 0x03, 0x01, 0x05, 0x8e, 0x1F},
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
			name: "ok, parse WriteMultipleRegistersRequestRTU with crc bytes",
			when: []byte{0x11, 0x10, 0x04, 0x10, 0x00, 0x03, 0x06, 0x00, 0xC8, 0x00, 0x82, 0x87, 0x01, 0x2f, 0x7d},
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
			name: "ok, parse ReadWriteMultipleRegistersRequestRTU with crc",
			when: []byte{0x11, 0x17, 0x04, 0x10, 0x00, 0x01, 0x01, 0x12, 0x00, 0x02, 0x04, 0x00, 0xc8, 0x00, 0x82, 0x64, 0xe2},
			expect: &ReadWriteMultipleRegistersRequestRTU{
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
			name:        "nok, too short",
			when:        []byte{0x10, 0x01, 0x00, 0x6B},
			expect:      nil,
			expectError: "data is too short to be a Modbus RTU packet",
		},
		{
			name:        "nok, invalid CRC",
			when:        []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03, 0xff, 0xff}, // correct CRC is 0x0e, 0x96
			expect:      nil,
			expectError: "packet cyclic redundancy check does not match Modbus RTU packet bytes",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x10, 0x00, 0x00, 0x6B, 0x00, 0x03, 0x33, 0x56},
			expect:      nil,
			expectError: "unknown function code parsed: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseRTURequestWithCRC(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseRTURequest(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      interface{}
		expectError string
	}{
		{
			name: "ok, parse ReadCoilsRequestRTU, with crc bytes",
			when: []byte{0x10, 0x01, 0x00, 0x6B, 0x00, 0x03, 0x0e, 0x96},
			expect: &ReadCoilsRequestRTU{
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       0x10,
					StartAddress: 0x6b,
					Quantity:     0x03,
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x10, 0x01, 0x00, 0x6B},
			expect:      nil,
			expectError: "data is too short to be a Modbus RTU packet",
		},
		{
			name:        "nok, invalid function code",
			when:        []byte{0x10, 0x00, 0x00, 0x6B, 0x00, 0x03, 0x33, 0x56},
			expect:      nil,
			expectError: "unknown function code parsed: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseRTURequest(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
