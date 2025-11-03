package infrastructure

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

const (
	KAFKA_BROKER_URL = "kafka:9092"
	KAFKA_TOPIC      = "full_video"
)

type Publisher struct {
	syncProducer sarama.SyncProducer
}

func (p *Publisher) Close() error {
	return p.syncProducer.Close()
}

func NewProducer() (*Publisher, error) {
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

func (p *Publisher) ReceiveMessage(ctx context.Context, handler func(id string, msg string)) error {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{KAFKA_BROKER_URL}, config)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(KAFKA_TOPIC, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to consume partition: %w", err)
	}
	defer partitionConsumer.Close()

	fmt.Println("Listening to topic:", KAFKA_TOPIC)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("Message received: %s\n", string(msg.Value))

			handler(string(msg.Key), string(msg.Value))
		case err := <-partitionConsumer.Errors():
			fmt.Println("Consumer error:", err)
		case <-ctx.Done():
			fmt.Println("Context canceled, stopping consumer.")
			return nil
		}
	}
}
