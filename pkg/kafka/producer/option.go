package producer

import (
	"time"

	"github.com/IBM/sarama"
)

// Option - опция для кофига сарамы
type Option func(*sarama.Config)

// WithIdempotent - нужен ли ключ идемпотентности
func WithIdempotent() Option {
	return func(c *sarama.Config) {
		c.Producer.Idempotent = true
		c.Producer.RequiredAcks = sarama.WaitForAll
		c.Net.MaxOpenRequests = 1
	}
}

// WithFlush - нужен ли батч на отправку (например каждые 100 сообщений или каждые 0.5 сек)
func WithFlush(messages int, frequency time.Duration) Option {
	return func(c *sarama.Config) {
		c.Producer.Flush.Messages = messages
		c.Producer.Flush.Frequency = frequency
	}
}

// WithRetry - нужно ли ретраить сообщения которые упали с ошибкой
func WithRetry(max int) Option {
	return func(c *sarama.Config) {
		c.Producer.Retry.Max = max
	}
}

func NewSyncConfig(opts ...Option) *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// NewAsyncConfig - создаем конфиг с набором опций, например:
//
// cfg := producer.NewAsyncConfig(
// producer.WithIdempotent(),
// producer.WithFlush(100, 500*time.Millisecond),
// producer.WithRetry(3),
// )
func NewAsyncConfig(opts ...Option) *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	for _, opt := range opts {
		opt(config)
	}

	return config
}
