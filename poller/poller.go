package poller

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"github.com/aldas/go-modbus-client"
	"log/slog"
	"net/url"
	"sync"
	"time"
)

// Poller is service for sending modbus requests with interval to servers and emitting extracted
// values from request to result channel.
type Poller struct {
	timeNow     func() time.Time
	logger      *slog.Logger
	connectFunc func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (*modbus.Client, error)

	batches []modbus.BuilderRequest

	ResultChan chan Result
}

// Config is configuration for Poller
type Config struct {
	Logger      *slog.Logger
	ConnectFunc func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (*modbus.Client, error)
}

// NewPollerWithConfig creates new instance of Poller with given configuration
func NewPollerWithConfig(batches []modbus.BuilderRequest, conf Config) *Poller {
	p := &Poller{
		timeNow:     time.Now,
		logger:      conf.Logger,
		connectFunc: conf.ConnectFunc,
		ResultChan:  make(chan Result, len(batches)),

		batches: batches,
	}
	if conf.Logger == nil {
		p.logger = slog.Default()
	}
	if conf.ConnectFunc == nil {
		p.connectFunc = DefaultConnectClient
	}
	return p
}

// NewPoller creates new instance of Poller with default configuration
func NewPoller(batches []modbus.BuilderRequest) *Poller {
	return NewPollerWithConfig(batches, Config{})
}

// Poll starts polling until context is cancelled
func (p *Poller) Poll(ctx context.Context) error {
	wg := new(sync.WaitGroup)
	result := p.ResultChan
	for i, batch := range p.batches {
		wg.Add(1)
		j := job{
			timeNow:     p.timeNow,
			logger:      p.logger,
			connectFunc: p.connectFunc,
			stats: jobStatistics{
				FunctionCode:  batch.FunctionCode(),
				Protocol:      batch.Protocol,
				ServerAddress: batch.ServerAddress,
			},

			batchIndex:  i,
			batch:       batch,
			resultsChan: result,
		}
		go func(ctx context.Context, wg *sync.WaitGroup, job job) {
			defer wg.Done()
			job.Start(ctx)
		}(ctx, wg, j)
	}
	wg.Wait()
	return nil
}

type job struct {
	timeNow     func() time.Time
	logger      *slog.Logger
	connectFunc func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (*modbus.Client, error)

	batchIndex int
	batch      modbus.BuilderRequest
	stats      jobStatistics

	resultsChan chan Result
}

type jobStatistics struct {
	BatchIndex int

	FunctionCode  uint8
	Protocol      modbus.ProtocolType
	ServerAddress string

	IsPolling       bool
	StartCount      uint64
	RequestOKCount  uint64
	RequestErrCount uint64
	SendSkipCount   uint64
}

func (j *job) Start(ctx context.Context) {
	const defaultRetry = 1 * time.Second
	retryTime := defaultRetry
	delay := time.NewTimer(retryTime)
	defer delay.Stop()

	for {
		start := j.timeNow()
		j.stats.IsPolling = true
		j.stats.StartCount++
		err := j.poll(ctx)
		j.stats.IsPolling = false
		if err == nil || ctx.Err() != nil {
			return
		}
		elapsed := j.timeNow().Sub(start)
		if elapsed > 1*time.Minute {
			retryTime = defaultRetry
		} else {
			retryTime = cmp.Or(retryTime*2, 1*time.Minute)
		}
		j.logger.Error("poll failed",
			"error", err,
			"elapsed", elapsed,
			"retry_time", retryTime,
		)

		delay.Reset(retryTime)
		select {
		case <-delay.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

// Result contains extracted values from response with request start time
type Result struct {
	// BatchIndex is index of modbus.BuilderRequest that Poller was created and produced these results
	BatchIndex int
	// Time contains request start time
	Time time.Time
	// Values contains extracted values from response
	Values []modbus.FieldValue
}

func (j *job) poll(ctx context.Context) error {
	batch := j.batch
	client, err := j.connectFunc(ctx, batch.Protocol, batch.ServerAddress)
	if err != nil {
		return err
	}
	defer client.Close()

	statsTicker := time.NewTicker(60 * time.Second)
	defer statsTicker.Stop()
	ticker := time.NewTicker(batch.RequestInterval)
	defer ticker.Stop()

	const maxDoRetryCount = 5
	countDoErr := 0
	for {
		select {
		case <-ticker.C:
			start := j.timeNow()
			resp, err := client.Do(ctx, batch.Request)
			reqDuration := j.timeNow().Sub(start)
			if err != nil {
				countDoErr++
				j.stats.RequestErrCount++

				j.logger.Error("request failed",
					"err", err,
					"req_duration", reqDuration,
					"fc", batch.FunctionCode(),
					"server", batch.ServerAddress,
					"err_count", countDoErr,
				)

				if countDoErr >= maxDoRetryCount {
					return err
				}
				continue
			}
			countDoErr = 0
			j.stats.RequestOKCount++

			values, err := batch.ExtractFields(resp, true)
			if err != nil && !errors.Is(err, modbus.ErrorFieldExtractHadError) {
				j.logger.Error("request extraction failed",
					"err", err,
					"fc", batch.FunctionCode(),
					"server", batch.ServerAddress,
				)
				continue
			}
			result := Result{
				BatchIndex: j.batchIndex,
				Time:       start,
				Values:     values,
			}
			select {
			case j.resultsChan <- result:
				j.logger.Debug("request success",
					"count_ok", j.stats.RequestOKCount,
					"req_duration", reqDuration,
					"values", values,
				)
			default:
				j.stats.SendSkipCount++
				j.logger.Warn("skipped values send to result chan",
					"server", batch.ServerAddress,
				)
			}
		case <-statsTicker.C:
			j.logger.Info("statistics tick",
				"fc", batch.FunctionCode(),
				"server", batch.ServerAddress,
				"stats", j.stats,
			)
		case <-ctx.Done():
			j.logger.Info("poll done",
				"fc", batch.FunctionCode(),
				"server", batch.ServerAddress,
			)
			return ctx.Err()
		}
	}
}

// DefaultConnectClient is default implementation to create and connect to Modbus server
func DefaultConnectClient(ctx context.Context, protocol modbus.ProtocolType, address string) (*modbus.Client, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server address, err: %w", err)
	}
	addr := fmt.Sprintf("%s://%s", cmp.Or(u.Scheme, "tcp"), u.Host)

	writeTimeout, err := durationParam(u.Query(), "write_timeout", 1*time.Second)
	if err != nil {
		return nil, err
	}
	readTimeout, err := durationParam(u.Query(), "read_timeout", 1*time.Second)
	if err != nil {
		return nil, err
	}
	config := modbus.ClientConfig{
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}

	var client *modbus.Client
	switch protocol {
	case modbus.ProtocolTCP:
		client = modbus.NewTCPClientWithConfig(config)
	case modbus.ProtocolRTU:
		client = modbus.NewRTUClientWithConfig(config)
	default:
		return nil, fmt.Errorf("invalid protocol in server address")
	}
	if err := client.Connect(ctx, addr); err != nil {
		return nil, err
	}
	return client, nil
}

func durationParam(queryParams url.Values, param string, defaultValue time.Duration) (time.Duration, error) {
	raw := queryParams.Get(param)
	if raw != "" {
		dur, err := time.ParseDuration(raw)
		if err != nil {
			return 0, fmt.Errorf("failed to parse duration from parameter '%s', err: %w", param, err)
		}
		return dur, nil
	}
	return defaultValue, nil
}
