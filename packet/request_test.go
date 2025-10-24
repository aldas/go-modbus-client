package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "ok, FunctionReadServerID",
			when: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x2, 0x1, 0x11},
			expect: &ReadServerIDRequestTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
				},
				ReadServerIDRequest: ReadServerIDRequest{
					UnitID: 0x1,
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10},
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
			name: "ok, parse ReadServerIDRequestRTU with crc",
			when: []byte{0x10, 0x11, 0xcc, 0x7c},
			expect: &ReadServerIDRequestRTU{
				ReadServerIDRequest: ReadServerIDRequest{
					UnitID: 0x10,
				},
			},
		},
		{
			name:        "nok, too short",
			when:        []byte{0x10, 0x00, 0x6B},
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
			when:        []byte{0x10, 0x01, 0x00},
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

func TestExtractRequestDestination(t *testing.T) {
	var testCases = []struct {
		name      string
		when      Request
		expect    RequestDestination
		expectErr string
	}{
		{
			name: "read coils tcp",
			when: &ReadCoilsRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x0102},
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       0x11,
					StartAddress: 0x0001,
					Quantity:     0x0004,
				},
			},
			expect: RequestDestination{
				UnitID:       0x11,
				FunctionCode: FunctionReadCoils,
				StartAddress: 0x0001,
				Quantity:     0x0004,
			},
		},
		{
			name: "read coils rtu",
			when: &ReadCoilsRequestRTU{
				ReadCoilsRequest: ReadCoilsRequest{
					UnitID:       0x12,
					StartAddress: 0x0010,
					Quantity:     8,
				},
			},
			expect: RequestDestination{
				UnitID:       0x12,
				FunctionCode: FunctionReadCoils,
				StartAddress: 0x0010,
				Quantity:     8,
			},
		},
		{
			name: "read discrete inputs tcp",
			when: &ReadDiscreteInputsRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 1},
				ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
					UnitID:       0x21,
					StartAddress: 0x0015,
					Quantity:     0x0002,
				},
			},
			expect: RequestDestination{
				UnitID:       0x21,
				FunctionCode: FunctionReadDiscreteInputs,
				StartAddress: 0x0015,
				Quantity:     0x0002,
			},
		},
		{
			name: "read discrete inputs rtu",
			when: &ReadDiscreteInputsRequestRTU{
				ReadDiscreteInputsRequest: ReadDiscreteInputsRequest{
					UnitID:       0x22,
					StartAddress: 0x0007,
					Quantity:     3,
				},
			},
			expect: RequestDestination{
				UnitID:       0x22,
				FunctionCode: FunctionReadDiscreteInputs,
				StartAddress: 0x0007,
				Quantity:     3,
			},
		},
		{
			name: "read holding registers tcp",
			when: &ReadHoldingRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x0a0b},
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       0x31,
					StartAddress: 0x0100,
					Quantity:     6,
				},
			},
			expect: RequestDestination{
				UnitID:       0x31,
				FunctionCode: FunctionReadHoldingRegisters,
				StartAddress: 0x0100,
				Quantity:     6,
			},
		},
		{
			name: "read holding registers rtu",
			when: &ReadHoldingRegistersRequestRTU{
				ReadHoldingRegistersRequest: ReadHoldingRegistersRequest{
					UnitID:       0x32,
					StartAddress: 0x0012,
					Quantity:     4,
				},
			},
			expect: RequestDestination{
				UnitID:       0x32,
				FunctionCode: FunctionReadHoldingRegisters,
				StartAddress: 0x0012,
				Quantity:     4,
			},
		},
		{
			name: "read input registers tcp",
			when: &ReadInputRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 7},
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       0x41,
					StartAddress: 0x0200,
					Quantity:     2,
				},
			},
			expect: RequestDestination{
				UnitID:       0x41,
				FunctionCode: FunctionReadInputRegisters,
				StartAddress: 0x0200,
				Quantity:     2,
			},
		},
		{
			name: "read input registers rtu",
			when: &ReadInputRegistersRequestRTU{
				ReadInputRegistersRequest: ReadInputRegistersRequest{
					UnitID:       0x42,
					StartAddress: 0x0300,
					Quantity:     5,
				},
			},
			expect: RequestDestination{
				UnitID:       0x42,
				FunctionCode: FunctionReadInputRegisters,
				StartAddress: 0x0300,
				Quantity:     5,
			},
		},
		{
			name: "write single coil rtu",
			when: &WriteSingleCoilRequestRTU{
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:  0x51,
					Address: 0x0005,
				},
			},
			expect: RequestDestination{
				UnitID:       0x51,
				FunctionCode: FunctionWriteSingleCoil,
				StartAddress: 0x0005,
			},
		},
		{
			name: "write single coil tcp",
			when: &WriteSingleCoilRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x1001},
				WriteSingleCoilRequest: WriteSingleCoilRequest{
					UnitID:  0x52,
					Address: 0x0006,
				},
			},
			expect: RequestDestination{
				UnitID:       0x52,
				FunctionCode: FunctionWriteSingleCoil,
				StartAddress: 0x0006,
			},
		},
		{
			name: "write single register rtu",
			when: &WriteSingleRegisterRequestRTU{
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  0x61,
					Address: 0x0010,
					Data:    [2]byte{0xAA, 0x55},
				},
			},
			expect: RequestDestination{
				UnitID:       0x61,
				FunctionCode: FunctionWriteSingleRegister,
				StartAddress: 0x0010,
			},
		},
		{
			name: "write single register tcp",
			when: &WriteSingleRegisterRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x2020},
				WriteSingleRegisterRequest: WriteSingleRegisterRequest{
					UnitID:  0x62,
					Address: 0x0011,
					Data:    [2]byte{0x01, 0x02},
				},
			},
			expect: RequestDestination{
				UnitID:       0x62,
				FunctionCode: FunctionWriteSingleRegister,
				StartAddress: 0x0011,
			},
		},
		{
			name: "write multiple coils rtu",
			when: &WriteMultipleCoilsRequestRTU{
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       0x71,
					StartAddress: 0x0101,
					CoilCount:    0x0003,
					Data:         []byte{0x07},
				},
			},
			expect: RequestDestination{
				UnitID:       0x71,
				FunctionCode: FunctionWriteMultipleCoils,
				StartAddress: 0x0101,
				Quantity:     0x0003,
			},
		},
		{
			name: "write multiple coils tcp",
			when: &WriteMultipleCoilsRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x3030},
				WriteMultipleCoilsRequest: WriteMultipleCoilsRequest{
					UnitID:       0x72,
					StartAddress: 0x0102,
					CoilCount:    0x0004,
					Data:         []byte{0x0F},
				},
			},
			expect: RequestDestination{
				UnitID:       0x72,
				FunctionCode: FunctionWriteMultipleCoils,
				StartAddress: 0x0102,
				Quantity:     0x0004,
			},
		},
		{
			name: "write multiple registers rtu",
			when: &WriteMultipleRegistersRequestRTU{
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:        0x81,
					StartAddress:  0x0201,
					RegisterCount: 0x0002,
					Data:          []byte{0x00, 0x01, 0x00, 0x02},
				},
			},
			expect: RequestDestination{
				UnitID:       0x81,
				FunctionCode: FunctionWriteMultipleRegisters,
				StartAddress: 0x0201,
				Quantity:     0x0002,
			},
		},
		{
			name: "write multiple registers tcp",
			when: &WriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x4040},
				WriteMultipleRegistersRequest: WriteMultipleRegistersRequest{
					UnitID:        0x82,
					StartAddress:  0x0202,
					RegisterCount: 0x0003,
					Data:          []byte{0x00, 0x03, 0x00, 0x04, 0x00, 0x05},
				},
			},
			expect: RequestDestination{
				UnitID:       0x82,
				FunctionCode: FunctionWriteMultipleRegisters,
				StartAddress: 0x0202,
				Quantity:     0x0003,
			},
		},
		{
			name: "read server id rtu",
			when: &ReadServerIDRequestRTU{
				ReadServerIDRequest: ReadServerIDRequest{
					UnitID: 0x91,
				},
			},
			expect: RequestDestination{
				UnitID:       0x91,
				FunctionCode: FunctionReadServerID,
			},
		},
		{
			name: "read server id tcp",
			when: &ReadServerIDRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x5050},
				ReadServerIDRequest: ReadServerIDRequest{
					UnitID: 0x92,
				},
			},
			expect: RequestDestination{
				UnitID:       0x92,
				FunctionCode: FunctionReadServerID,
			},
		},
		{
			name: "read write multiple registers rtu",
			when: &ReadWriteMultipleRegistersRequestRTU{
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            0xA1,
					ReadStartAddress:  0x0301,
					ReadQuantity:      0x0002,
					WriteStartAddress: 0x0400,
					WriteQuantity:     0x0001,
					WriteData:         []byte{0x00, 0x64},
				},
			},
			expect: RequestDestination{
				UnitID:       0xA1,
				FunctionCode: FunctionReadWriteMultipleRegisters,
				StartAddress: 0x0301,
				Quantity:     0x0002,
			},
		},
		{
			name: "read write multiple registers tcp",
			when: &ReadWriteMultipleRegistersRequestTCP{
				MBAPHeader: MBAPHeader{TransactionID: 0x6060},
				ReadWriteMultipleRegistersRequest: ReadWriteMultipleRegistersRequest{
					UnitID:            0xA2,
					ReadStartAddress:  0x0302,
					ReadQuantity:      0x0003,
					WriteStartAddress: 0x0401,
					WriteQuantity:     0x0002,
					WriteData:         []byte{0x00, 0x65, 0x00, 0x66},
				},
			},
			expect: RequestDestination{
				UnitID:       0xA2,
				FunctionCode: FunctionReadWriteMultipleRegisters,
				StartAddress: 0x0302,
				Quantity:     0x0003,
			},
		},
		{
			name:      "unsupported request type",
			when:      dummyRequest{},
			expectErr: "extract request destination: unknown function code parsed: 153",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ExtractRequestDestination(tc.when)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
				assert.Equal(t, RequestDestination{}, actual)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, actual)
		})
	}
}

type dummyRequest struct{}

func (dummyRequest) FunctionCode() uint8         { return 0x99 }
func (dummyRequest) Bytes() []byte               { return nil }
func (dummyRequest) ExpectedResponseLength() int { return 0 }
