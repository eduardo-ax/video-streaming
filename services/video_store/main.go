package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/eduardo-ax/video-streaming/services/video_store/api"
	"github.com/eduardo-ax/video-streaming/services/video_store/domain"
	"github.com/eduardo-ax/video-streaming/services/video_store/infrastructure"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("Warning: Could not load .env file, falling back to environment variables.")
	}

	pool := infrastructure.NewPool()
	db := infrastructure.NewDatabase(pool)
	defer db.Close()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	s3Client := s3.NewFromConfig(cfg)
	bucketName := os.Getenv("S3_BUCKET_NAME")
	objectStore := infrastructure.NewObjectStore(s3Client, bucketName)

	pub, err := infrastructure.NewPublisher()
	if err != nil {
		log.Fatalf("FATAL ERROR: Could not initialize Kafka Publisher: %v", err)
	}
	defer pub.Close()

	videoUpload := domain.NewVideoManager(db, pub, objectStore)

	echoServer := echo.New()
	echoServer.Use(middleware.CORS())
	v1Group := echoServer.Group("/v1")
	handler := api.NewVideoHandler(videoUpload)
	handler.Register(v1Group)

	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
