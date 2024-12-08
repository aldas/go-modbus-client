package modbus

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// tcpPacketMaxLen is maximum length in bytes that valid Modbus TCP packet can be
	//
	// Quote from MODBUS Application Protocol Specification V1.1b3:
	//   The size of the MODBUS PDU is limited by the size constraint inherited from the first
	//   MODBUS implementation on Serial Line network (max. RS485 ADU = 256 bytes).
	//   Therefore:
	//   MODBUS PDU for serial line communication = 256 - Server address (1 byte) - CRC (2bytes) = 253 bytes.
	//   Consequently:
	//   RS232 / RS485 ADU = 253 bytes + Server address (1 byte) + CRC (2 bytes) = 256 bytes.
	//   TCP MODBUS ADU = 253 bytes + MBAP (7 bytes) = 260 bytes.
	tcpPacketMaxLen = 7 + 253 // 2 trans id + 2 proto + 2 pdu len + 1 unit id + 253 max data len
	rtuPacketMaxLen = 256     // 1 unit id + 253 max data len + 2 crc

	defaultWriteTimeout   = 1 * time.Second
	defaultReadTimeout    = 2 * time.Second
	defaultConnectTimeout = 1 * time.Second
)

// ErrPacketTooLong is error indicating that modbus server sent amount of data that is bigger than any modbus packet could be
var ErrPacketTooLong = &ClientError{Err: errors.New("received more bytes than valid Modbus packet size can be")}

// ErrClientNotConnected is error indicating that Client has not yet connected to the modbus server
var ErrClientNotConnected = &ClientError{Err: errors.New("client is not connected")}

// Client provides mechanisms to send requests to modbus server over network connection
type Client struct {
	timeNow func() time.Time

	// writeTimeout is total amount of time writing the request can take after client returns error
	writeTimeout time.Duration
	// readTimeout is total amount of time reading the response can take before client returns error
	readTimeout time.Duration

	dialContextFunc     func(ctx context.Context, address string) (net.Conn, error)
	asProtocolErrorFunc func(data []byte) error
	parseResponseFunc   func(data []byte) (packet.Response, error)

	mu      sync.RWMutex
	address string
	conn    net.Conn
	hooks   ClientHooks
}

// ClientHooks allows to log bytes send/received by client.
// NB: Do not modify given slice - it is not a copy.
type ClientHooks interface {
	BeforeWrite(toWrite []byte)
	AfterEachRead(received []byte, n int, err error)
	BeforeParse(received []byte)
}

// ClientConfig is configuration for Client
type ClientConfig struct {
	// WriteTimeout is total amount of time writing the request can take after client returns error
	WriteTimeout time.Duration
	// ReadTimeout is total amount of time reading the response can take before client returns error
	ReadTimeout time.Duration

	DialContextFunc     func(ctx context.Context, address string) (net.Conn, error)
	AsProtocolErrorFunc func(data []byte) error
	ParseResponseFunc   func(data []byte) (packet.Response, error)

	Hooks ClientHooks
}

func defaultClient(conf ClientConfig) *Client {
	c := &Client{
		timeNow:      time.Now,
		writeTimeout: defaultWriteTimeout,
		readTimeout:  defaultReadTimeout,

		dialContextFunc: dialContext,
		// TCP is our default protocol
		asProtocolErrorFunc: packet.AsTCPErrorPacket,
		parseResponseFunc:   packet.ParseTCPResponse,
	}

	if conf.WriteTimeout > 0 {
		c.writeTimeout = conf.WriteTimeout
	}
	if conf.ReadTimeout > 0 {
		c.readTimeout = conf.ReadTimeout
	}
	if conf.DialContextFunc != nil {
		c.dialContextFunc = conf.DialContextFunc
	}
	if conf.AsProtocolErrorFunc != nil {
		c.asProtocolErrorFunc = conf.AsProtocolErrorFunc
	}
	if conf.ParseResponseFunc != nil {
		c.parseResponseFunc = conf.ParseResponseFunc
	}
	if conf.Hooks != nil {
		c.hooks = conf.Hooks
	}
	return c
}

// NewTCPClient creates new instance of Modbus Client for Modbus TCP protocol
func NewTCPClient() *Client {
	return NewTCPClientWithConfig(ClientConfig{})
}

// NewTCPClientWithConfig creates new instance of Modbus Client for Modbus TCP protocol with given configuration options
func NewTCPClientWithConfig(conf ClientConfig) *Client {
	client := defaultClient(conf)
	client.asProtocolErrorFunc = packet.AsTCPErrorPacket
	client.parseResponseFunc = packet.ParseTCPResponse
	return client
}

// NewRTUClient creates new instance of Modbus Client for Modbus RTU protocol
func NewRTUClient() *Client {
	return NewRTUClientWithConfig(ClientConfig{})
}

// NewRTUClientWithConfig creates new instance of Modbus Client for Modbus RTU protocol with given configuration options
func NewRTUClientWithConfig(conf ClientConfig) *Client {
	client := defaultClient(conf)
	client.asProtocolErrorFunc = packet.AsRTUErrorPacket
	client.parseResponseFunc = packet.ParseRTUResponseWithCRC
	return client
}

// NewClient creates new instance of Modbus Client with given configuration options
func NewClient(conf ClientConfig) *Client {
	return defaultClient(conf)
}

// Connect opens network connection to Client to server. Context lifetime is only meant for this call.
// ctx is to be used for to cancel connection attempt.
//
// `address` should be formatted as url.URL scheme `[scheme:][//[userinfo@]host][/]path[?query]`
// Example:
// * `127.0.0.1:502` (library defaults to `tcp` as scheme)
// * `udp://127.0.0.1:502`
// * `/dev/ttyS0?BaudRate=4800`
// * `file:///dev/ttyUSB?BaudRate=4800`
func (c *Client) Connect(ctx context.Context, address string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := c.dialContextFunc(ctx, address)
	if err != nil {
		return err
	}
	c.conn = conn
	c.address = address
	return nil
}

func dialContext(ctx context.Context, address string) (net.Conn, error) {
	dialer := &net.Dialer{
		// Timeout is the maximum amount of time a dial will wait for a connect to complete.
		Timeout: defaultConnectTimeout,
		// KeepAlive specifies the interval between keep-alive probes for an active network connection.
		KeepAlive: 15 * time.Second,
	}
	network, addr := addressExtractor(address)
	return dialer.DialContext(ctx, network, addr)
}

func addressExtractor(address string) (string, string) {
	network, addr, ok := strings.Cut(address, "://")
	if !ok {
		return "tcp", address
	}
	return network, addr
}

// Close closes network connection to Modbus server
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// ClientError indicates errors returned by Client that network related and are possibly retryable
type ClientError struct {
	Err error
}

// Error returns contained error message
func (e *ClientError) Error() string { return e.Err.Error() }

// Unwrap allows unwrapping errors with errors.Is and errors.As
func (e *ClientError) Unwrap() error { return e.Err }

// Do sends given Modbus request to modbus server and returns parsed Response.
// ctx is to be used for to cancel connection attempt.
// On modbus exception nil is returned as response and error wraps value of type packet.ErrorResponseTCP or packet.ErrorResponseRTU
// User errors.Is and errors.As to check if error wraps packet.ErrorResponseTCP or packet.ErrorResponseRTU
func (c *Client) Do(ctx context.Context, req packet.Request) (packet.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req == nil {
		return nil, errors.New("request can not be nil")
	}
	if c.conn == nil {
		return nil, ErrClientNotConnected
	}

	resp, err := c.do(ctx, req.Bytes(), req.ExpectedResponseLength())
	if err != nil {
		return nil, err
	}
	if c.hooks != nil {
		c.hooks.BeforeParse(resp)
	}
	return c.parseResponseFunc(resp)
}

func (c *Client) do(ctx context.Context, data []byte, expectedLen int) ([]byte, error) {
	if err := c.conn.SetWriteDeadline(c.timeNow().Add(c.writeTimeout)); err != nil {
		return nil, err
	}
	if c.hooks != nil {
		c.hooks.BeforeWrite(data)
	}
	if _, err := c.conn.Write(data); err != nil {
		return nil, &ClientError{Err: err}
	}

	// make buffer a little bit bigger than would be valid to see problems when somehow more bytes are sent
	const maxBytes = tcpPacketMaxLen + 10
	received := [maxBytes]byte{}
	total := 0
	readTimeout := time.After(c.readTimeout)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-readTimeout:
			return nil, &ClientError{Err: errors.New("total read timeout exceeded")}
		default:
		}

		_ = c.conn.SetReadDeadline(c.timeNow().Add(500 * time.Microsecond)) // max 0.5ms block time for read per iteration
		n, err := c.conn.Read(received[total:maxBytes])
		if c.hooks != nil {
			c.hooks.AfterEachRead(received[total:total+n], n, err)
		}
		// on read errors we do not return immediately as for:
		// os.ErrDeadlineExceeded - we set new deadline on next iteration
		// io.EOF - we check if read + received is enough to form complete packet
		if err != nil && !(errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF)) {
			return nil, &ClientError{Err: err}
		}
		total += n
		if total > tcpPacketMaxLen {
			return nil, ErrPacketTooLong
		}
		// check if we have exactly the error packet. Error packets are shorter than regulars packets
		if errPacket := c.asProtocolErrorFunc(received[0:total]); errPacket != nil {
			return nil, &ClientError{Err: errPacket}
		}
		if total >= expectedLen {
			break
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	if total == 0 {
		return nil, &ClientError{Err: errors.New("no bytes received")}
	}

	result := make([]byte, total)
	copy(result, received[:total])
	return result, nil
}
