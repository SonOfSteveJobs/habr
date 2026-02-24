package producer

import (
	"context"
	"sync"

	"github.com/IBM/sarama"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

// OnSuccessFunc - способ узнать что конкретное сообщение доставлено и выполнить
// действие, например для передачи eventID в outbox
type OnSuccessFunc func(metadata any)

type AsyncProducer struct {
	producer  sarama.AsyncProducer
	topic     string
	onSuccess OnSuccessFunc
	wg        sync.WaitGroup
}

func NewAsync(producer sarama.AsyncProducer, topic string, onSuccess OnSuccessFunc) *AsyncProducer {
	p := &AsyncProducer{
		producer:  producer,
		topic:     topic,
		onSuccess: onSuccess,
	}

	p.wg.Add(2)
	go p.readSuccesses()
	go p.readErrors()

	return p
}

func (p *AsyncProducer) Send(ctx context.Context, msg kafka.Message) error {
	select {
	case p.producer.Input() <- toSaramaMsg(p.topic, msg):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *AsyncProducer) Close() error {
	p.producer.AsyncClose()
	p.wg.Wait()

	return nil
}

func (p *AsyncProducer) readSuccesses() {
	defer p.wg.Done()

	for msg := range p.producer.Successes() {
		if p.onSuccess != nil && msg.Metadata != nil {
			p.onSuccess(msg.Metadata)
		}
	}
}

func (p *AsyncProducer) readErrors() {
	defer p.wg.Done()

	log := logger.Logger()
	for err := range p.producer.Errors() {
		log.Error().
			Err(err.Err).
			Str("topic", p.topic).
			Msg("kafka: async send failed")
	}
}
