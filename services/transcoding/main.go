package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/eduardo-ax/video-streaming/services/transcoding/domain"
	"github.com/eduardo-ax/video-streaming/services/transcoding/infrastructure"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	pool := infrastructure.NewPool()
	defer pool.Close()
	db := infrastructure.NewDatabase(pool)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	s3Client := s3.NewFromConfig(cfg)
	bucketName := os.Getenv("S3_BUCKET_NAME")
	objectStore := infrastructure.NewObjectStore(s3Client, bucketName)

	producer, err := infrastructure.NewProducer()
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	ctx := context.Background()
	videoTranscoder := domain.NewVideoTranscoder(db, producer, objectStore)

	err = producer.ReceiveMessage(ctx, func(msg string) {
		if err := videoTranscoder.TranscodeVideo(ctx, msg); err != nil {
			log.Printf("transcoding error %v", err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
