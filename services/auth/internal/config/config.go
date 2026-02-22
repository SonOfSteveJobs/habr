package config

import (
	"os"

	"github.com/joho/godotenv"
)

var appConfig *Config

type Config struct {
	grpcPort string
	logger   LoggerConfig
}

func (c *Config) GRPCPort() string { return c.grpcPort }

func (c *Config) Logger() LoggerConfig {
	return c.logger
}

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	logger, err := NewLoggerConfig()
	if err != nil {
		return err
	}

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		return ErrGRPCPortNotProvided
	}

	appConfig = &Config{
		grpcPort: grpcPort,
		logger:   logger,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
