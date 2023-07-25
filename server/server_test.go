package server

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestRequestToServer(t *testing.T) {
	mbs := new(mbServer)

	serverAddrCh := make(chan string)
	s := Server{
		OnServeFunc: func(addr net.Addr) {
			serverAddrCh <- addr.String()
		},
		OnErrorFunc:  nil,
		OnAcceptFunc: nil,
	}

	tCtx, tCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer tCancel()
	ctx, cancel := signal.NotifyContext(tCtx, os.Kill, os.Interrupt)
	defer cancel()

	// we start the server and listen for incoming connections/data in separate goroutine. ListenAndServe is blocking call.
	go func() {
		err := s.ListenAndServe(ctx, "localhost:5020", mbs)
		if err != nil && !errors.Is(err, ErrServerClosed) {
			assert.NoError(t, err)
		}
	}()

	select {
	case <-ctx.Done():
		return
	case serverAddr := <-serverAddrCh: // wait for server to "start"
		register11, err := doRequest(ctx, serverAddr)
		assert.NoError(t, err)
		assert.Equal(t, uint16(258), register11)
	}

	graceful, gCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer gCancel()
	if err := s.Shutdown(graceful); err != nil {
		assert.NoError(t, err)
	}
}

func doRequest(ctx context.Context, serverAddress string) (uint16, error) {
	client := modbus.NewTCPClientWithConfig(modbus.ClientConfig{
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	})
	if err := client.Connect(ctx, serverAddress); err != nil {
		return 0, err
	}
	defer client.Close()

	unitID := uint8(1)
	startAddress := uint16(10)
	quantity := uint16(2)
	req, err := packet.NewReadHoldingRegistersRequestTCP(unitID, startAddress, quantity)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(ctx, req)
	if err != nil {
		return 0, err
	}

	registers, err := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(startAddress)
	if err != nil {
		return 0, err
	}

	return registers.Uint16(11)
}

type mbServer struct {
}

func (s *mbServer) Handle(ctx context.Context, received packet.Request) (packet.Response, error) {
	switch req := received.(type) {
	case *packet.ReadHoldingRegistersRequestTCP:
		p := packet.ReadHoldingRegistersResponseTCP{
			MBAPHeader: req.MBAPHeader,
			ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
				UnitID:          req.UnitID,
				RegisterByteLen: 4,
				Data:            []byte{0x0, 0x1, 0x01, 0x02}, // register[0] = 0x0001, register[1] = 0x0102
			},
		}
		return p, nil
	}
	return nil, packet.NewErrorParseTCP(packet.ErrIllegalFunction, "nope")
}

func TestServer_Addr(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if !assert.NoError(t, err) {
		return
	}
	defer listener.Close()

	lAddr := listener.Addr().String()

	s := Server{
		listener: listener,
	}
	assert.Equal(t, lAddr, s.Addr().String())
}
