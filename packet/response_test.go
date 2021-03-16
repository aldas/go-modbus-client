package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTCPResponse(t *testing.T) {
	var testCases = []struct {
		name        string
		whenData    []byte
		expect      Response
		expectError string
	}{
		{
			name:     "ok, ReadCoilsResponseTCP (fc01)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1},
			expect: &ReadCoilsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        5,
				},
				ReadCoilsResponse: ReadCoilsResponse{
					UnitID: 1,
					// +1 function code
					CoilsByteLength: 2,
					Data:            []byte{0x0, 0x1},
				},
			},
		},
		{
			name:     "ok, ReadDiscreteInputsResponseTCP (fc02)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x2, 0x2, 0x0, 0x1},
			expect: &ReadDiscreteInputsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        5,
				},
				ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
					UnitID: 1,
					// +1 function code
					InputsByteLength: 2,
					Data:             []byte{0x0, 0x1},
				},
			},
		},
		{
			name:     "ok, ReadHoldingRegistersResponseTCP (fc03)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x3, 0x2, 0x0, 0x1},
			expect: &ReadHoldingRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        5,
				},
				ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
					UnitID: 1,
					// +1 function code
					RegisterByteLen: 2,
					Data:            []byte{0x0, 0x1},
				},
			},
		},
		{
			name:     "ok, ReadInputRegistersResponseTCP (fc04)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x4, 0x2, 0x0, 0x1},
			expect: &ReadInputRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        5,
				},
				ReadInputRegistersResponse: ReadInputRegistersResponse{
					UnitID: 1,
					// +1 function code
					RegisterByteLen: 2,
					Data:            []byte{0x0, 0x1},
				},
			},
		},
		{
			name:     "ok, WriteSingleCoilResponseTCP (fc05)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x5, 0x0, 0x2, 0xff, 0x0},
			expect: &WriteSingleCoilResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        6,
				},
				WriteSingleCoilResponse: WriteSingleCoilResponse{
					UnitID: 1,
					// +1 function code
					StartAddress: 2,
					CoilState:    true,
				},
			},
		},
		{
			name:     "ok, WriteSingleRegisterResponseTCP (fc06)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x6, 0x0, 0x2, 0x1, 0x2},
			expect: &WriteSingleRegisterResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        6,
				},
				WriteSingleRegisterResponse: WriteSingleRegisterResponse{
					UnitID: 1,
					// +1 function code
					Address: 2,
					Data:    [2]byte{0x1, 0x2},
				},
			},
		},
		{
			name:     "ok, WriteMultipleCoilsResponseTCP (fc15)",
			whenData: []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x6, 0x3, 0xf, 0x0, 0x2, 0x0, 0x1},
			expect: &WriteMultipleCoilsResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
					Length:        6,
				},
				WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
					UnitID:       3,
					StartAddress: 2,
					CoilCount:    1,
				},
			},
		},
		{
			name:     "ok, WriteMultipleRegistersResponseTCP (fc16)",
			whenData: []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x10, 0x0, 0x2, 0x0, 0x1},
			expect: &WriteMultipleRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 0x1234,
					ProtocolID:    0,
					Length:        6,
				},
				WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
					UnitID: 1,
					// +1 function code
					StartAddress:  2,
					RegisterCount: 1,
				},
			},
		},
		{
			name:     "ok, ReadWriteMultipleRegistersResponseTCP (fc23)",
			whenData: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x17, 0x02, 0xCD, 0x6B},
			expect: &ReadWriteMultipleRegistersResponseTCP{
				MBAPHeader: MBAPHeader{
					TransactionID: 33152,
					ProtocolID:    0,
					Length:        5,
				},
				ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
					UnitID:          3,
					RegisterByteLen: 2,
					Data:            []byte{0xcd, 0x6b},
				},
			},
		},
		{
			name:        "ok, ErrorResponseTCP (code=3)",
			whenData:    []byte{0x4, 0xdd, 0x0, 0x0, 0x0, 0x3, 0x1, 0x82, 0x3},
			expect:      nil,
			expectError: "Illegal data value",
		},
		{
			name:        "nok, data too short",
			whenData:    []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x6},
			expect:      nil,
			expectError: "data is too short to be a Modbus TCP packet",
		},
		{
			name:        "nok, unknown function code",
			whenData:    []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 30 /*fc30*/, 0x0, 0x2, 0x1, 0x2},
			expect:      nil,
			expectError: "unknown function code parsed: 30",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseTCPResponse(tc.whenData)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseRTUResponse(t *testing.T) {
	var testCases = []struct {
		name        string
		whenData    []byte
		expect      Response
		expectError string
	}{
		{
			name:     "ok, ReadCoilsResponseRTU (fc01)",
			whenData: []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xec, 0xd2},
			expect: &ReadCoilsResponseRTU{
				ReadCoilsResponse: ReadCoilsResponse{
					UnitID:          16,
					CoilsByteLength: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:     "ok, ReadDiscreteInputsResponseRTU (fc02)",
			whenData: []byte{0x1, 0x2, 0x2, 0x1, 0x2, 0x22, 0x22},
			expect: &ReadDiscreteInputsResponseRTU{
				ReadDiscreteInputsResponse: ReadDiscreteInputsResponse{
					UnitID:           1,
					InputsByteLength: 2,
					Data:             []byte{0x1, 0x2},
				},
			},
		},
		{
			name:     "ok, ReadHoldingRegistersResponseRTU (fc03)",
			whenData: []byte{0x10, 0x3, 0x2, 0x1, 0x2, 0xe, 0xd3},
			expect: &ReadHoldingRegistersResponseRTU{
				ReadHoldingRegistersResponse: ReadHoldingRegistersResponse{
					UnitID:          16,
					RegisterByteLen: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:     "ok, ReadInputRegistersResponseRTU (fc04)",
			whenData: []byte{0x10, 0x4, 0x2, 0x1, 0x2, 0xb9, 0xd2},
			expect: &ReadInputRegistersResponseRTU{
				ReadInputRegistersResponse: ReadInputRegistersResponse{
					UnitID:          16,
					RegisterByteLen: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:     "ok, WriteSingleCoilResponseRTU (fc05)",
			whenData: []byte{0x1, 0x5, 0x0, 0x2, 0xff, 0x0, 0x13, 0x9d},
			expect: &WriteSingleCoilResponseRTU{
				WriteSingleCoilResponse: WriteSingleCoilResponse{
					UnitID:       1,
					StartAddress: 2,
					CoilState:    true,
				},
			},
		},
		{
			name:     "ok, WriteSingleRegisterResponseRTU (fc06)",
			whenData: []byte{0x1, 0x6, 0x0, 0x2, 0x1, 0x2, 0x3b, 0x3e},
			expect: &WriteSingleRegisterResponseRTU{
				WriteSingleRegisterResponse: WriteSingleRegisterResponse{
					UnitID:  1,
					Address: 2,
					Data:    [2]byte{0x1, 0x2},
				},
			},
		},
		{
			name:     "ok, WriteMultipleCoilsResponseRTU (fc15)",
			whenData: []byte{0x1, 0xf, 0x0, 0x2, 0x0, 0x1, 0xc7, 0x56},
			expect: &WriteMultipleCoilsResponseRTU{
				WriteMultipleCoilsResponse: WriteMultipleCoilsResponse{
					UnitID:       1,
					StartAddress: 2,
					CoilCount:    1,
				},
			},
		},
		{
			name:     "ok, WriteMultipleRegistersResponseRTU (fc16)",
			whenData: []byte{0x1, 0x10, 0x0, 0x2, 0x0, 0x1, 0x6, 0xb8},
			expect: &WriteMultipleRegistersResponseRTU{
				WriteMultipleRegistersResponse: WriteMultipleRegistersResponse{
					UnitID: 1,
					// +1 function code
					StartAddress:  2,
					RegisterCount: 1,
				},
			},
		},
		{
			name:     "ok, ReadWriteMultipleRegistersResponseRTU (fc23)",
			whenData: []byte{0x10, 0x17, 0x2, 0x1, 0x2, 0xe, 0xd3},
			expect: &ReadWriteMultipleRegistersResponseRTU{
				ReadWriteMultipleRegistersResponse: ReadWriteMultipleRegistersResponse{
					UnitID:          16,
					RegisterByteLen: 2,
					Data:            []byte{0x1, 0x2},
				},
			},
		},
		{
			name:        "ok, ErrorResponseRTU (code=3)",
			whenData:    []byte{0x1, 0x82, 0x3, 0xa1, 0x0},
			expect:      nil,
			expectError: "Illegal data value",
		},
		{
			name:        "nok, data too short",
			whenData:    []byte{0x1, 0x82, 0x3, 0xa1},
			expect:      nil,
			expectError: "data is too short to be a Modbus RTU packet",
		},
		{
			name:        "nok, unknown function code",
			whenData:    []byte{0x10, 30 /* fc30 */, 0x2, 0x1, 0x2, 0xec, 0xd2},
			expect:      nil,
			expectError: "unknown function code parsed: 30",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseRTUResponse(tc.whenData)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
