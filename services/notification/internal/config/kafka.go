package config

import (
	"os"
	"strings"
)

type KafkaConfig struct {
	brokers []string
	topic   string
	groupID string
}

func (c KafkaConfig) Brokers() []string { return c.brokers }
func (c KafkaConfig) Topic() string     { return c.topic }
func (c KafkaConfig) GroupID() string   { return c.groupID }

func newKafkaConfig() (KafkaConfig, error) {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		return KafkaConfig{}, ErrKafkaBrokersNotProvided
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		return KafkaConfig{}, ErrKafkaTopicNotProvided
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		return KafkaConfig{}, ErrKafkaGroupIDNotProvided
	}

	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	return KafkaConfig{
		brokers: brokers,
		topic:   topic,
		groupID: groupID,
	}, nil
}
