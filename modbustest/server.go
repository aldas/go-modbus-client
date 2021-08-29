package modbustest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// ErrServerClosed is returned when server context is ended
var ErrServerClosed = errors.New("modbus test server closed")

// Server simple TCP server implementation for testing modbus packets
type Server struct {
	mu       sync.RWMutex
	listener net.Listener // for simplicity we only allow serving one listener

	OnServeAddrChan chan<- string
	OnErrorFunc     func(err error)
}

// ReadHandler is function called when server reads bytes from client connection.
//
// Handler can be called even if no bytes are read from connection. In that case bytesRead==0.
// This is so you can emulate writing modbus packet as multiple fragments.
// return with closeConnection=true when you are done sending fragments and want to close connection
type ReadHandler func(received []byte, bytesRead int) (response []byte, closeConnection bool)

// ListenAndServe starts accepting connection on given address and handles received data with handler function.
// Method blocks until context is cancelled
func (s *Server) ListenAndServe(ctx context.Context, address string, handler ReadHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("modbustest listnener creation error: %w", err)
	}
	return s.serve(ctx, listener, handler)
}

// Serve accepts connections from listener and handles received data with handler function.
// Method blocks until context is cancelled
func (s *Server) Serve(ctx context.Context, listener net.Listener, handler ReadHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.serve(ctx, listener, handler)
}

func (s *Server) serve(ctx context.Context, listener net.Listener, handler ReadHandler) error {
	if handler == nil {
		return errors.New("handler can not be nil")
	}
	if s.OnServeAddrChan != nil {
		// when listener is started with ":0" (random port) this chan will be helpful knowing where to connect
		// and if server is listening already
		s.OnServeAddrChan <- listener.Addr().String()
	}
	onErrorFunc := s.OnErrorFunc
	if onErrorFunc == nil {
		onErrorFunc = func(err error) {
			log.Printf("modbus test server connection error: %v", err)
		}
	}

	s.listener = listener

	l := onceCloseListener{Listener: listener}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ErrServerClosed
		default:
		}
		go handleConnection(ctx, conn, handler, onErrorFunc)
	}
}

func handleConnection(ctx context.Context, conn net.Conn, handler ReadHandler, onErrorFunc func(error)) {
	defer conn.Close()
	received := make([]byte, 300)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Microsecond)) // max 0.5ms block time for read per iteration
		n, err := conn.Read(received)
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			if !errors.Is(err, io.EOF) {
				onErrorFunc(err)
			}
			return // when read fails due some unknown error we close connection
		}
		// NB: handler can be called even if client did not send anything. It is up to developer to handle that case.
		toSend, closeConn := handler(received[:n], n)
		if toSend != nil {
			_ = conn.SetWriteDeadline(time.Now().Add(500 * time.Microsecond))
			if _, err := conn.Write(toSend); err != nil {
				onErrorFunc(err)
				return // when write fails to client we close connection
			}
		}
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
