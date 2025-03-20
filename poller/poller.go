package poller

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/packet"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	jobHealthTickInterval = 60 * time.Second
)

// Client is interface that Modbus client needs to implement for Poller to be able to request data from server.
type Client interface {
	Do(ctx context.Context, req packet.Request) (packet.Response, error)
	Close() error
}

// Poller is service for sending modbus requests with interval to servers and emitting extracted
// values from request to result channel.
type Poller struct {
	logger      *slog.Logger
	connectFunc func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (Client, error)

	isRunning atomic.Bool
	jobs      []job

	ResultChan chan Result
}

// Config is configuration for Poller
type Config struct {
	// Logger is logger instance used by poller to log.
	// Defaults to slog.Default
	Logger *slog.Logger

	// ConnectFunc is used by poller jobs to open connection to modbus server and request data from it
	// Defaults to DefaultConnectClient
	ConnectFunc func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (Client, error)

	// OnClientDoErrorFunc is called when Client.Do returns with an error.
	// User can decide do suppress certain errors by not returning from this function. In that
	// case these errors will not be included in statistics.
	//
	// Use-case for this callback:
	// Some engine controllers will return packet.ErrIllegalDataValue error code when the engine
	// is not turned on, and you might not want to pollute logs with modbus errors like that.
	OnClientDoErrorFunc func(err error, batchIndex int) error

	// TimeNow allows mocking Result.Time value in tests
	// Defaults to time.Now
	TimeNow func() time.Time
}

// NewPollerWithConfig creates new instance of Poller with given configuration
func NewPollerWithConfig(batches []modbus.BuilderRequest, conf Config) *Poller {
	p := &Poller{
		logger:      conf.Logger,
		connectFunc: conf.ConnectFunc,
		ResultChan:  make(chan Result, 2*len(batches)),

		jobs: make([]job, len(batches)),
	}
	if conf.Logger == nil {
		p.logger = slog.Default()
	}
	if conf.ConnectFunc == nil {
		p.connectFunc = DefaultConnectClient
	}
	timeNow := time.Now
	if conf.TimeNow != nil {
		timeNow = conf.TimeNow
	}
	for i, batch := range batches {
		p.jobs[i] = job{
			timeNow:             timeNow,
			logger:              p.logger,
			connectFunc:         p.connectFunc,
			onClientDoErrorFunc: conf.OnClientDoErrorFunc,

			stats: jobBatchStatistics{
				lock: sync.RWMutex{},
				stats: BatchStatistics{
					BatchIndex:    i,
					FunctionCode:  batch.FunctionCode(),
					Protocol:      batch.Protocol,
					ServerAddress: batch.ServerAddress,
				},
			},
			batchIndex:  i,
			batch:       batch,
			resultsChan: p.ResultChan,
		}
	}

	return p
}

// NewPoller creates new instance of Poller with default configuration
func NewPoller(batches []modbus.BuilderRequest) *Poller {
	return NewPollerWithConfig(batches, Config{})
}

// BatchStatistics returns statistics of all Poller batches.
func (p *Poller) BatchStatistics() []BatchStatistics {
	result := make([]BatchStatistics, len(p.jobs))
	for i := range p.jobs {
		result[i] = p.jobs[i].stats.Stats()
	}
	return result
}

// Poll starts polling until context is cancelled
func (p *Poller) Poll(ctx context.Context) error {
	if isRunning := p.isRunning.Swap(true); isRunning {
		return errors.New("poller is already running")
	}
	defer func() {
		p.isRunning.Store(false)
	}()
	if len(p.jobs) == 0 {
		<-ctx.Done()
		return nil
	}

	wg := new(sync.WaitGroup)
	for i := range p.jobs {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, job *job) {
			defer wg.Done()
			job.Start(ctx)
		}(ctx, wg, &p.jobs[i])
	}
	wg.Wait()
	return nil
}

type job struct {
	timeNow             func() time.Time
	logger              *slog.Logger
	connectFunc         func(ctx context.Context, batchProtocol modbus.ProtocolType, address string) (Client, error)
	onClientDoErrorFunc func(err error, batchIndex int) error

	batchIndex int
	batch      modbus.BuilderRequest
	stats      jobBatchStatistics

	resultsChan chan Result
}

func (j *job) Start(ctx context.Context) {
	const defaultRetry = 1 * time.Second
	retryTime := defaultRetry
	delay := time.NewTimer(retryTime)
	defer delay.Stop()

	for {
		start := j.timeNow()
		j.stats.IncStartCount()
		j.stats.IsPolling(true)
		err := j.poll(ctx)
		j.stats.IsPolling(false)

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

	healthTicker := time.NewTicker(jobHealthTickInterval)
	defer healthTicker.Stop()
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

			if err != nil && j.onClientDoErrorFunc != nil {
				// user can decide do suppress certain errors
				// for example some vessel engine controllers will return packet.ErrIllegalDataValue
				// error code when engine is not turned on, and you might not want to pollute
				// logs with modbus errors like that
				err = j.onClientDoErrorFunc(err, j.batchIndex)
				if err == nil {
					continue
				}
			}

			if err != nil {
				countDoErr++
				j.stats.IncRequestErrCount()

				var mbErr packet.ModbusError
				if errors.As(err, &mbErr) {
					j.stats.IncRequestModbusErrCount()
				}

				j.logger.Error("request failed",
					"err", err,
					"req_duration", reqDuration,
					"fc", functionCode,
					"server", batch.ServerAddress,
					"err_count", countDoErr,
				)

				if errors.Is(err, modbus.ErrClientNotConnected) ||
					errors.Is(err, context.DeadlineExceeded) ||
					errors.Is(err, context.Canceled) {
					return err
				}
				if countDoErr >= maxDoRetryCount {
					return err
				}
				continue
			}
			countDoErr = 0
			j.stats.IncRequestOKCount()

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
					"count_ok", j.stats.stats.RequestOKCount,
					"req_duration", reqDuration,
					"values", values,
				)
			default:
				j.stats.IncSendSkipCount()
				j.logger.Warn("skipped values send to result chan",
					"server", batch.ServerAddress,
				)
			}
		case <-healthTicker.C:
			j.logger.Debug("job health tick",
				"fc", functionCode,
				"server", batch.ServerAddress,
				"stats", j.stats.stats,
			)
		case <-ctx.Done():
			j.logger.Info("job done",
				"fc", functionCode,
				"server", batch.ServerAddress,
			)
			return ctx.Err()
		}
	}
}

func parseAddress(addressURL string) (string, modbus.ClientConfig, error) {
	if !strings.Contains(addressURL, "://") {
		addressURL = "tcp://" + addressURL
	}

	u, err := url.Parse(addressURL)
	if err != nil {
		return "", modbus.ClientConfig{}, fmt.Errorf("failed to parse server address, err: %w", err)
	}
	host := cmp.Or(u.Hostname(), addressURL)
	port := cmp.Or(u.Port(), "502")
	addr := fmt.Sprintf("%s://%s:%s", u.Scheme, host, port)

	writeTimeout, err := durationParam(u.Query(), "write_timeout", 1*time.Second)
	if err != nil {
		return "", modbus.ClientConfig{}, err
	}
	readTimeout, err := durationParam(u.Query(), "read_timeout", 1*time.Second)
	if err != nil {
		return "", modbus.ClientConfig{}, err
	}
	config := modbus.ClientConfig{
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}
	return addr, config, nil
}

// DefaultConnectClient is default implementation to create and connect to Modbus server
func DefaultConnectClient(ctx context.Context, protocol modbus.ProtocolType, addressURL string) (Client, error) {
	addr, config, err := parseAddress(addressURL)
	if err != nil {
		return nil, err
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

	// RequestErrCount is total count how many request have failed for that job
	// this count does not distinguish modbus errors from network errors
	RequestErrCount uint64

	// RequestModbusErrCount is count how many request have failed with modbus error code for that job
	RequestModbusErrCount uint64

	// SendSkipCount is count how many ResultChan sends were skipped due blocked Result channel
	SendSkipCount uint64
}

type jobBatchStatistics struct {
	lock  sync.RWMutex
	stats BatchStatistics
}

func (j *jobBatchStatistics) IsPolling(isPolling bool) {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stats.IsPolling = isPolling
}

func (j *jobBatchStatistics) IncStartCount() {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stats.StartCount++
}

func (j *jobBatchStatistics) IncRequestOKCount() {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stats.RequestOKCount++
}

func (j *jobBatchStatistics) IncRequestErrCount() {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stats.RequestErrCount++
}

func (j *jobBatchStatistics) IncRequestModbusErrCount() {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stats.RequestModbusErrCount++
}

func (j *jobBatchStatistics) IncSendSkipCount() {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stats.SendSkipCount++
}

func (j *jobBatchStatistics) Stats() BatchStatistics {
	j.lock.RLock()
	defer j.lock.RUnlock()
	return j.stats
}
