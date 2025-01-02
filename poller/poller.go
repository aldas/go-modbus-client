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

	statsLock     sync.RWMutex
	statsInterval time.Duration
	batchStats    []BatchStatistics

	batches []modbus.BuilderRequest

	ResultChan chan Result
}

// Config is configuration for Poller
type Config struct {
	// Logger is logger instance used by poller to log. Defaults to slog.Default
	Logger *slog.Logger
	// StatisticsInterval is interval used by poller job to update statistics
	StatisticsInterval time.Duration
	// ConnectFunc is used by poller jobs to open connection to modbus server
	ConnectFunc func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (*modbus.Client, error)
}

// NewPollerWithConfig creates new instance of Poller with given configuration
func NewPollerWithConfig(batches []modbus.BuilderRequest, conf Config) *Poller {
	p := &Poller{
		timeNow:       time.Now,
		logger:        conf.Logger,
		connectFunc:   conf.ConnectFunc,
		statsLock:     sync.RWMutex{},
		statsInterval: 60 * time.Second,
		batchStats:    make([]BatchStatistics, len(batches)),
		ResultChan:    make(chan Result, 2*len(batches)),

		batches: batches,
	}
	if conf.StatisticsInterval > 0 {
		p.statsInterval = conf.StatisticsInterval
	}
	if conf.Logger == nil {
		p.logger = slog.Default()
	}
	if conf.ConnectFunc == nil {
		p.connectFunc = DefaultConnectClient
	}
	for i, batch := range batches {
		p.batchStats[i] = BatchStatistics{
			BatchIndex:    i,
			FunctionCode:  batch.FunctionCode(),
			Protocol:      batch.Protocol,
			ServerAddress: batch.ServerAddress,
		}
	}

	return p
}

// NewPoller creates new instance of Poller with default configuration
func NewPoller(batches []modbus.BuilderRequest) *Poller {
	return NewPollerWithConfig(batches, Config{})
}

// Poll starts polling until context is cancelled
func (p *Poller) Poll(ctx context.Context) error {
	if len(p.batches) == 0 {
		<-ctx.Done()
		return nil
	}

	wg := new(sync.WaitGroup)
	jobStatsChan := make(chan BatchStatistics, 2)
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, statsChan chan BatchStatistics) {
		defer wg.Done()
		p.runJobStatistics(ctx, jobStatsChan)
	}(ctx, wg, jobStatsChan)

	result := p.ResultChan
	for i, batch := range p.batches {
		wg.Add(1)
		j := job{
			timeNow:     p.timeNow,
			logger:      p.logger,
			connectFunc: p.connectFunc,

			statsInterval:  p.statsInterval,
			stats:          p.batchStats[i],
			statisticsChan: jobStatsChan,

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

	statsInterval time.Duration
	batchIndex    int
	batch         modbus.BuilderRequest
	stats         BatchStatistics

	resultsChan    chan Result
	statisticsChan chan BatchStatistics
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
		j.statisticsChan <- j.stats

		err := j.poll(ctx)
		j.stats.IsPolling = false
		j.statisticsChan <- j.stats

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

	statsTicker := time.NewTicker(j.statsInterval)
	defer statsTicker.Stop()
	ticker := time.NewTicker(batch.RequestInterval)
	defer ticker.Stop()

	functionCode := batch.FunctionCode()
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
					"fc", functionCode,
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
					"fc", functionCode,
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
				j.logger.Log(ctx, slog.Level(-8), "request success",
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
			j.statisticsChan <- j.stats

			j.logger.Debug("statistics tick",
				"fc", functionCode,
				"server", batch.ServerAddress,
				"stats", j.stats,
			)
		case <-ctx.Done():
			j.logger.Info("poll done",
				"fc", functionCode,
				"server", batch.ServerAddress,
			)
			return ctx.Err()
		}
	}
}

// DefaultConnectClient is default implementation to create and connect to Modbus server
func DefaultConnectClient(ctx context.Context, protocol modbus.ProtocolType, addressURL string) (*modbus.Client, error) {
	u, err := url.Parse(addressURL)
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

// BatchStatistics holds statistics about specific Poller batch internal state. Batch is identified by BatchIndex.
type BatchStatistics struct {
	BatchIndex int

	FunctionCode  uint8
	Protocol      modbus.ProtocolType
	ServerAddress string

	// IsPolling shows if that batch job currently in polling or waiting for retry
	IsPolling bool
	// StartCount is count how many times the poll job has (re)started
	StartCount uint64
	// RequestOKCount is count how many modbus request have succeeded for that job
	RequestOKCount uint64
	// RequestErrCount is count how many modbus request have failed for that job
	RequestErrCount uint64
	// SendSkipCount is count how many ResultChan sends were skipped due blocked Result channel
	SendSkipCount uint64
}

// BatchStatistics returns statistics of all Poller batches. These are updated when batch job state changes, mostly at Config.StatisticsInterval interval.
func (p *Poller) BatchStatistics() []BatchStatistics {
	p.statsLock.RLock()
	defer p.statsLock.RUnlock()
	return append([]BatchStatistics{}, p.batchStats...)
}

func (p *Poller) runJobStatistics(ctx context.Context, statsChan chan BatchStatistics) {
	for {
		select {
		case jobStatistics := <-statsChan:
			p.statsLock.Lock()
			p.batchStats[jobStatistics.BatchIndex] = jobStatistics
			p.statsLock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
