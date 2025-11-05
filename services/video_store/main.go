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
	metrics "github.com/eduardo-ax/video-streaming/services/video_store/observability"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	m.DevicesSet(5)

	echoServer := echo.New()
	echoServer.Use(middleware.CORS())

	echoServer.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	v1Group := echoServer.Group("/v1")
	handler := api.NewVideoHandler(videoUpload)
	handler.Register(v1Group)

	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
