package infrastructure

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

const (
	KAFKA_BROKER_URL = "localhost:9092"
	KAFKA_TOPIC      = "full_video"
)

type Publisher struct {
	syncProducer sarama.SyncProducer
}

func (p *Publisher) Close() error {
	return p.syncProducer.Close()
}

func NewPublisher() (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Version = sarama.V3_0_0_0
	syncProducer, err := sarama.NewSyncProducer([]string{KAFKA_BROKER_URL}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Sarama producer: %w", err)
	}

	return &Publisher{
		syncProducer: syncProducer,
	}, nil
}

func (p *Publisher) SendMessage(ctx context.Context, key string) error {
	msg := &sarama.ProducerMessage{
		Topic: KAFKA_TOPIC,
		Key:   sarama.StringEncoder(key),
	}
	_, _, err := p.syncProducer.SendMessage(msg)
	return err
}
