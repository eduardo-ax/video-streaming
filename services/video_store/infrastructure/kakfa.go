package infrastructure

import "github.com/IBM/sarama"

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

// NewProducer inicializa o produtor do Sarama e retorna a interface Producer.
func NewPublisher() *Publisher {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Version = sarama.V3_0_0_0 // Ajuste a versão conforme necessário

	syncProducer, err := sarama.NewSyncProducer([]string{KAFKA_BROKER_URL}, config)
	if err != nil {
		return nil
	}

	return &Publisher{
		syncProducer: syncProducer,
	}
}

func (p *Publisher) SendMessage(key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: KAFKA_TOPIC,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := p.syncProducer.SendMessage(msg)
	return err
}
