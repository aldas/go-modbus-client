package modbustest

import (
	"context"
	"errors"
	"log"
	"time"
)

// RunServerOnRandomPort is low level helper function for testing modbus packets. Method starts server in separate
// goroutine and runs it until given context is cancelled. Given ReadHandler is used by server to handle incoming data.
func RunServerOnRandomPort(ctx context.Context, handler ReadHandler) (string, error) {
	addrChan := make(chan string)
	serverErrChan := make(chan error)
	server := Server{OnServeAddrChan: addrChan}
	go func() {
		if err := server.ListenAndServe(ctx, ":0", handler); err != nil {
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
