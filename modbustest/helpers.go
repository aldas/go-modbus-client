package modbustest

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/aldas/go-modbus-client/server"
	"log"
	"net"
	"time"
)

// RunServerOnRandomPort is low level helper function for testing modbus packets. Method starts server in separate
// goroutine and runs it until given context is cancelled. Given PacketAssembler is used by server to handle incoming data.
func RunServerOnRandomPort(
	ctx context.Context,
	handler func(received []byte, bytesRead int) (response []byte, closeConnection bool),
) (string, error) {
	addrChan := make(chan string)
	serverErrChan := make(chan error)

	rr := &rawReader{
		handler: handler,
	}
	srv := server.Server{
		AssemblerCreatorFunc: func(_ server.ModbusHandler) server.PacketAssembler {
			return rr
		},
		OnServeFunc: func(addr net.Addr) {
			addrChan <- addr.String()
		},
	}
	go func() {
		if err := srv.ListenAndServe(ctx, ":0", rr); err != nil {
			log.Printf("server err: %v", err)
			serverErrChan <- err
		}
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		return "", errors.New("timeout when waiting for test server startup")
	case err := <-serverErrChan:
		return "", err
	case addr := <-addrChan:
		return addr, nil
	}
}

type rawReader struct {
	handler func(received []byte, bytesRead int) (response []byte, closeConnection bool)
}

func (r *rawReader) Handle(ctx context.Context, received packet.Request) (packet.Response, error) {
	panic("this is not called")
}

func (r *rawReader) ReceiveRead(ctx context.Context, received []byte, bytesRead int) (response []byte, closeConnection bool) {
	return r.handler(received, bytesRead)
}
