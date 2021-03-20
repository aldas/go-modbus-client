package modbus

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net"
	"testing"
	"time"
)

type netConnMock struct {
	mock.Mock
}

func (m *netConnMock) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *netConnMock) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *netConnMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *netConnMock) LocalAddr() net.Addr {
	return &mockAddr{}
}

func (m *netConnMock) RemoteAddr() net.Addr {
	return &mockAddr{}
}

func (m *netConnMock) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *netConnMock) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *netConnMock) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

type mockAddr struct {
}

func (a *mockAddr) Network() string {
	return "tcp"
}

func (a *mockAddr) String() string {
	return "127.0.2.1:5020"
}

func exampleFC1Request() packet.Request {
	return &packet.ReadCoilsRequestTCP{
		MBAPHeader: packet.MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadCoilsRequest: packet.ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     9,
		},
	}
}

func exampleFC1RTURequest() packet.Request {
	return &packet.ReadCoilsRequestRTU{
		ReadCoilsRequest: packet.ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     9,
		},
	}
}

func exampleFC1Response() packet.Response {
	return &packet.ReadCoilsResponseTCP{
		MBAPHeader: packet.MBAPHeader{
			TransactionID: 0x1234,
			ProtocolID:    0,
		},
		ReadCoilsResponse: packet.ReadCoilsResponse{
			UnitID: 1,
			// +1 function code
			CoilsByteLength: 2,
			Data:            []byte{0x0, 0x1},
		},
	}
}

type mockLogger struct {
	mock.Mock
}

func (l *mockLogger) BeforeWrite(toWrite []byte) {
	l.Called(toWrite)
}

func (l *mockLogger) AfterEachRead(received []byte, n int, err error) {
	l.Called(received, n, err)
}

func (l *mockLogger) BeforeParse(received []byte) {
	l.Called(received)
}

func TestWithOptions(t *testing.T) {
	client := NewClient(
		WithProtocolErrorFunc(packet.AsRTUErrorPacket),
		WithParseResponseFunc(packet.ParseRTUResponse),
		WithTimeouts(99*time.Second, 98*time.Second),
		WithHooks(new(mockLogger)),
	)
	assert.NotNil(t, client.asProtocolErrorFunc)
	assert.NotNil(t, client.parseResponseFunc)
	assert.Equal(t, 99*time.Second, client.writeTimeout)
	assert.Equal(t, 98*time.Second, client.readTimeout)
	assert.Equal(t, new(mockLogger), client.hooks)
}

func TestClient_Do_receivePacketWith1Read(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)

	// full packet []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1}
	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(11, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1})
		}).Once()

	logger := new(mockLogger)
	logger.On("BeforeWrite", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once()
	logger.On("AfterEachRead", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1}, 11, nil).Once()
	logger.On("BeforeParse", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1}).Once()

	client := NewTCPClient(WithHooks(logger))
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Equal(t, exampleFC1Response(), response)
	assert.NoError(t, err)

	conn.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestClientRTU_Do_receivePacketWith1Read(t *testing.T) {
	req := &packet.ReadCoilsRequestRTU{
		ReadCoilsRequest: packet.ReadCoilsRequest{
			UnitID:       1,
			StartAddress: 200,
			Quantity:     9,
		},
	}
	resp := &packet.ReadCoilsResponseRTU{
		ReadCoilsResponse: packet.ReadCoilsResponse{
			UnitID: 16,
			// +1 function code
			CoilsByteLength: 2,
			Data:            []byte{0x1, 0x2},
		},
	}

	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", req.Bytes()).Once().Return(0, nil)

	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(7, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae})
		}).Once()

	client := NewRTUClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), req)

	assert.Equal(t, resp, response)
	assert.NoError(t, err)

	conn.AssertExpectations(t)
}

func TestClient_Do_receivePacketWith2Reads(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)

	// full packet []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1}
	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(8, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1}) // first 8 bytes
		}).Once()

	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(3, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x2, 0x0, 0x1}) // last 3 bytes
		}).Once()

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Equal(t, exampleFC1Response(), response)
	assert.NoError(t, err)

	conn.AssertExpectations(t)
}

func TestClient_Do_receiveErrorPacket(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)

	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(9, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x4, 0xdd, 0x0, 0x0, 0x0, 0x3, 0x1, 0x82, 0x3})
		}).Once()

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	expectedErr := &packet.ErrorResponseTCP{TransactionID: 1245, UnitID: 1, Function: 2, Code: 3}
	assert.EqualError(t, err, expectedErr.Error())

	conn.AssertExpectations(t)
}

func TestClient_Do_ReadSomeBytesWithEOF(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)

	// full packet []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1, 0x2, 0x0, 0x1}
	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(8, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1}) // first 8 bytes
		}).Once()

	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(1, io.EOF) // second read should return 3 bytes but returns 1 with io.EOF

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, "received data length too short to be valid packet")

	conn.AssertExpectations(t)
}

func TestClient_Do_contextCancelAfterFirstRead(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)
	ctx, cancel := context.WithCancel(context.Background())

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)
	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(8, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x5, 0x1, 0x1})
			cancel()
		})

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(ctx, exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, context.Canceled.Error())
	conn.AssertExpectations(t)
}

func TestClient_Do_RequestShouldBeSet(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	client := NewTCPClient()
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), nil)

	assert.Nil(t, response)
	assert.EqualError(t, err, "request can not be nil")
	conn.AssertExpectations(t)
}

func TestClient_Do_ClientShouldBeConnected(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	client := NewTCPClient()
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, "client is not connected")
	conn.AssertExpectations(t)
}

func TestClient_Do_SetWriteDeadlineError(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).
		Once().
		Return(errors.New("SetWriteDeadline error"))

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, "SetWriteDeadline error")
	conn.AssertExpectations(t)
}

func TestClient_Do_writeError(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).
		Once().
		Return(0, errors.New("write error"))

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, "write error")
	conn.AssertExpectations(t)
}

func TestClient_Do_unknownReadError(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)
	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(0, io.ErrUnexpectedEOF)

	client := NewTCPClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())
	conn.AssertExpectations(t)
}

func TestClient_Do_ReadMoreBytesThanPacketCanBe(t *testing.T) {
	exampleNow := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00

	conn := new(netConnMock)

	conn.On("SetWriteDeadline", exampleNow.Add(defaultWriteTimeout)).Once().Return(nil)
	conn.On("Write", []byte{0x12, 0x34, 0x0, 0x0, 0x0, 0x6, 0x1, 0x1, 0x0, 0xc8, 0x0, 0x9}).Once().Return(0, nil)
	conn.On("SetReadDeadline", exampleNow.Add(500*time.Microsecond)).Return(nil)
	conn.On("Read", mock.Anything).
		Return(tcpPacketMaxLen+1, nil)

	client := NewClient()
	client.conn = conn
	client.timeNow = func() time.Time {
		return exampleNow
	}

	response, err := client.Do(context.Background(), exampleFC1Request())

	assert.Nil(t, response)
	assert.EqualError(t, err, "received more bytes than valid Modbus packet size can be")
	conn.AssertExpectations(t)
}

func TestClient_Close(t *testing.T) {
	var testCases = []struct {
		name              string
		givenNotConnected bool
		whenError         error
		expectClose       bool
		expectError       string
	}{
		{
			name:        "ok",
			expectClose: true,
		},
		{
			name:              "ok, no connection is no-op",
			givenNotConnected: true,
			expectClose:       false,
		},
		{
			name:        "nok, error on close",
			expectClose: true,
			whenError:   errors.New("close error"),
			expectError: "close error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := new(netConnMock)
			if tc.expectClose {
				conn.On("Close").Once().Return(tc.whenError)
			}

			client := NewTCPClient()
			if !tc.givenNotConnected {
				client.conn = conn
			}

			err := client.Close()
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			conn.AssertExpectations(t)
		})
	}
}

func TestClient_Connect(t *testing.T) {
	var testCases = []struct {
		name               string
		whenAddress        string
		whenDialContextErr error

		expectAddr  string
		expectError string
	}{
		{
			name:        "ok, tcp is default",
			whenAddress: ":502",
			expectAddr:  ":502",
		},
		{
			name:        "ok, domain name, tcp is default",
			whenAddress: "cool.test.com:502",
			expectAddr:  "cool.test.com:502",
		},
		{
			name:        "ok, with specific tcp4",
			whenAddress: "tcp4://192.168.0.1:502",
			expectAddr:  "tcp4://192.168.0.1:502",
		},
		{
			name:        "ok, with specific tcp6",
			whenAddress: "tcp6://::1:502",
			expectAddr:  "tcp6://::1:502",
		},
		{
			name:               "nok, dialContext error",
			whenAddress:        "localhost:502",
			whenDialContextErr: errors.New("dialContext error"),
			expectAddr:         "localhost:502",
			expectError:        "dialContext error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := new(netConnMock)

			client := NewTCPClient()
			client.dialContextFunc = func(_ context.Context, addr string) (net.Conn, error) {
				assert.Equal(t, tc.expectAddr, addr)

				return new(netConnMock), tc.whenDialContextErr
			}

			err := client.Connect(context.Background(), tc.whenAddress)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client.conn)
				assert.Equal(t, tc.whenAddress, client.address)
			}
			conn.AssertExpectations(t)
		})
	}
}

func TestAddressExtractor(t *testing.T) {
	var testCases = []struct {
		name        string
		whenAddress string

		expectNetwork string
		expectAddr    string
	}{
		{
			name:          "ok, tcp is default",
			whenAddress:   ":502",
			expectNetwork: "tcp",
			expectAddr:    ":502",
		},
		{
			name:          "ok, domain name, tcp is default",
			whenAddress:   "cool.test.com:502",
			expectNetwork: "tcp",
			expectAddr:    "cool.test.com:502",
		},
		{
			name:          "ok, with specific tcp4",
			whenAddress:   "tcp4://192.168.0.1:502",
			expectNetwork: "tcp4",
			expectAddr:    "192.168.0.1:502",
		},
		{
			name:          "ok, with specific tcp6",
			whenAddress:   "tcp6://::1:502",
			expectNetwork: "tcp6",
			expectAddr:    "::1:502",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			network, addr := addressExtractor(tc.whenAddress)

			assert.Equal(t, network, tc.expectNetwork)
			assert.Equal(t, addr, tc.expectAddr)
		})
	}
}
