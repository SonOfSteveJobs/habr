package config

import "time"

type LoggerConfig interface {
	Level() string
	AsJson() bool
}

type KafkaConfig interface {
	Brokers() []string
	Topic() string
	OutboxPollInterval() time.Duration
	OutboxCleanupInterval() time.Duration
	OutboxFetchLimit() int
}
