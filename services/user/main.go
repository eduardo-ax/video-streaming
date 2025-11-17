package main

import (
	"log"
	"os"

	"github.com/eduardo-ax/video-streaming/services/user/api"
	_ "github.com/eduardo-ax/video-streaming/services/user/docs"
	"github.com/eduardo-ax/video-streaming/services/user/domain"
	"github.com/eduardo-ax/video-streaming/services/user/infrastructure"
	metrics "github.com/eduardo-ax/video-streaming/services/user/observability"
	"github.com/eduardo-ax/video-streaming/services/user/token"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const minSecretKeySize = 32

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @BasePath /v1
// @host localhost:8080
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

	reg := prometheus.NewRegistry()
	m := metrics.NewUserMetrics(reg)

	handler := api.NewUserHander(u, m)
	echoServer := echo.New()
	//echoServer.Use(middleware.CORS())

	echoServer.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	echoServer.GET("/swagger/*", echoSwagger.WrapHandler)
	v1Group := echoServer.Group("/v1")
	handler.Register(v1Group, token)
	echoServer.Logger.Fatal(echoServer.Start(":8080"))

}
