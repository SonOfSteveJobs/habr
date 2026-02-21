package config

import (
	"os"

	"github.com/joho/godotenv"
)

var appConfig *config

type config struct {
	grpcPort string
	logger   LoggerConfig
}

func (c *config) GRPCPort() string {
	return c.grpcPort
}

func (c *config) Logger() LoggerConfig {
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

	appConfig = &config{
		grpcPort: grpcPort,
		logger:   logger,
	}

	return nil
}

func AppConfig() *config {
	return appConfig
}
