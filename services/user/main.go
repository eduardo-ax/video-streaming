package main

import (
	"log"
	"os"

	"github.com/eduardo-ax/video-streaming/services/user/api"
	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/eduardo-ax/video-streaming/services/user/infrastructure"
	"github.com/eduardo-ax/video-streaming/services/user/token"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

const minSecretKeySize = 32

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("Warning: Could not load .env file, falling back to environment variables.")
	}

	secretKey := os.Getenv("SECRET_KEY")

	if len(secretKey) < minSecretKeySize {
		log.Println("Warning: secret key size incorrect")
	}

	pool := infrastructure.NewPool()
	db := infrastructure.NewDatabase(pool)
	defer db.Close()

	token := token.NewJWTMaker(secretKey)

	u := domain.NewUserManager(db, token)

	handler := api.NewUserHander(u)

	echoServer := echo.New()
	v1Group := echoServer.Group("/v1")
	handler.Register(v1Group)
	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
