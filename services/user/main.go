package main

import (
	"github.com/eduardo-ax/video-streaming/services/user/api"
	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/labstack/echo/v4"
)

func main() {

	user := domain.NewUserManager()
	handler := api.NewUserHander(user)

	echoServer := echo.New()
	v1Group := echoServer.Group("/v1")
	handler.Register(v1Group)
	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
