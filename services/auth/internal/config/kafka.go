package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultOutboxPollInterval    = 2 * time.Second
	defaultOutboxCleanupInterval = 60 * time.Second
	defaultOutboxFetchLimit      = 100
)

type KafkaConfig struct {
	brokers               []string
	topic                 string
	outboxPollInterval    time.Duration
	outboxCleanupInterval time.Duration
	outboxFetchLimit      int
}

func (c KafkaConfig) Brokers() []string                    { return c.brokers }
func (c KafkaConfig) Topic() string                        { return c.topic }
func (c KafkaConfig) OutboxPollInterval() time.Duration    { return c.outboxPollInterval }
func (c KafkaConfig) OutboxCleanupInterval() time.Duration { return c.outboxCleanupInterval }
func (c KafkaConfig) OutboxFetchLimit() int                { return c.outboxFetchLimit }

func newKafkaConfig() (KafkaConfig, error) {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		return KafkaConfig{}, ErrKafkaBrokersNotProvided
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		return KafkaConfig{}, ErrKafkaTopicNotProvided
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	outboxPollInterval := defaultOutboxPollInterval
	if v := os.Getenv("OUTBOX_POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			outboxPollInterval = d
		}
	}

	outboxCleanupInterval := defaultOutboxCleanupInterval
	if v := os.Getenv("OUTBOX_CLEANUP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			outboxCleanupInterval = d
		}
	}

	outboxFetchLimit := defaultOutboxFetchLimit
	if v := os.Getenv("OUTBOX_FETCH_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			outboxFetchLimit = n
		}
	}

	return KafkaConfig{
		brokers:               brokers,
		topic:                 topic,
		outboxPollInterval:    outboxPollInterval,
		outboxCleanupInterval: outboxCleanupInterval,
		outboxFetchLimit:      outboxFetchLimit,
	}, nil
}
