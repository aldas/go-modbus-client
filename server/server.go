package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/aldas/go-modbus-client/packet"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	readTimeout  = 5 * time.Millisecond
	writeTimeout = 50 * time.Millisecond
	idleTimeout  = 25 * time.Second
)

var (
	// ErrServerClosed is returned when server context is ended (by shutdown)
	ErrServerClosed = errors.New("modbus server closed")
	// ErrServerIdleTimeout is returned when server closes connection that has been idle too long
	ErrServerIdleTimeout = errors.New("modbus server closed idle connection")
)

// PacketAssembler is called when server reads data from client connection. Is responsible for assembling data read
// from the connection to whole modbus packet.
//
// return with closeConnection=true when you are done sending and want to close connection
type PacketAssembler interface {
	ReceiveRead(ctx context.Context, received []byte, bytesRead int) (response []byte, closeConnection bool)
}

// RawReadTracer is PacketAssembler optional interface that is called for each Read from connection to allow tracing read results
type RawReadTracer interface {
	Read(data []byte, n int, err error)
}

// ModbusHandler calls Handle method when it has received enough data to be parsed into Modbus packet.
type ModbusHandler interface {
	Handle(ctx context.Context, received packet.Request) (packet.Response, error)
}

// Server simple TCP server implementation for server to serve modbus packets.
// Each connection is handled in separate goroutine, in which panics are recovered.
//
// Public fields are not designed to be goroutine safe. Do not mutate after server has been started
type Server struct {
	mu                    sync.RWMutex
	listener              net.Listener // for simplicity, we only allow serving one listener
	isShutdown            atomic.Bool
	activeConnections     map[*connection]struct{}
	activeConnectionCount atomic.Int64

	// WriteTimeout is amount of time writing the request can take after it errors out
	WriteTimeout time.Duration
	// ReadTimeout is amount of time reading the response can take
	ReadTimeout time.Duration

	// AssemblerCreatorFunc creates Assembler for each connetion to assemble different read byte fragments into complete
	// modbus packet. Could have different implementations for TCP or RTU packets
	AssemblerCreatorFunc func(handler ModbusHandler) PacketAssembler

	// OnServeFunc allows capturing listener address just before server starts to accepting connections. This is useful
	// for testing when listener is started with random port `:0`.
	OnServeFunc func(addr net.Addr)

	OnErrorFunc func(err error)

	// OnAcceptConnFunc is called when server accepts new connection. When method returns an error the connection will be closed.
	// connectionCount indicated currently active connection count.
	//
	// This is where firewall rules and other limits can be implemented
	OnAcceptConnFunc func(ctx context.Context, remoteAddr net.Addr, connectionCount uint64) error

	// OnCloseConnFunc is called at the end of connection. isServerShutdown indicated if method is called at server shutdown.
	OnCloseConnFunc func(ctx context.Context, remoteAddr net.Addr, isServerShutdown bool)
}

type connection struct {
	conn           net.Conn
	isBeingHandled atomic.Bool
	assembler      PacketAssembler

	writeTimeout time.Duration
	readTimeout  time.Duration

	onErrorFunc func(error)
}

// ListenAndServe starts accepting connection on given address and handles received data with handler function.
// Method blocks until context is cancelled
func (s *Server) ListenAndServe(ctx context.Context, address string, handler ModbusHandler) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("modbus listnener creation error: %w", err)
	}
	return s.serve(ctx, listener, handler)
}

// Serve accepts connections from listener and handles received data with handler function.
// Method blocks until context is cancelled
func (s *Server) Serve(ctx context.Context, listener net.Listener, handler ModbusHandler) error {
	return s.serve(ctx, listener, handler)
}

// ContextRemoteAddr is context.Context value containing clients remote address
type ContextRemoteAddr struct{}

func (s *Server) serve(ctx context.Context, listener net.Listener, handler ModbusHandler) error {
	if s.AssemblerCreatorFunc == nil {
		s.AssemblerCreatorFunc = func(handler ModbusHandler) PacketAssembler {
			return &ModbusTCPAssembler{Handler: handler}
		}
	}
	onErrorFunc := s.OnErrorFunc
	if onErrorFunc == nil {
		onErrorFunc = func(err error) {
			log.Printf("modbus server connection error: %v", err)
		}
	}
	if s.OnServeFunc != nil {
		// when listener is started with ":0" (random port) this will be helpful knowing where to connect
		// and if server is listening already
		s.OnServeFunc(listener.Addr())
	}

	s.listener = listener
	l := onceCloseListener{Listener: listener}
	defer l.Close()

	for {
		netConn, err := l.Accept()
		if err != nil {
			if s.isShutdown.Load() {
				return ErrServerClosed
			}
			return err
		}

		if s.OnAcceptConnFunc != nil {
			if err := s.OnAcceptConnFunc(ctx, netConn.RemoteAddr(), uint64(s.activeConnectionCount.Load()+1)); err != nil {
				if err := netConn.Close(); err != nil {
					onErrorFunc(fmt.Errorf("connection.close error, err: %w", err))
				}
				continue
			}
		}

		select {
		case <-ctx.Done():
			return ErrServerClosed
		default:
		}

		cCtx := context.WithValue(ctx, ContextRemoteAddr{}, netConn.RemoteAddr())
		c := &connection{
			conn:           netConn,
			isBeingHandled: atomic.Bool{},
			assembler:      s.AssemblerCreatorFunc(handler),
			writeTimeout:   s.WriteTimeout,
			readTimeout:    s.ReadTimeout,
			onErrorFunc:    onErrorFunc,
		}
		s.trackConn(c, true)
		go func(ctx context.Context, conn *connection) {
			defer func() {
				if rec := recover(); rec != nil {
					conn.onErrorFunc(fmt.Errorf("recovered panic in handler, %v", rec))
				}
				if err := conn.conn.Close(); err != nil {
					conn.onErrorFunc(fmt.Errorf("failed to close handler connection, err: %w", err))
				}
				s.trackConn(c, false)
				if s.OnAcceptConnFunc != nil {
					s.OnCloseConnFunc(ctx, conn.conn.RemoteAddr(), s.isShutdown.Load())
				}
			}()
			conn.handle(ctx)
		}(cCtx, c)
	}
}

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceCloseListener) close() {
	oc.closeErr = oc.Listener.Close()
}

func (s *Server) trackConn(c *connection, isAdd bool) {
	// this is how http.Server does it
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeConnections == nil {
		s.activeConnections = make(map[*connection]struct{})
	}
	if isAdd {
		s.activeConnections[c] = struct{}{}
		s.activeConnectionCount.Add(1)
	} else {
		delete(s.activeConnections, c)
		s.activeConnectionCount.Add(-1)
	}
}

func (c *connection) handle(ctx context.Context) {
	cCtx, cCancel := context.WithCancel(ctx)
	defer cCancel()

	rTimeout := readTimeout
	if c.readTimeout > 0 {
		rTimeout = c.readTimeout
	}
	wTimeout := writeTimeout
	if c.writeTimeout > 0 {
		wTimeout = c.writeTimeout
	}

	rrt, debugRawRead := c.assembler.(RawReadTracer)
	conn := c.conn
	lastReceived := time.Now()
	received := make([]byte, 300)
	for {
		select {
		case <-cCtx.Done():
			return
		default:
		}

		_ = conn.SetReadDeadline(time.Now().Add(rTimeout))
		n, err := conn.Read(received)
		if debugRawRead {
			rrt.Read(received[0:n], n, err)
		}
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			if !errors.Is(err, io.EOF) {
				c.onErrorFunc(err)
			}
			return // when read fails due some unknown error we close connection
		}
		if n > 0 {
			lastReceived = time.Now()
		} else if time.Now().Sub(lastReceived) > idleTimeout {
			c.onErrorFunc(ErrServerIdleTimeout)
			return // close idle connection
		} else {
			continue // nothing read and not idle yet
		}

		c.isBeingHandled.Store(true)
		toSend, closeConn := c.assembler.ReceiveRead(cCtx, received[0:n], n)
		if toSend != nil {
			_ = conn.SetWriteDeadline(time.Now().Add(wTimeout))
			if _, err := conn.Write(toSend); err != nil {
				c.onErrorFunc(err)
				return // when write fails to client we close connection
			}
		}
		c.isBeingHandled.Store(false)
		if closeConn {
			return
		}
	}
}

// Addr returns currently running server address
func (s *Server) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listener.Addr()
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
// Works similarly as `http.Server.Shutdown()`
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isShutdown.Store(true)

	err := s.listener.Close()

	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()
	for {
		allIdle := true
		for c := range s.activeConnections {
			if c.isBeingHandled.Load() {
				allIdle = false
				continue
			}
			(*c).conn.Close()
			delete(s.activeConnections, c)
		}
		if allIdle {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(50 * time.Millisecond)
		}
	}
}
