package config

import (
	"os"
	"strings"
)

type kafkaConfig struct {
	brokers []string
	topic   string
}

func (c kafkaConfig) Brokers() []string { return c.brokers }
func (c kafkaConfig) Topic() string     { return c.topic }

func newKafkaConfig() (kafkaConfig, error) {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		return kafkaConfig{}, ErrKafkaBrokersNotProvided
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		return kafkaConfig{}, ErrKafkaTopicNotProvided
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	return kafkaConfig{
		brokers: brokers,
		topic:   topic,
	}, nil
}
