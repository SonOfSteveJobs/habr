package config

import (
	"os"
	"strconv"
)

const (
	defaultOtelEnvironment = "local"
	defaultOtelVersion     = "0.1.0"
	defaultSamplingRate    = 1.0
)

type TracingConfig struct {
	endpoint     string
	serviceName  string
	environment  string
	version      string
	samplingRate float64
}

func (c *TracingConfig) Endpoint() string      { return c.endpoint }
func (c *TracingConfig) ServiceName() string   { return c.serviceName }
func (c *TracingConfig) Environment() string   { return c.environment }
func (c *TracingConfig) Version() string       { return c.version }
func (c *TracingConfig) SamplingRate() float64 { return c.samplingRate }

func newTracingConfig() (*TracingConfig, error) {
	endpoint := os.Getenv("OTEL_COLLECTOR_ENDPOINT")
	if endpoint == "" {
		return nil, ErrOtelEndpointNotProvided
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		return nil, ErrOtelServiceNameNotProvided
	}

	environment := defaultOtelEnvironment
	if v := os.Getenv("OTEL_ENVIRONMENT"); v != "" {
		environment = v
	}

	version := defaultOtelVersion
	if v := os.Getenv("OTEL_SERVICE_VERSION"); v != "" {
		version = v
	}

	samplingRate := defaultSamplingRate
	if v := os.Getenv("OTEL_SAMPLING_RATE"); v != "" {
		if r, err := strconv.ParseFloat(v, 64); err == nil {
			samplingRate = r
		}
	}

	return &TracingConfig{
		endpoint:     endpoint,
		serviceName:  serviceName,
		environment:  environment,
		version:      version,
		samplingRate: samplingRate,
	}, nil
}
