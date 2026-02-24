package config

import (
	"os"

	"github.com/joho/godotenv"
)

var appConfig *Config

type Config struct {
	dbURI  string
	logger LoggerConfig
	kafka  KafkaConfig
}

func (c *Config) DBURI() string        { return c.dbURI }
func (c *Config) Logger() LoggerConfig { return c.logger }
func (c *Config) Kafka() KafkaConfig   { return c.kafka }

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	logger, err := NewLoggerConfig()
	if err != nil {
		return err
	}

	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		return ErrDBURINotProvided
	}

	kafka, err := newKafkaConfig()
	if err != nil {
		return err
	}

	appConfig = &Config{
		dbURI:  dbURI,
		logger: logger,
		kafka:  kafka,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
