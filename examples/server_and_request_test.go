package examples_test

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/aldas/go-modbus-client/server"
	"log"
	"net"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestRequestToServer(t *testing.T) {
	mbs := new(mbHandler)

	serverAddrCh := make(chan string)
	s := server.Server{
		// OnServeFunc is useful integration tests when in situations where actual server is spun up to serve requests
		// in that case it is useful it start it on random (":0") port. This callback is run just before server starts
		// listening for new connections
		OnServeFunc: func(addr net.Addr) {
			serverAddrCh <- addr.String()
			log.Printf("listening on: %v\n", addr.String())
		},
	}

	tCtx, tCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer tCancel()
	ctx, cancel := signal.NotifyContext(tCtx, os.Kill, os.Interrupt)
	defer cancel()

	// we start the server and listen for incoming connections/data in separate goroutine. ListenAndServe is blocking call.
	go func() {
		err := s.ListenAndServe(ctx, "localhost:0", mbs)
		if err != nil && !errors.Is(err, server.ErrServerClosed) {
			log.Printf("ListenAndServe end: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		return
	case serverAddr := <-serverAddrCh: // wait for server to "start"
		// do the FC03 request
		if err := doRequest(ctx, serverAddr); err != nil {
			log.Printf("doRequest err: %v\n", err)
			return
		}
	}

	// gracefully shut down the server.
	// We could have used here:
	//<-ctx.Done()
	// to wait for ctrl+c or kill signal but for example we close the server after request has been done
	graceful, gCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer gCancel()
	if err := s.Shutdown(graceful); err != nil {
		log.Printf("Shutdown end: %v", err)
	}
}

func doRequest(ctx context.Context, serverAddress string) error {
	client := modbus.NewTCPClientWithConfig(modbus.ClientConfig{
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	})
	if err := client.Connect(ctx, serverAddress); err != nil {
		return err
	}
	defer client.Close()

	unitID := uint8(1)
	startAddress := uint16(10)
	quantity := uint16(2)
	req, err := packet.NewReadHoldingRegistersRequestTCP(unitID, startAddress, quantity)
	if err != nil {
		return err
	}

	resp, err := client.Do(ctx, req)
	if err != nil {
		return err
	}

	registers, err := resp.(*packet.ReadHoldingRegistersResponseTCP).AsRegisters(startAddress)
	if err != nil {
		return err
	}
	uint16Var, err := registers.Uint16(11) // extract uint16 value from register 11
	if err != nil {
		return err
	}
	log.Printf("Received as register 11 value: %v (hex: %X)\n", uint16Var, uint16Var)

	return nil
}

type mbHandler struct {
}

func (s *mbHandler) Handle(ctx context.Context, received packet.Request) (packet.Response, error) {
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
