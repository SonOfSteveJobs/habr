package config

import (
	"os"
	"strconv"
)

type loggerConfig struct {
	level  string
	asJson bool
}

func NewLoggerConfig() (*loggerConfig, error) {
	level := os.Getenv("LOGGER_LEVEL")
	if level == "" {
		return nil, ErrLoggerLevelNotProvided
	}

	asJsonStr := os.Getenv("LOGGER_AS_JSON")
	if asJsonStr == "" {
		return nil, ErrLoggerAsJsonNotProvided
	}

	asJson, err := strconv.ParseBool(asJsonStr)
	if err != nil {
		return nil, ErrLoggerAsJsonInvalid
	}

	return &loggerConfig{level: level, asJson: asJson}, nil
}

func (cfg *loggerConfig) Level() string {
	return cfg.level
}

func (cfg *loggerConfig) AsJson() bool {
	return cfg.asJson
}
