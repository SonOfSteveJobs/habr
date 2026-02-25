//go:build integration

package testinfra

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

type KafkaContainer struct {
	Brokers []string
}

func NewKafka(t testing.TB, topics ...string) *KafkaContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	container, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.8.0")
	if err != nil {
		t.Fatalf("testinfra: start kafka container: %v", err)
	}

	brokers, err := container.Brokers(ctx)
	if err != nil {
		t.Fatalf("testinfra: get kafka brokers: %v", err)
	}

	if len(topics) > 0 {
		createTopics(t, brokers, topics)
	}

	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("testinfra: terminate kafka container: %v", err)
		}
	})

	return &KafkaContainer{
		Brokers: brokers,
	}
}

func createTopics(t testing.TB, brokers []string, topics []string) {
	t.Helper()

	cfg := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		t.Fatalf("testinfra: create kafka admin: %v", err)
	}
	defer func() {
		if err := admin.Close(); err != nil {
			t.Logf("testinfra: close kafka admin: %v", err)
		}
	}()

	for _, topic := range topics {
		err := admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     2,
			ReplicationFactor: 1,
		}, false)
		if err != nil {
			t.Fatalf("testinfra: create topic %s: %v", topic, err)
		}
	}
}
