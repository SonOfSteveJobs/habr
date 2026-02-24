package config

import (
	"os"
	"strings"
)

type kafkaConfig struct {
	brokers []string
	topic   string
	groupID string
}

func (c *kafkaConfig) Brokers() []string { return c.brokers }
func (c *kafkaConfig) Topic() string     { return c.topic }
func (c *kafkaConfig) GroupID() string   { return c.groupID }

func newKafkaConfig() (*kafkaConfig, error) {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		return nil, ErrKafkaBrokersNotProvided
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		return nil, ErrKafkaTopicNotProvided
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		return nil, ErrKafkaGroupIDNotProvided
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	return &kafkaConfig{
		brokers: brokers,
		topic:   topic,
		groupID: groupID,
	}, nil
}
