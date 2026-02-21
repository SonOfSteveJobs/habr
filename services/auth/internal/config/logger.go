package config

import (
	"github.com/caarlos0/env/v11"
)

type loggerEnvConfig struct {
	Level  string `env:"LOGGER_LEVEL,required"`
	AsJson bool   `env:"LOGGER_AS_JSON,required"`
}

type loggerConfig struct {
	// приватное поле для изоляции, чтобы из кода случайно не изменить, например, Level
	logger loggerEnvConfig
}

func NewLoggerConfig() (*loggerConfig, error) {
	var raw loggerEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &loggerConfig{logger: raw}, nil
}

func (cfg *loggerConfig) Level() string {
	return cfg.logger.Level
}

func (cfg *loggerConfig) AsJson() bool {
	return cfg.logger.AsJson
}
