package packet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseMBAPHeader(t *testing.T) {
	var testCases = []struct {
		name        string
		when        []byte
		expect      MBAPHeader
		expectError string
	}{
		{
			name: "ok, ErrorResponseTCP (code=3)",
			when: []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x3, 0x1, 0x82, 0x3},
			expect: MBAPHeader{
				TransactionID: 33152,
				ProtocolID:    0,
			},
		},
		{
			name: "ok, ReadCoilsRequestTCP (fc1)",
			when: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expect: MBAPHeader{
				TransactionID: 258,
				ProtocolID:    0,
			},
		},
		{
			name: "ok, ReadWriteMultipleRegistersResponseTCP (fc23)",
			when: []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x05, 0x03, 0x17, 0x02, 0xCD, 0x6B},
			expect: MBAPHeader{
				TransactionID: 33152,
				ProtocolID:    0,
			},
		},
		{
			name:        "nok, data to short to contain MBAPHeader",
			when:        []byte{0x81, 0x80, 0x00, 0x00, 0x00},
			expectError: "data to short to contain MBAPHeader",
		},
		{
			name:        "nok, data to short to contain MBAPHeader",
			when:        []byte{0x81, 0x80, 0x00, 0x01, 0x00, 0x00},
			expectError: "invalid protocol id",
		},
		{
			name:        "nok, pdu length in header not be 0",
			when:        []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x00},
			expectError: "pdu length in header not be 0",
		},
		{
			name:        "nok, packet length does not match length in header",
			when:        []byte{0x81, 0x80, 0x00, 0x00, 0x00, 0x02, 0xff},
			expectError: "packet length does not match length in header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseMBAPHeader(tc.when)

			assert.Equal(t, tc.expect, result)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsLikeModbusTCP(t *testing.T) {
	var testCases = []struct {
		name                   string
		when                   []byte
		whenAllowUnsupportedFC bool
		expectLength           int
		expectLooksLike        LooksLikeType
	}{
		{
			name:            "ok, full packet",
			when:            []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expectLength:    12,
			expectLooksLike: LooksLikeTCPPacket,
		},
		{
			name:            "ok, fragment of packet",
			when:            []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00},
			expectLength:    12,
			expectLooksLike: LooksLikeTCPPacket,
		},
		{
			name:            "nok, ErrorResponseTCP (code=3)",
			when:            []byte{0x81, 0x80, 0x0, 0x0, 0x0, 0x3, 0x1, 0x82, 0x3},
			expectLength:    9,
			expectLooksLike: UnsupportedFunctionCode,
		},
		{
			name:            "nok, too few bytes",
			when:            []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x01},
			expectLength:    0,
			expectLooksLike: DataTooShort,
		},
		{
			name:            "nok, invalid packet id, 1",
			when:            []byte{0x01, 0x02, 0x01 /* 0x00 */, 0x00, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expectLength:    0,
			expectLooksLike: IsNotTPCPacket,
		},
		{
			name:            "nok, invalid packet id, 2",
			when:            []byte{0x01, 0x02, 0x00, 0x01 /* 0x00 */, 0x00, 0x06, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expectLength:    0,
			expectLooksLike: IsNotTPCPacket,
		},
		{
			name:            "nok, pdu too short",
			when:            []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x02 /* 0x04+ */, 0x10, 0x01, 0x00, 0x6B, 0x00, 0x03},
			expectLength:    0,
			expectLooksLike: IsNotTPCPacket,
		},
		{
			name:            "nok, function code = 0",
			when:            []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x00 /* 0x01 */, 0x00, 0x6B, 0x00, 0x03},
			expectLength:    0,
			expectLooksLike: IsNotTPCPacket,
		},
		{
			name:                   "ok, allow unsupported function code = 1F",
			when:                   []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x1f /* 0x01 */, 0x00, 0x6B, 0x00, 0x03},
			whenAllowUnsupportedFC: true,
			expectLength:           12,
			expectLooksLike:        LooksLikeTCPPacket,
		},
		{
			name:                   "ok, unsupported function code = 1F",
			when:                   []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x06, 0x10, 0x1f /* 0x01 */, 0x00, 0x6B, 0x00, 0x03},
			whenAllowUnsupportedFC: false,
			expectLength:           12,
			expectLooksLike:        UnsupportedFunctionCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedLen, looksLike := IsLikeModbusTCP(tc.when, tc.whenAllowUnsupportedFC)

			assert.Equal(t, tc.expectLength, expectedLen)
			assert.Equal(t, tc.expectLooksLike, looksLike)
		})
	}
}

func TestCRC16(t *testing.T) {
	var testCases = []struct {
		name   string
		when   []byte
		expect uint16
	}{
		{
			name:   "ok",
			when:   []byte{0x01, 0x04, 0x02, 0xFF, 0xFF},
			expect: 0x80B8,
		},
		{
			name:   "ok",
			when:   []byte{0x01, 0x04, 0x02, 0xFF, 0xFF},
			expect: 0x80B8,
		},
		{
			name:   "ok",
			when:   []byte{0x11, 0x03, 0x00, 0x6B, 0x00, 0x03},
			expect: 0x8776,
		},
		{
			name:   "ok2",
			when:   []byte{0x03, 0x03, 0x02, 0xCD, 0x6B},
			expect: 0xFBD4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CRC16(tc.when)
			assert.Equal(t, tc.expect, result)
		})
	}
}
