package modbus

import (
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"
	"testing"
)

func TestSplit_validationError(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: "", UnitID: 0, // ServerAddress is empty
			Address: 1, Type: FieldTypeInt8,
		},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.EqualError(t, err, "field server address can not be empty")
	assert.Nil(t, batched)
}

func TestSplit_single(t *testing.T) {
	given := []Field{
		{ServerAddress: ":502", UnitID: 0, Address: 1, Type: FieldTypeInt8},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	pReq, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	pReq.TransactionID = 123
	expect := BuilderRequest{
		ServerAddress: ":502",
		StartAddress:  1,
		Protocol:      ProtocolTCP,
		Request:       pReq,
		Fields: []Field{
			{ServerAddress: ":502", UnitID: 0, Address: 1, Type: FieldTypeInt8, FunctionCode: 3, Protocol: ProtocolTCP},
		},
	}
	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0])
}

func TestSplit_quantity(t *testing.T) {
	var testCases = []struct {
		name               string
		givenFields        []Field
		expectStartAddress uint16
		expectQuantity     uint16
	}{
		{
			name: "ok, int64",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 10, Type: FieldTypeFloat64},
			},
			expectStartAddress: 10,
			expectQuantity:     4,
		},
		{
			name: "ok, multiple fields",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 10, Type: FieldTypeInt8},
				{ServerAddress: ":502", UnitID: 0, Address: 74, Type: FieldTypeUint32},
			},
			expectStartAddress: 10,
			expectQuantity:     66,
		},
		{
			name: "ok, multiple fields int8",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 10, Type: FieldTypeInt8},
				{ServerAddress: ":502", UnitID: 0, Address: 75, Type: FieldTypeInt8},
			},
			expectStartAddress: 10,
			expectQuantity:     66,
		},
		{
			name: "ok, multiple fields int64",
			givenFields: []Field{
				{ServerAddress: ":502", UnitID: 0, Address: 0, Type: FieldTypeInt8},
				{ServerAddress: ":502", UnitID: 0, Address: 1, Type: FieldTypeFloat64},
			},
			expectStartAddress: 0,
			expectQuantity:     5,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			batched, err := split(tc.givenFields, 3, ProtocolTCP)
			if assert.NoError(t, err) {
				assert.Len(t, batched, 1)
				req := batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP)
				assert.Equal(t, tc.expectStartAddress, req.StartAddress)
				assert.Equal(t, tc.expectQuantity, req.Quantity)
			}
		})
	}
}

func TestSplit_many(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeInt8,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 118, Length: 11, Type: FieldTypeString, // 118 + 6 + 124
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 121, Type: FieldTypeUint64,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeInt16,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeFloat32,
		},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 1)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 124)
	expect.TransactionID = 123

	batched[0].Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, batched[0].Request)
}

func TestSplit_to2RegisterBatches(t *testing.T) {
	given := []Field{
		{Name: "F1",
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeInt8,
		},
		{Name: "F2",
			ServerAddress: ":502", UnitID: 0,
			Address: 119, Length: 15, Type: FieldTypeString, // 119,120,121,122, 123,124,125,126 == new request
		},
		{Name: "F3",
			ServerAddress: ":502", UnitID: 0,
			Address: 121, Type: FieldTypeUint64, // 121,122,123,124
		},
		{Name: "F4",
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeFloat32, // 122, 123
		},
		{Name: "F5",
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeCoil, // should be ignored
		},
	}

	batched, err := split(given, 3, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 2)

	expect, _ := packet.NewReadHoldingRegistersRequestTCP(0, 1, 1)
	expect.TransactionID = 123

	firstBatch := batched[0]
	firstBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 123
	assert.Equal(t, expect, firstBatch.Request)
	assert.Len(t, firstBatch.Fields, 1)

	expect2, _ := packet.NewReadHoldingRegistersRequestTCP(0, 119, 8)
	expect2.TransactionID = 124

	secondBatch := batched[1]
	secondBatch.Request.(*packet.ReadHoldingRegistersRequestTCP).TransactionID = 124
	assert.Equal(t, expect2, secondBatch.Request)
	assert.Len(t, secondBatch.Fields, 3)
}

func TestSplit_to2CoilsBatches(t *testing.T) {
	given := []Field{
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeCoil,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 1, Type: FieldTypeCoil, // at same place previous field
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 100, Type: FieldTypeCoil,
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 2001, Type: FieldTypeCoil, // should go to next batch
		},
		{
			ServerAddress: ":502", UnitID: 0,
			Address: 122, Type: FieldTypeFloat32, // should be ignored
		},
	}

	batched, err := split(given, 1, ProtocolTCP)
	assert.NoError(t, err)
	assert.Len(t, batched, 2)

	expect, _ := packet.NewReadCoilsRequestTCP(0, 1, 100)
	expect.TransactionID = 123

	firstBatch := batched[0]
	firstBatch.Request.(*packet.ReadCoilsRequestTCP).TransactionID = 123
	assert.Equal(t, expect, firstBatch.Request)
	assert.Len(t, firstBatch.Fields, 3)

	expect2, _ := packet.NewReadCoilsRequestTCP(0, 2001, 1)
	expect2.TransactionID = 124

	secondBatch := batched[1]
	secondBatch.Request.(*packet.ReadCoilsRequestTCP).TransactionID = 124
	assert.Equal(t, expect2, secondBatch.Request)
	assert.Len(t, secondBatch.Fields, 1)
}

func TestSplit_Serialto2CoilsBatches(t *testing.T) {
	given := []Field{
		{
			ServerAddress: "/dev/ttyS0:9600?invalid_addr=70", UnitID: 0,
			Address: 1, Type: FieldTypeUint16,
		},
		{
			ServerAddress: "/dev/ttyS0:9600?invalid_addr=70", UnitID: 1,
			Address: 2, Type: FieldTypeUint16,
		},
	}

	batched, err := split(given, 3, ProtocolRTU)
	assert.NoError(t, err)
	assert.Len(t, batched, 2)
}

func TestSplit_maxQuantityPerRequest(t *testing.T) {
	given := []Field{
		{
			ServerAddress: "/dev/ttyS0:9600?max_quantity_per_request=2", UnitID: 1,
			Address: 1, Type: FieldTypeUint16,
		},
		{
			ServerAddress: "/dev/ttyS0:9600?max_quantity_per_request=2", UnitID: 1,
			Address: 2, Type: FieldTypeUint16,
		},
		{
			ServerAddress: "/dev/ttyS0:9600?max_quantity_per_request=2", UnitID: 1,
			Address: 3, Type: FieldTypeUint16,
		},
	}

	batched, err := split(given, 3, ProtocolRTU)
	assert.NoError(t, err)
	assert.Len(t, batched, 2)
}

func TestInvalidRange_Overlaps(t *testing.T) {
	testCases := []struct {
		name             string
		givenRange       *invalidRange
		whenRequestStart uint16
		whenSlotStart    uint16
		whenSlotEnd      uint16
		expectOverlap    bool
		expectErr        bool
		expectErrSubstr  string
	}{
		{
			name:             "Nil receiver",
			givenRange:       nil,
			whenRequestStart: 100,
			whenSlotStart:    150,
			whenSlotEnd:      200,
			// No invalidRanges to check => no overlap
			expectOverlap: false,
			expectErr:     false,
		},
		{
			name:             "Empty slice",
			givenRange:       &invalidRange{},
			whenRequestStart: 100,
			whenSlotStart:    150,
			whenSlotEnd:      200,
			// No elements => no overlap
			expectOverlap: false,
			expectErr:     false,
		},
		{
			name: "No overlap",
			givenRange: &invalidRange{
				{from: 10, to: 20}, // well below the slot range
			},
			whenRequestStart: 100,
			whenSlotStart:    150,
			whenSlotEnd:      200,
			expectOverlap:    false,
			expectErr:        false,
		},
		{
			name: "Overlap with error (slot range overlap)",
			givenRange: &invalidRange{
				{from: 150, to: 160},
				{from: 170, to: 180},
			},
			whenRequestStart: 100,
			whenSlotStart:    155,
			whenSlotEnd:      158,
			// Overlaps [155..158] with [150..160] => error
			expectOverlap:   true,
			expectErr:       true,
			expectErrSubstr: "field overlaps invalid address range",
		},
		{
			name: "Overlap with no error (whenRequestStart overlap only)",
			givenRange: &invalidRange{
				{from: 120, to: 130},
				{from: 140, to: 145},
			},
			// Here we assume that overlapping only on the whenRequestStart
			// triggers a true but no error (based on second check).
			whenRequestStart: 100,
			whenSlotStart:    150,
			whenSlotEnd:      155,
			// [150..155] doesn’t overlap with either range,
			// but whenRequestStart=100 does overlap with [120..130] check?
			// In practice, this example won't return an error, but does it return true or false?
			// Because "ir.from <= whenSlotEnd && whenRequestStart <= ir.to"
			// => 120 <= 155 && 100 <= 130 => true => (true, nil).
			expectOverlap: true,
			expectErr:     false,
		},
		{
			name: "Edge boundary overlap (exactly at boundaries)",
			givenRange: &invalidRange{
				{from: 100, to: 200},
			},
			whenRequestStart: 100,
			whenSlotStart:    200,
			whenSlotEnd:      200,
			// The condition `ir.from <= whenSlotEnd && whenSlotStart <= ir.to`
			// => 100 <= 200 && 200 <= 200 => true => overlap => error
			expectOverlap:   true,
			expectErr:       true,
			expectErrSubstr: "field overlaps invalid address range",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotOverlap, err := tt.givenRange.Overlaps(tt.whenRequestStart, tt.whenSlotStart, tt.whenSlotEnd)
			if gotOverlap != tt.expectOverlap {
				t.Errorf("Overlaps() gotOverlap = %v, want %v", gotOverlap, tt.expectOverlap)
			}
			if (err != nil) != tt.expectErr {
				t.Errorf("Overlaps() error = %v, expectErr = %v", err, tt.expectErr)
			}
			if err != nil && tt.expectErrSubstr != "" && !strings.Contains(err.Error(), tt.expectErrSubstr) {
				t.Errorf("Overlaps() error = %q, want substring %q", err.Error(), tt.expectErrSubstr)
			}
		})
	}
}

func TestAddressToSplitterConfig(t *testing.T) {
	var testCases = []struct {
		name        string
		whenAddress string
		expect      splitterConfig
		expectErr   string
	}{
		{
			name:        "ok, max_quantity_per_request",
			whenAddress: "/dev/ttyS0?max_quantity_per_request=16",
			expect: splitterConfig{
				MaxQuantityPerRequest: 16,
				InvalidRange:          nil,
			},
		},
		{
			name:        "nok, invalid max_quantity_per_request",
			whenAddress: "/dev/ttyS0?max_quantity_per_request=-1",
			expect:      splitterConfig{},
			expectErr:   `failed to parse max_quantity_per_request, err: strconv.ParseUint: parsing "-1": invalid syntax`,
		},
		{
			name:        "ok, single invalid address",
			whenAddress: "/dev/ttyS0?invalid_addr=10",
			expect: splitterConfig{
				MaxQuantityPerRequest: 0,
				InvalidRange:          invalidRange{{from: 10, to: 10}},
			},
		},
		{
			name:        "ok, random port url is problematic",
			whenAddress: "[::]:45310?invalid_addr=10",
			expectErr:   `failed to parse server adddres for invalid range: "[::]:45310?invalid_addr=10", err: parse "[::]:45310?invalid_addr=10": first path segment in URL cannot contain colon`,
		},
		{
			name:        "ok, empty random port url",
			whenAddress: "[::]:45310",
		},
		{
			name:        "ok, empty ip",
			whenAddress: "192.168.100.101:502",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := addressToSplitterConfig(tc.whenAddress)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddressToInvalidRange(t *testing.T) {
	var testCases = []struct {
		name        string
		whenAddress string
		expect      invalidRange
		expectErr   string
	}{
		{
			name:        "ok, empty serial device url",
			whenAddress: "/dev/ttyS0",
		},
		{
			name:        "ok, serial device url without param",
			whenAddress: "/dev/ttyS0?something=else",
		},
		{
			name:        "ok, single address",
			whenAddress: "/dev/ttyS0?invalid_addr=10",
			expect: invalidRange{
				{from: 10, to: 10},
			},
		},
		{
			name:        "ok, address range",
			whenAddress: "/dev/ttyS0?invalid_addr=11-52",
			expect: invalidRange{
				{from: 11, to: 52},
			},
		},
		{
			name:        "ok, multiple address ranges",
			whenAddress: "/dev/ttyS0?invalid_addr=11-52,5",
			expect: invalidRange{
				{from: 11, to: 52},
				{from: 5, to: 5},
			},
		},
		{
			name:        "ok, multiple address ranges",
			whenAddress: "/dev/ttyS0?invalid_addr=11-52&invalid_addr=5",
			expect: invalidRange{
				{from: 11, to: 52},
				{from: 5, to: 5},
			},
		},
		{
			name:        "nok, invalid single address",
			whenAddress: "tcp://192.168.1.2:502?invalid_addr=1x",
			expect:      nil,
			expectErr:   `failed to parse invalid range: "1x", err: strconv.ParseUint: parsing "1x": invalid syntax`,
		},
		{
			name:        "nok, invalid address range",
			whenAddress: "tcp://192.168.1.2:502?invalid_addr=11-5x2",
			expect:      nil,
			expectErr:   `failed to parse invalid range: "11-5x2", err: strconv.ParseUint: parsing "5x2": invalid syntax`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			address, err := url.Parse(tc.whenAddress)
			if err != nil {
				t.Fatalf("failed to parse address: %v", err)
			}
			result, err := addressToInvalidRange(address)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBatchToRequests(t *testing.T) {
	var testCases = []struct {
		name      string
		when      []builderSlotGroup
		expect    []requestBatch
		expectErr string
	}{
		{
			name: "ok, split at invalid address",
			when: []builderSlotGroup{
				{
					group: groupID{
						serverAddress: "/dev/ttyS0?invalid_addr=15-20&invalid_addr=5",
						functionCode:  3,
						unitID:        1,
						protocol:      ProtocolRTU,
					},
					slots: builderSlots{
						{address: 2, size: 1, fields: nil},  // 2
						{address: 3, size: 2, fields: nil},  // 3,4
						{address: 10, size: 4, fields: nil}, // 10,11,12,13
					},
				},
			},
			expect: []requestBatch{
				{
					ServerAddress:   "/dev/ttyS0?invalid_addr=15-20&invalid_addr=5",
					FunctionCode:    3,
					UnitID:          1,
					Protocol:        ProtocolRTU,
					StartAddress:    2,
					Quantity:        3,
					RequestInterval: 0,
					fields:          nil,
				},
				{
					ServerAddress:   "/dev/ttyS0?invalid_addr=15-20&invalid_addr=5",
					FunctionCode:    3,
					UnitID:          1,
					Protocol:        ProtocolRTU,
					StartAddress:    10,
					Quantity:        4,
					RequestInterval: 0,
					fields:          nil,
				},
			},
		},
		{
			name: "nok, error when slot falls into range",
			when: []builderSlotGroup{
				{
					group: groupID{
						serverAddress: "/dev/ttyS0?invalid_addr=15-20&invalid_addr=5",
						functionCode:  3,
						unitID:        1,
						protocol:      ProtocolRTU,
					},
					slots: builderSlots{
						{address: 2, size: 1, fields: nil},  // 2
						{address: 3, size: 2, fields: nil},  // 3,4
						{address: 20, size: 4, fields: nil}, // 20,21,22,23
					},
				},
			},
			expect:    nil,
			expectErr: `field overlaps invalid address range, addr: 20, range: 15-20`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := batchToRequests(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
