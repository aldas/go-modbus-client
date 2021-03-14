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

	defaultWriteTimeout   = 2 * time.Second
	defaultReadTimeout    = 4 * time.Second
	defaultConnectTimeout = 5 * time.Second
)

// ErrPacketTooLong is error indicating that modbus server sent amount of data that is bigger than any modbus packet could be
var ErrPacketTooLong = errors.New("received more bytes than valid Modbus packet size can be")

// Client provides mechanisms to send requests to modbus server
type Client struct {
	timeNow func() time.Time

	writeTimeout time.Duration
	readTimeout  time.Duration

	dialContextFunc     func(ctx context.Context, address string) (net.Conn, error)
	asProtocolErrorFunc func(data []byte) error
	parseResponseFunc   func(data []byte) (packet.Response, error)

	mu      sync.RWMutex
	address string
	conn    net.Conn // FIXME: maybe use `io.ReadWriteCloser` so we can use serial connection also here
	logger  ClientLogger
}

// ClientLogger allows to log bytes send/received by client.
// NB: Do not modify given slice - it is not a copy.
type ClientLogger interface {
	BeforeWrite(toWrite []byte)
	AfterEachRead(received []byte, n int, err error)
	BeforeParse(received []byte)
}

func defaultClient() *Client {
	return &Client{
		timeNow:      time.Now,
		writeTimeout: defaultWriteTimeout,
		readTimeout:  defaultReadTimeout,

		dialContextFunc: dialContext,
		// TCP is our default protocol
		asProtocolErrorFunc: packet.AsTCPErrorPacket,
		parseResponseFunc:   packet.ParseTCPResponse,
	}
}

// NewTCPClient creates new instance of Modbus Client for Modbus TCP protocol
func NewTCPClient(opts ...ClientOptionFunc) *Client {
	client := defaultClient()
	if opts != nil {
		for _, o := range opts {
			o(client)
		}
	}
	return client
}

// NewRTUClient creates new instance of Modbus Client for Modbus RTU protocol
func NewRTUClient(opts ...ClientOptionFunc) *Client {
	client := defaultClient()
	client.asProtocolErrorFunc = packet.AsRTUErrorPacket
	client.parseResponseFunc = packet.ParseRTUResponse
	// TODO: add CRC/noCRC check option

	if opts != nil {
		for _, o := range opts {
			o(client)
		}
	}
	return client
}

// NewClient creates new instance of Modbus Client with given options
func NewClient(opts ...ClientOptionFunc) *Client {
	client := defaultClient()
	if opts != nil {
		for _, o := range opts {
			o(client)
		}
	}
	return client
}

// ClientOptionFunc is options type for NewClient function
type ClientOptionFunc func(c *Client)

// WithProtocolErrorFunc is option to provide custom function for parsing error packet
func WithProtocolErrorFunc(errorFunc func(data []byte) error) func(c *Client) {
	return func(c *Client) {
		c.asProtocolErrorFunc = errorFunc
	}
}

// WithParseResponseFunc is option to provide custom function for parsing protocol packet
func WithParseResponseFunc(parseFunc func(data []byte) (packet.Response, error)) func(c *Client) {
	return func(c *Client) {
		c.parseResponseFunc = parseFunc
	}
}

// WithDialContextFunc is option to provide custom function for creating new connection
func WithDialContextFunc(dialContextFunc func(ctx context.Context, address string) (net.Conn, error)) func(c *Client) {
	return func(c *Client) {
		c.dialContextFunc = dialContextFunc
	}
}

// WithTimeouts is option to for setting writing packet or reading packet timeouts
func WithTimeouts(writeTimeout time.Duration, readTimeout time.Duration) func(c *Client) {
	return func(c *Client) {
		c.writeTimeout = writeTimeout
		c.readTimeout = readTimeout
	}
}

// WithLogger is option to set logger in client
func WithLogger(logger ClientLogger) func(c *Client) {
	return func(c *Client) {
		c.logger = logger
	}
}

// Connect opens network connection to Client to server. Context lifetime is only meant for this call.
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
	network := "tcp"
	addr := address
	if strings.HasPrefix(addr, "tcp4://") {
		network = "tcp4"
		addr = strings.TrimPrefix(addr, "tcp4://")
	} else if strings.HasPrefix(addr, "tcp6://") {
		network = "tcp6"
		addr = strings.TrimPrefix(addr, "tcp6://")
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

// Do sends given Modbus request to modbus server and returns parsed Response
func (c *Client) Do(ctx context.Context, req packet.Request) (packet.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req == nil {
		return nil, errors.New("request can not be nil")
	}
	if c.conn == nil {
		return nil, errors.New("client is not connected")
	}

	resp, err := c.do(ctx, req.Bytes(), req.ExpectedResponseLength())
	if err != nil {
		return nil, err
	}
	if c.logger != nil {
		c.logger.BeforeParse(resp)
	}
	return c.parseResponseFunc(resp)
}

func (c *Client) do(ctx context.Context, data []byte, expectedLen int) ([]byte, error) {
	if err := c.conn.SetWriteDeadline(c.timeNow().Add(c.writeTimeout)); err != nil {
		return nil, err
	}
	if c.logger != nil {
		c.logger.BeforeWrite(data)
	}
	if _, err := c.conn.Write(data); err != nil {
		return nil, err
	}

	received := [tcpPacketMaxLen]byte{}
	tmp := [tcpPacketMaxLen]byte{}
	total := 0
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		_ = c.conn.SetReadDeadline(c.timeNow().Add(500 * time.Microsecond)) // 0.5ms timeout for read per iteration
		n, err := c.conn.Read(tmp[:tcpPacketMaxLen])
		if c.logger != nil {
			c.logger.AfterEachRead(tmp[:n], n, err)
		}
		// on read errors we do not return immediately as for:
		// os.ErrDeadlineExceeded - we set new deadline on next iteration
		// io.EOF - we check if read + received is enough to form complete packet
		if err != nil && !(errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF)) {
			// TODO: call flush
			return nil, err
		}
		total += n
		// TODO: log trace bytes as hex
		if total > tcpPacketMaxLen {
			// TODO: call flush
			return nil, ErrPacketTooLong
		}
		if n > 0 {
			copy(received[total-n:], tmp[:n])
		}
		// check if we have exactly the error packet. Error packets are shorter than regulars packets
		if errPacket := c.asProtocolErrorFunc(received[0:total]); errPacket != nil {
			// TODO: call flush
			return nil, errPacket
		}
		if total >= expectedLen {
			// TODO: call flush if needed
			break
		}
		if errors.Is(err, io.EOF) {
			return nil, err
		}
	}
	if total == 0 {
		return nil, errors.New("no bytes received")
	}

	result := make([]byte, total)
	copy(result, received[:total])
	return result, nil
}

func (c *Client) flush() error {
	// fixme: implement and use when returning error from .do()
	return nil
}
