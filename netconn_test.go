package modbus

import (
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

// slowTestConn, noopConn, dummyAddr are copied from: https://github.com/golang/go/blob/a8a85281caf21831ee51ea8c879cbba94bcce256/src/net/http/serve_test.go#L2146

// slowTestConn is a net.Conn that provides a means to simulate parts of a
// request being received piecemeal. Deadlines can be set and enforced in both
// Read and Write.
type slowTestConn struct {
	// over multiple calls to Read, time.Durations are slept, strings are read.
	script []interface{}
	closec chan bool

	mu     sync.Mutex // guards rd/wd
	rd, wd time.Time  // read, write deadline
	noopConn
}

func (c *slowTestConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *slowTestConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rd = t
	return nil
}

func (c *slowTestConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wd = t
	return nil
}

func (c *slowTestConn) Read(b []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
restart:
	if !c.rd.IsZero() && time.Now().After(c.rd) {
		return 0, syscall.ETIMEDOUT
	}
	if len(c.script) == 0 {
		return 0, io.EOF
	}

	switch cue := c.script[0].(type) {
	case time.Duration:
		if !c.rd.IsZero() {
			// If the deadline falls in the middle of our sleep window, deduct
			// part of the sleep, then return a timeout.
			if remaining := time.Until(c.rd); remaining < cue {
				c.script[0] = cue - remaining
				time.Sleep(remaining)
				return 0, syscall.ETIMEDOUT
			}
		}
		c.script = c.script[1:]
		time.Sleep(cue)
		goto restart

	case string:
		n = copy(b, cue)
		// If cue is too big for the buffer, leave the end for the next Read.
		if len(cue) > n {
			c.script[0] = cue[n:]
		} else {
			c.script = c.script[1:]
		}

	default:
		panic("unknown cue in slowTestConn script")
	}

	return
}

func (c *slowTestConn) Close() error {
	select {
	case c.closec <- true:
	default:
	}
	return nil
}

func (c *slowTestConn) Write(b []byte) (int, error) {
	if !c.wd.IsZero() && time.Now().After(c.wd) {
		return 0, syscall.ETIMEDOUT
	}
	return len(b), nil
}

type noopConn struct{}

func (noopConn) LocalAddr() net.Addr                { return dummyAddr("local-addr") }
func (noopConn) RemoteAddr() net.Addr               { return dummyAddr("remote-addr") }
func (noopConn) SetDeadline(t time.Time) error      { return nil }
func (noopConn) SetReadDeadline(t time.Time) error  { return nil }
func (noopConn) SetWriteDeadline(t time.Time) error { return nil }

type dummyAddr string

func (a dummyAddr) Network() string {
	return string(a)
}

func (a dummyAddr) String() string {
	return string(a)
}
