package main

import (
	"github.com/eduardo-ax/video-streaming/services/video_store/api"
	"github.com/eduardo-ax/video-streaming/services/video_store/domain"
	"github.com/eduardo-ax/video-streaming/services/video_store/infrastructure"
	"github.com/labstack/echo"
)

func main() {
	echoServer := echo.New()
	v1Group := echoServer.Group("/v1")

	pool := infrastructure.NewPool()
	db := infrastructure.NewDatabase(pool)
	defer db.Close()

	pub := infrastructure.NewPublisher()
	defer pub.Close()

	videoUpload := domain.NewVideoManager(db, pub)
	handler := api.NewVideoHandler(videoUpload)
	handler.Register(v1Group)

	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
