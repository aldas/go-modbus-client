package poller

import (
	"context"
	"github.com/aldas/go-modbus-client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSingleConnectionPerAddressClientFunc(t *testing.T) {
	calledCount := 0
	clientFunc := NewSingleConnectionPerAddressClientFunc(func(ctx context.Context, protocol modbus.ProtocolType, addressURL string) (Client, error) {
		c := new(mockClient)
		c.doCount = 999
		calledCount++
		return c, nil
	})

	ctx := context.Background()

	address := "localhost:502?max_quantity_per_request=16"
	client1, err := clientFunc(ctx, modbus.ProtocolTCP, address)
	assert.NoError(t, err)
	client2, err := clientFunc(ctx, modbus.ProtocolTCP, address)
	assert.NoError(t, err)

	assert.Equal(t, &client1, &client2)
	assert.Equal(t, 1, calledCount)
}
