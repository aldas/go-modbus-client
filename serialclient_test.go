package modbus

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
	"time"
)

type serialMock struct {
	mock.Mock
}

func (m *serialMock) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *serialMock) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *serialMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *serialMock) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func exampleFC1RTURequest() packet.Request {
	return &packet.ReadCoilsRequestRTU{
		ReadCoilsRequest: packet.ReadCoilsRequest{
			UnitID:       16,
			StartAddress: 200,
			Quantity:     9,
		},
	}
}

func exampleFC1RTUResponse() packet.Response {
	return &packet.ReadCoilsResponseRTU{
		ReadCoilsResponse: packet.ReadCoilsResponse{
			UnitID: 16,
			// +1 function code
			CoilsByteLength: 2,
			Data:            []byte{0x1, 0x2},
		},
	}
}

type mockSerialLogger struct {
	mock.Mock
}

func (l *mockSerialLogger) BeforeWrite(toWrite []byte) {
	l.Called(toWrite)
}

func (l *mockSerialLogger) AfterEachRead(received []byte, n int, err error) {
	l.Called(received, n, err)
}

func (l *mockSerialLogger) BeforeParse(received []byte) {
	l.Called(received)
}

func TestSerialClient_WithOptions(t *testing.T) {
	serialMock := new(serialMock)
	client := NewSerialClient(
		serialMock,
		WithSerialReadTimeout(4*time.Second),
		WithSerialHooks(new(mockSerialLogger)),
	)
	assert.Equal(t, 4*time.Second, client.readTimeout)
	assert.NotNil(t, client.asProtocolErrorFunc)
	assert.NotNil(t, client.parseResponseFunc)
	assert.Equal(t, new(mockSerialLogger), client.hooks)
}

func TestSerialClient_Do_receivePacketWith1Read(t *testing.T) {
	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Flush").Once().Return(nil)

	// full packet []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae}
	serialPort.On("Read", mock.Anything).
		Return(7, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae})
		}).Once()

	logger := new(mockLogger)
	logger.On("BeforeWrite", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once()
	logger.On("AfterEachRead", []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae}, 7, nil).Once()
	logger.On("BeforeParse", []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae}).Once()

	client := NewSerialClient(serialPort, WithSerialHooks(logger))

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Equal(t, exampleFC1RTUResponse(), response)
	assert.NoError(t, err)

	serialPort.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestSerialClient_Do_receivePacketWith2Reads(t *testing.T) {
	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Flush").Once().Return(nil)

	// full packet []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae}
	serialPort.On("Read", mock.Anything).
		Return(5, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x10, 0x1, 0x2, 0x1, 0x2}) // first 5 bytes
		}).Once()

	serialPort.On("Read", mock.Anything).
		Return(2, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0xc5, 0xae}) // last 2 bytes
		}).Once()

	client := NewSerialClient(serialPort)

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Equal(t, exampleFC1RTUResponse(), response)
	assert.NoError(t, err)

	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_receiveErrorPacket(t *testing.T) {
	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Flush").Once().Return(nil)

	serialPort.On("Read", mock.Anything).
		Return(5, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x01, 0x81, 0x01, 0x81, 0x90})
		}).Once()

	client := NewSerialClient(serialPort)
	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Nil(t, response)
	expectedErr := &packet.ErrorResponseRTU{UnitID: 1, Function: 1, Code: 1}
	assert.EqualError(t, err, expectedErr.Error())

	var target *ClientError
	assert.True(t, errors.As(err, &target))

	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_ReadSomeBytesWithEOF(t *testing.T) {
	// when `github.com/tarm/serial` serialport is created as nonblocking. `Read` can return
	// 0 bytes return EOF error on Linux / POSIX on read deadline timeout.

	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Flush").Once().Return(nil)

	// full packet []byte{0x10, 0x1, 0x2, 0x1, 0x2, 0xc5, 0xae}
	serialPort.On("Read", mock.Anything).
		Return(5, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x10, 0x1, 0x2, 0x1, 0x2}) // first 5 bytes
		}).Once()

	serialPort.On("Read", mock.Anything).
		Return(0, io.EOF).
		Once() // second read should return 2 bytes but returns 1 with io.EOF

	serialPort.On("Read", mock.Anything).
		Return(2, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0xc5, 0xae}) // last 2 bytes
		}).Once()

	client := NewSerialClient(serialPort)

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Equal(t, exampleFC1RTUResponse(), response)
	assert.NoError(t, err)

	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_contextCancelAfterFirstRead(t *testing.T) {
	serialPort := new(serialMock)
	ctx, cancel := context.WithCancel(context.Background())

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Read", mock.Anything).
		Return(5, nil).
		Run(func(args mock.Arguments) {
			b := args.Get(0).([]byte)
			copy(b, []byte{0x10, 0x1, 0x2, 0x1, 0x2})
			cancel()
		})

	client := NewSerialClient(serialPort)

	response, err := client.Do(ctx, exampleFC1RTURequest())

	assert.Nil(t, response)
	assert.EqualError(t, err, context.Canceled.Error())
	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_RequestShouldBeSet(t *testing.T) {
	serialPort := new(serialMock)

	client := NewSerialClient(serialPort)

	response, err := client.Do(context.Background(), nil)

	assert.Nil(t, response)
	assert.EqualError(t, err, "request can not be nil")
	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_SerialPortShouldBeSet(t *testing.T) {
	client := NewSerialClient(nil)

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Nil(t, response)
	assert.EqualError(t, err, "serial port is not set")
}

func TestSerialClient_Do_writeError(t *testing.T) {
	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).
		Once().
		Return(0, errors.New("write error"))
	serialPort.On("Flush").Once().Return(nil)

	client := NewSerialClient(serialPort)

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Nil(t, response)
	assert.EqualError(t, err, "write error")

	var target *ClientError
	assert.True(t, errors.As(err, &target))

	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_unknownReadError(t *testing.T) {
	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Flush").Once().Return(nil)
	serialPort.On("Read", mock.Anything).
		Return(0, io.ErrUnexpectedEOF)

	client := NewSerialClient(serialPort)

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Nil(t, response)
	assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())
	serialPort.AssertExpectations(t)
}

func TestSerialClient_Do_ReadMoreBytesThanPacketCanBe(t *testing.T) {
	serialPort := new(serialMock)

	serialPort.On("Write", []byte{0x10, 0x1, 0x0, 0xc8, 0x0, 0x9, 0x7e, 0xb3}).Once().Return(0, nil)
	serialPort.On("Flush").Once().Return(nil)
	serialPort.On("Read", mock.Anything).
		Return(tcpPacketMaxLen+1, nil)

	client := NewSerialClient(serialPort)

	response, err := client.Do(context.Background(), exampleFC1RTURequest())

	assert.Nil(t, response)
	assert.EqualError(t, err, "received more bytes than valid Modbus packet size can be")

	var target *ClientError
	assert.True(t, errors.As(err, &target))

	serialPort.AssertExpectations(t)
}
