//go:build integration

package testinfra_test

import (
	"testing"

	"github.com/IBM/sarama"

	"github.com/SonOfSteveJobs/habr/tests/testinfra"
)

func TestNewKafka(t *testing.T) {
	kc := testinfra.NewKafka(t, "test-topic")

	if len(kc.Brokers) == 0 {
		t.Fatal("no brokers returned")
	}

	cfg := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(kc.Brokers, cfg)
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	defer admin.Close()

	topics, err := admin.ListTopics()
	if err != nil {
		t.Fatalf("list topics: %v", err)
	}

	if _, ok := topics["test-topic"]; !ok {
		t.Fatal("test-topic not found")
	}
}
