package packet

import (
	"encoding/binary"
	"errors"
)

// ReadServerIDResponseTCP is TCP Response for Read Server ID (FC=17) 0x11
//
// Example packet: 0x81 0x80 0x00 0x00 0x00 0x08 0x10 0x11 0x02 0x01 0x02 0x00 0x01 0x02
// 0x81 0x80 - transaction id  (0,1)
// 0x00 0x00 - protocol id (2,3)
// 0x00 0x08 - number of bytes in the message (PDU = ProtocolDataUnit) to follow (4,5)
// 0x10 - unit id (6)
// 0x11 - function code (7)
// 0x02 - byte count for server id (8)
// 0x01 0x02 - N bytes for server id (device specific, variable length) (9,10)
// 0x00 - run status (11)
// 0x01 0x02 - optional N bytes for additional data (device specific, variable length) (12,13)
type ReadServerIDResponseTCP struct {
	MBAPHeader
	ReadServerIDResponse
}

// ReadServerIDResponseRTU is RTU Response for Read Server ID (FC=17) 0x11
//
// Example packet: 0x10 0x11 0x02 0x01 0x02 0x00 0x01 0x02 0xd5 0x43
// 0x10 - unit id (0)
// 0x11 - function code (1)
// 0x02 - byte count for server id (2)
// 0x01 0x02 - N bytes for server id (device specific, variable length) (3,4)
// 0x00 - run status (5)
// 0x01 0x02 - optional N bytes for additional data (device specific, variable length) (6,7)
// 0xd5 0x43 - CRC16 (n-2,n-1)
type ReadServerIDResponseRTU struct {
	ReadServerIDResponse
}

// ReadServerIDResponse is Response for Read Server ID (FC=17) 0x11
type ReadServerIDResponse struct {
	UnitID         uint8
	Status         uint8
	ServerID       []byte
	AdditionalData []byte
}

// Bytes returns ReadServerIDResponseTCP packet as bytes form
func (r ReadServerIDResponseTCP) Bytes() []byte {
	length := r.ReadServerIDResponse.len()
	result := make([]byte, tcpMBAPHeaderLen+length)
	r.MBAPHeader.bytes(result[0:6], length)
	r.ReadServerIDResponse.bytes(result[6:])
	return result
}

// ParseReadServerIDResponseTCP parses given bytes into ReadServerIDResponseTCP
func ParseReadServerIDResponseTCP(data []byte) (*ReadServerIDResponseTCP, error) {
	dLen := len(data)
	if dLen < 11 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	serverIDLen := int(data[8])
	if serverIDLen == 0 {
		return nil, errors.New("server id length too small to be valid packet")
	}
	statusIdx := 8 + serverIDLen + 1
	if statusIdx >= len(data) {
		return nil, errors.New("received data length too short to be valid packet")
	}
	serverID := make([]byte, serverIDLen)
	copy(serverID, data[9:9+serverIDLen])

	status := data[statusIdx]

	var additionalData []byte
	if len(data) > statusIdx+1 {
		additionalData = make([]byte, len(data)-statusIdx-1)
		copy(additionalData, data[statusIdx+1:])
	}

	return &ReadServerIDResponseTCP{
		MBAPHeader: MBAPHeader{
			TransactionID: binary.BigEndian.Uint16(data[0:2]),
			ProtocolID:    0,
		},
		ReadServerIDResponse: ReadServerIDResponse{
			UnitID: data[6],
			// fc (7)
			// server id len (8)
			ServerID:       serverID,
			Status:         status,
			AdditionalData: additionalData,
		},
	}, nil
}

// Bytes returns ReadServerIDResponseRTU packet as bytes form
func (r ReadServerIDResponseRTU) Bytes() []byte {
	length := r.len() + 2
	result := make([]byte, length)
	bytes := r.ReadServerIDResponse.bytes(result)
	crc := CRC16(bytes[:length-2])
	result[length-2] = uint8(crc)
	result[length-1] = uint8(crc >> 8)
	return result
}

// ParseReadServerIDResponseRTU parses given bytes into ReadServerIDResponseRTU
func ParseReadServerIDResponseRTU(data []byte) (*ReadServerIDResponseRTU, error) {
	dLen := len(data)
	if dLen < 7 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	serverIDLen := int(data[2])
	if serverIDLen == 0 {
		return nil, errors.New("server id length too small to be valid packet")
	}
	statusIdx := 2 + serverIDLen + 1
	if statusIdx >= len(data)-2 {
		return nil, errors.New("received data length too short to be valid packet")
	}
	serverID := make([]byte, serverIDLen)
	copy(serverID, data[3:3+serverIDLen])

	status := data[statusIdx]

	var additionalData []byte
	if len(data) > statusIdx+1 {
		additionalData = make([]byte, len(data)-2-statusIdx-1)
		copy(additionalData, data[statusIdx+1:len(data)-2])
	}
	return &ReadServerIDResponseRTU{
		ReadServerIDResponse: ReadServerIDResponse{
			UnitID: data[0],
			// fc (1)
			// server id len (2)
			ServerID:       serverID,
			Status:         status,
			AdditionalData: additionalData,
		},
	}, nil
}

// FunctionCode returns function code of this request
func (r ReadServerIDResponse) FunctionCode() uint8 {
	return FunctionReadServerID
}

func (r ReadServerIDResponse) len() uint16 {
	// unit id (1) + fc (1) + server id len (1) + server id (N) + status (1) + additional data (N)
	return 4 + uint16(len(r.ServerID)) + uint16(len(r.AdditionalData))
}

// Bytes returns ReadServerIDResponse packet as bytes form
func (r ReadServerIDResponse) Bytes() []byte {
	return r.bytes(make([]byte, r.len()))
}

func (r ReadServerIDResponse) bytes(data []byte) []byte {
	data[0] = r.UnitID
	data[1] = FunctionReadServerID

	serverIDLen := uint8(len(r.ServerID))
	data[2] = serverIDLen
	copy(data[3:3+serverIDLen], r.ServerID)

	data[3+serverIDLen] = r.Status
	if r.AdditionalData != nil {
		copy(data[4+serverIDLen:], r.AdditionalData)
	}

	return data
}
