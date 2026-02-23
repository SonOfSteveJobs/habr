package producer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

type SyncProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewSync(producer sarama.SyncProducer, topic string) *SyncProducer {
	return &SyncProducer{
		producer: producer,
		topic:    topic,
	}
}

func (p *SyncProducer) Send(ctx context.Context, msg kafka.Message) error {
	_, _, err := p.producer.SendMessage(toSaramaMsg(p.topic, msg))
	if err != nil {
		log := logger.Ctx(ctx)
		log.Error().
			Err(err).
			Str("topic", p.topic).
			Msg("kafka: send failed")

		return fmt.Errorf("kafka sync send: %w", err)
	}

	return nil
}

func (p *SyncProducer) Close() error {
	return p.producer.Close()
}

func toSaramaMsg(topic string, msg kafka.Message) *sarama.ProducerMessage {
	pm := &sarama.ProducerMessage{
		Topic:    topic,
		Value:    sarama.ByteEncoder(msg.Value),
		Metadata: msg.Metadata,
	}

	if msg.Key != nil {
		pm.Key = sarama.ByteEncoder(msg.Key)
	}

	for k, v := range msg.Headers {
		pm.Headers = append(pm.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: v,
		})
	}

	return pm
}
