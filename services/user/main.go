package main

import (
	"github.com/eduardo-ax/video-streaming/services/user/api"
	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/eduardo-ax/video-streaming/services/user/infrastructure"
	"github.com/labstack/echo/v4"
)

func main() {

	pool := infrastructure.NewPool()
	db := infrastructure.NewDatabase(pool)
	defer db.Close()

	u := domain.NewUserManager(db)

	handler := api.NewUserHander(u)

	echoServer := echo.New()
	v1Group := echoServer.Group("/v1")
	handler.Register(v1Group)
	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
