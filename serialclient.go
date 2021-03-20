package modbus

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"io"
	"os"
	"sync"
	"time"
)

// SerialClient provides mechanisms to send requests to modbus server over serial port
type SerialClient struct {
	// readTimeout is total amount of time reading the response can take before client returns error.
	// NB: if you have set long reading timeout on your serial port implementation this timeout will not help you
	// as it works for cases when there are multiple read calls.
	readTimeout time.Duration

	asProtocolErrorFunc func(data []byte) error
	parseResponseFunc   func(data []byte) (packet.Response, error)

	mu         sync.RWMutex
	isFlusher  bool
	serialPort io.ReadWriteCloser
	hooks      ClientHooks
}

// NewSerialClient creates new instance of Modbus SerialClient for Modbus RTU protocol
func NewSerialClient(serialPort io.ReadWriteCloser, opts ...SerialClientOptionFunc) *SerialClient {
	_, isFlusher := serialPort.(Flusher)

	client := &SerialClient{
		readTimeout:         defaultReadTimeout,
		asProtocolErrorFunc: packet.AsRTUErrorPacket,
		parseResponseFunc:   packet.ParseRTUResponseWithCRC,
		serialPort:          serialPort,
		hooks:               nil,
		isFlusher:           isFlusher,
	}
	if opts != nil {
		for _, o := range opts {
			o(client)
		}
	}
	return client
}

// SerialClientOptionFunc is options type for NewSerialClient function
type SerialClientOptionFunc func(c *SerialClient)

// WithSerialHooks is option to set hooks for SerialClient
func WithSerialHooks(hooks ClientHooks) func(c *SerialClient) {
	return func(c *SerialClient) {
		c.hooks = hooks
	}
}

// WithSerialReadTimeout is option to for setting total timeout for reading the whole packet
func WithSerialReadTimeout(readTimeout time.Duration) func(c *SerialClient) {
	return func(c *SerialClient) {
		c.readTimeout = readTimeout
	}
}

// Do sends given Modbus request to modbus server and returns parsed Response.
// ctx is to be used for to cancel connection attempt.
func (c *SerialClient) Do(ctx context.Context, req packet.Request) (packet.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req == nil {
		return nil, errors.New("request can not be nil")
	}
	if c.serialPort == nil {
		return nil, errors.New("serial port is not set")
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

func (c *SerialClient) do(ctx context.Context, data []byte, expectedLen int) ([]byte, error) {
	if c.hooks != nil {
		c.hooks.BeforeWrite(data)
	}
	if _, err := c.serialPort.Write(data); err != nil {
		if err := c.flush(); err != nil {
			return nil, err
		}
		return nil, err
	}

	// make buffer a little bit bigger than would be valid to see problems when somehow more bytes are sent
	const maxBytes = rtuPacketMaxLen + 10
	received := [maxBytes]byte{}
	total := 0
	readTimeout := time.After(c.readTimeout)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-readTimeout:
			return nil, errors.New("total read timeout exceeded")
		default:
		}

		n, err := c.serialPort.Read(received[total:maxBytes])
		if c.hooks != nil {
			c.hooks.AfterEachRead(received[total:total+n], n, err)
		}
		// on read errors we do not return immediately as for:
		// os.ErrDeadlineExceeded - we set new deadline on next iteration
		// io.EOF - we check if read + received is enough to form complete packet
		if err != nil && !(errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF)) {
			if err := c.flush(); err != nil {
				return nil, err
			}
			return nil, err
		}
		total += n
		if total > rtuPacketMaxLen {
			if err := c.flush(); err != nil {
				return nil, err
			}
			return nil, ErrPacketTooLong
		}
		// check if we have exactly the error packet. Error packets are shorter than regulars packets
		if errPacket := c.asProtocolErrorFunc(received[0:total]); errPacket != nil {
			if err := c.flush(); err != nil {
				return nil, err
			}
			return nil, errPacket
		}
		if total >= expectedLen {
			if err := c.flush(); err != nil {
				return nil, err
			}
			break
		}
	}
	if total == 0 {
		return nil, errors.New("no bytes received")
	}

	result := make([]byte, total)
	copy(result, received[:total])
	return result, nil
}

func (c *SerialClient) flush() error {
	if !c.isFlusher {
		return nil
	}
	return c.serialPort.(Flusher).Flush()
}

// Flusher is interface for flushing unread/unwritten data from serial port buffer
type Flusher interface {
	Flush() error
}