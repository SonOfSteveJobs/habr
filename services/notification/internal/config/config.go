package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var appConfig *Config

const (
	defaultEventTTL        = 15 * time.Minute
	defaultCleanupInterval = 1 * time.Hour
	defaultRetentionPeriod = 7 * 24 * time.Hour
)

type Config struct {
	dbURI           string
	logger          LoggerConfig
	kafka           KafkaConfig
	eventTTL        time.Duration
	cleanupInterval time.Duration
	retentionPeriod time.Duration
}

func (c *Config) DBURI() string                  { return c.dbURI }
func (c *Config) Logger() LoggerConfig           { return c.logger }
func (c *Config) Kafka() KafkaConfig             { return c.kafka }
func (c *Config) EventTTL() time.Duration        { return c.eventTTL }
func (c *Config) CleanupInterval() time.Duration { return c.cleanupInterval }
func (c *Config) RetentionPeriod() time.Duration { return c.retentionPeriod }

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

	cleanupInterval := defaultCleanupInterval
	if v := os.Getenv("CLEANUP_INTERVAL"); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parse CLEANUP_INTERVAL: %w", err)
		}
		cleanupInterval = parsed
	}

	retentionPeriod := defaultRetentionPeriod
	if v := os.Getenv("RETENTION_PERIOD"); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("parse RETENTION_PERIOD: %w", err)
		}
		retentionPeriod = parsed
	}

	appConfig = &Config{
		dbURI:           dbURI,
		logger:          logger,
		kafka:           kafka,
		eventTTL:        eventTTL,
		cleanupInterval: cleanupInterval,
		retentionPeriod: retentionPeriod,
	}

	return nil
}

func AppConfig() *Config { return appConfig }
