package main

import (
	"github.com/eduardo-ax/video-streaming/services/video_store/api"
	"github.com/eduardo-ax/video-streaming/services/video_store/domain"
	"github.com/labstack/echo"
)

func main() {
	echoServer := echo.New()
	v1Group := echoServer.Group("/v1")

	videoUpload := &domain.VideoManager{}
	handler := api.NewVideoHandler(videoUpload)
	handler.Register(v1Group)

	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
