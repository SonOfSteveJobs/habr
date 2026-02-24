package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var appConfig *Config

const defaultEventTTL = 15 * time.Minute

type Config struct {
	dbURI    string
	logger   LoggerConfig
	kafka    KafkaConfig
	eventTTL time.Duration
}

func (c *Config) DBURI() string           { return c.dbURI }
func (c *Config) Logger() LoggerConfig    { return c.logger }
func (c *Config) Kafka() KafkaConfig      { return c.kafka }
func (c *Config) EventTTL() time.Duration { return c.eventTTL }

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

	eventTTL := defaultEventTTL
	if v := os.Getenv("EVENT_TTL"); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parse EVENT_TTL: %w", err)
		}
		eventTTL = parsed
	}

	appConfig = &Config{
		dbURI:    dbURI,
		logger:   logger,
		kafka:    kafka,
		eventTTL: eventTTL,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
