package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aldas/go-modbus-client"
	"github.com/aldas/go-modbus-client/poller"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

/*
Example `config.json` content to poll "Victron Energy Meter VM-3P75CT" over UDP

{
  "defaults": {
    "server_address": "udp://192.168.0.200:502?invalid_addr=1000,12000-12100&read_timeout=1s",
    "function_code": 3,
    "unit_id": 1,
    "protocol": "tcp",
    "interval": "1s"
  },
  "fields": [
    {"name": "AcL1Voltage", "address": 12352, "type": "Int16", "scale": 0.01},
    {"name": "AcL1Current", "address": 12353, "type": "Int16", "scale": 0.01},
    {"name": "AcL1EnergyForward", "address": 12354, "type": "Uint32", "scale": 0.01},
    {"name": "AcL1EnergyReverse", "address": 12356, "type": "Uint32", "scale": 0.01},
    {"name": "AcL1ErrorCode", "address": 12358, "type": "Uint16"},
  ]
}
*/

type config struct {
	Defaults modbus.BuilderDefaults `json:"defaults"  mapstructure:"defaults"`
	Fields   []field                `json:"fields"  mapstructure:"fields"`
}

type field struct {
	modbus.Field
	Scale float64 `json:"scale,omitempty" mapstructure:"scale"`
}

// usage: ./modbus-poller -config=config.json
func main() {
	var configLoc string
	flag.StringVar(&configLoc, "config", "config.json", "path to json configuration")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	rawConfig, err := os.ReadFile(configLoc) // #nosec G304
	if err != nil {
		logger.Error("reading config.json failed", "err", err)
		return
	}

	var conf config
	if err := json.Unmarshal(rawConfig, &conf); err != nil {
		logger.Error("config json unmarshalling failed", "err", err)
		return
	}

	scales := map[string]float64{}
	b := modbus.NewRequestBuilderWithConfig(conf.Defaults)
	for _, f := range conf.Fields {
		if f.Scale != 0 {
			scales[f.Name] = f.Scale
		}
		b.AddField(f.Field)
	}
	batches, err := b.Split()
	if err != nil {
		logger.Error("splitting fields to requests failed", "err", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	p := poller.NewPollerWithConfig(batches, poller.Config{Logger: logger})
	go func() {
		for {
			select {
			case result := <-p.ResultChan:
				values := map[string]any{}
				for _, v := range result.Values {
					if v.Error != nil {
						continue
					}
					value := v.Value
					if scale, ok := scales[v.Field.Name]; ok {
						value = scaleValue(scale, value)
					}
					values[v.Field.Name] = value
				}
				if len(values) == 0 {
					continue
				}
				raw, err := json.Marshal(struct {
					Time   time.Time      `json:"time"`
					Values map[string]any `json:"values"`
				}{
					Time:   result.Time,
					Values: values,
				})
				if err != nil {
					logger.Error("failed to marshal result", "err", err)
					continue
				}
				fmt.Printf("%s\n", raw)
			case <-ctx.Done():
				return
			}
		}
	}()

	if err = p.Poll(ctx); err != nil {
		logger.Error("polling ended with failure", "err", err)
		return
	}
	logger.Info("polling ended")
}

func scaleValue(scale float64, value any) any {
	// when scale!=0 value will be converted to float64 type - this is a deliberate feature
	if scale == 0 {
		return value
	}

	switch v := value.(type) {
	case uint8:
		return float64(v) * scale
	case int8:
		return float64(v) * scale
	case uint16:
		return float64(v) * scale
	case int16:
		return float64(v) * scale
	case uint32:
		return float64(v) * scale
	case int32:
		return float64(v) * scale
	case uint64:
		return float64(v) * scale
	case int64:
		return float64(v) * scale
	case float32:
		return float64(v) * scale
	case float64:
		return v * scale
	}
	return value
}
