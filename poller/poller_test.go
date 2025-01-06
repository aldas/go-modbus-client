package poller

import (
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestNewPollerWithConfig(t *testing.T) {
	requestTCP, _ := packet.NewReadHoldingRegistersRequestTCP(1, 1, 1)
	batches := []modbus.BuilderRequest{
		{ServerAddress: "1", Request: requestTCP},
		{ServerAddress: "2", Request: requestTCP},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	conf := Config{
		Logger:             logger,
		StatisticsInterval: 1 * time.Minute,
		ConnectFunc:        nil,
	}
	p := NewPollerWithConfig(batches, conf)

	assert.Equal(t, 1*time.Minute, p.statsInterval)
	assert.Equal(t, logger, p.logger)
	assert.NotNil(t, logger, p.connectFunc)
	assert.Len(t, p.batches, 2)
	assert.Len(t, p.batchStats, 2)
}
