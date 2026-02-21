package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCPort string
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		return nil, errors.New("GRPC_PORT is not provided")
	}

	return &Config{
		GRPCPort: grpcPort,
	}, nil
}
