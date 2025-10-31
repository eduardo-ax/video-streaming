package main

import (
	"context"
	"log"

	"github.com/eduardo-ax/video-streaming/services/transcoding/domain"
	"github.com/eduardo-ax/video-streaming/services/transcoding/infrastructure"
)

func main() {
	producer, err := infrastructure.NewProducer()
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	ctx := context.Background()

	err = producer.ReceiveMessage(ctx, func(msg string) {
		// quando chega mensagem do Kafka, chama o domínio
		if err := domain.TranscodeVideo(msg); err != nil {
			log.Printf("Erro ao transcodificar: %v", err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
