package poller

import (
	"context"
	"errors"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

type mockClient struct {
	doCount    int
	onDo       func(doCount int, req packet.Request) (packet.Response, error)
	closeCount int
	onClose    func(closeCount int) error
}

func (c *mockClient) Do(ctx context.Context, req packet.Request) (packet.Response, error) {
	c.doCount++
	if c.onDo != nil {
		return c.onDo(c.doCount, req)
	}
	return nil, errors.New("not implemented")
}

func (c *mockClient) Close() error {
	c.closeCount++
	if c.onClose != nil {
		return c.onClose(c.closeCount)
	}
	return nil
}

func TestNewPollerWithConfig(t *testing.T) {
	f1 := modbus.Field{
		Name:            "f1",
		Address:         10,
		Type:            modbus.FieldTypeInt16,
		ServerAddress:   "server",
		FunctionCode:    packet.FunctionReadHoldingRegisters,
		UnitID:          1,
		Protocol:        modbus.ProtocolTCP,
		RequestInterval: modbus.Duration(50 * time.Millisecond),
	}
	f2 := modbus.Field{
		Name:            "f2",
		Address:         11,
		Type:            modbus.FieldTypeUint32,
		ServerAddress:   "server",
		FunctionCode:    packet.FunctionReadHoldingRegisters,
		UnitID:          1,
		Protocol:        modbus.ProtocolTCP,
		RequestInterval: modbus.Duration(50 * time.Millisecond),
	}
	b := modbus.NewRequestBuilder("x", 1)
	b.AddField(f1)
	b.AddField(f2)
	batches, err := b.Split()
	if !assert.NoError(t, err) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	actualStartAddress := uint16(0)
	actualQuantity := uint16(0)
	actualUnitID := uint8(0)

	client := &mockClient{
		onDo: func(doCount int, req packet.Request) (packet.Response, error) {
			if doCount > 1 {
				cancel() // second request will end the test
				return nil, errors.New("end")
			}
			r, ok := req.(*packet.ReadHoldingRegistersRequestTCP)
			if !ok {
				return nil, errors.New("not fc3")
			}
			actualQuantity = r.Quantity
			actualStartAddress = r.StartAddress
			actualUnitID = r.UnitID

			data := []byte{
				0x0, 0x1, // f1(int16)
				0x0, 0x0, 0x1, 0x1, // f2(int32)
			}
			resp := packet.ReadHoldingRegistersResponseTCP{
				MBAPHeader: packet.MBAPHeader{},
				ReadHoldingRegistersResponse: packet.ReadHoldingRegistersResponse{
					UnitID:          r.UnitID,
					RegisterByteLen: uint8(len(data)),
					Data:            data,
				},
			}
			return resp, nil
		},
	}

	//logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	testTime := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00
	conf := Config{
		Logger: logger,
		ConnectFunc: func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (Client, error) {
			return client, nil
		},
		TimeNow: func() time.Time { return testTime },
	}
	p := NewPollerWithConfig(batches, conf)
	assert.Len(t, p.jobs, 1)

	err = p.Poll(ctx)
	assert.NoError(t, err)

	assert.Equal(t, uint8(1), actualUnitID)
	assert.Equal(t, uint16(10), actualStartAddress)
	assert.Equal(t, uint16(3), actualQuantity)

	result := <-p.ResultChan
	expect := Result{
		BatchIndex: 0,
		Time:       testTime,
		Values: []modbus.FieldValue{
			{Field: f1, Value: int16(1), Error: error(nil)},
			{Field: f2, Value: uint32(0x101), Error: error(nil)},
		},
	}
	assert.Equal(t, expect, result)

}

func TestPoller_PollWithError(t *testing.T) {
	f1 := modbus.Field{
		Name:            "f1",
		Address:         10,
		Type:            modbus.FieldTypeInt16,
		ServerAddress:   "server",
		FunctionCode:    packet.FunctionReadHoldingRegisters,
		UnitID:          1,
		Protocol:        modbus.ProtocolTCP,
		RequestInterval: modbus.Duration(50 * time.Millisecond),
	}
	b := modbus.NewRequestBuilder("x", 1)
	b.AddField(f1)
	batches, err := b.Split()
	if !assert.NoError(t, err) {
		return
	}

	var testCases = []struct {
		name            string
		whenResponse    packet.Response
		whenResponseErr error
		expectStats     []BatchStatistics
	}{
		{
			name: "ok",
			whenResponseErr: &packet.ErrorResponseTCP{
				TransactionID: 1245,
				UnitID:        1,
				Function:      2,
				Code:          packet.ErrIllegalDataAddress,
			},
			expectStats: []BatchStatistics{
				{
					BatchIndex:            0,
					FunctionCode:          0x3,
					Protocol:              0x1,
					ServerAddress:         "server",
					IsPolling:             false,
					StartCount:            0x1,
					RequestOKCount:        0x0,
					RequestErrCount:       0x2,
					RequestModbusErrCount: 0x1,
					SendSkipCount:         0x0,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			client := &mockClient{
				onDo: func(doCount int, req packet.Request) (packet.Response, error) {
					if doCount > 1 {
						cancel() // second request will end the test
						return nil, errors.New("end")
					}
					return tc.whenResponse, tc.whenResponseErr
				},
			}

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			testTime := time.Unix(1615662935, 0).In(time.UTC) // 2021-03-13T19:15:35+00:00
			conf := Config{
				Logger: logger,
				ConnectFunc: func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (Client, error) {
					return client, nil
				},
				OnClientDoErrorFunc: func(err error, batchIndex int) error {
					return err
				},
				TimeNow: func() time.Time { return testTime },
			}
			p := NewPollerWithConfig(batches, conf)
			assert.Len(t, p.jobs, 1)

			err = p.Poll(ctx)
			assert.NoError(t, err)

			actual := p.BatchStatistics()
			assert.Equal(t, tc.expectStats, actual)
		})
	}
}
