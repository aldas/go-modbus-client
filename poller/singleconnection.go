package poller

import (
	"context"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"strings"
	"sync"
)

type singleConnectionPerAddress struct {
	mu            sync.Mutex
	clientPool    map[string]*sharedClient
	newClientFunc ClientConnectFunc
}

// NewSingleConnectionPerAddressClientFunc creates clients that limit themselves to single instance
// per server address and use mutexes to guard against parallel Client.Do invocations.
// Use-case for this is connections to serial devices where you can not run multiple requests in parallel.
func NewSingleConnectionPerAddressClientFunc(newClientFunc ClientConnectFunc) ClientConnectFunc {
	pool := &singleConnectionPerAddress{
		newClientFunc: newClientFunc,
		clientPool:    map[string]*sharedClient{},
	}
	return pool.NewClientFunc
}

func (scp *singleConnectionPerAddress) NewClientFunc(ctx context.Context, protocol modbus.ProtocolType, address string) (Client, error) {
	scp.mu.Lock()
	defer scp.mu.Unlock()

	// address could have params `tcp://192.168.1.2:502??invalid_addr=70`
	if i := strings.IndexByte(address, '?'); i != -1 {
		address = address[:i]
	}
	client, ok := scp.clientPool[address]
	if ok {
		return client, nil
	}

	orgClient, err := scp.newClientFunc(ctx, protocol, address)
	if err != nil {
		return nil, err
	}

	client = &sharedClient{
		mu:      sync.Mutex{},
		client:  orgClient,
		address: address,
		onClose: scp.onClose,
	}
	scp.clientPool[address] = client
	return client, nil
}

func (scp *singleConnectionPerAddress) onClose(address string) {
	scp.mu.Lock()
	defer scp.mu.Unlock()

	delete(scp.clientPool, address)
}

type sharedClient struct {
	mu      sync.Mutex
	client  Client
	address string
	onClose func(address string)
}

func (c *sharedClient) Do(ctx context.Context, req packet.Request) (packet.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil, modbus.ErrClientNotConnected
	}

	return c.client.Do(ctx, req)
}

func (c *sharedClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil
	}

	err := c.client.Close()
	c.client = nil

	c.onClose(c.address)

	return err
}
