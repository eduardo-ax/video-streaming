package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eduardo-ax/video-streaming/services/video_store/domain"
	"github.com/eduardo-ax/video-streaming/services/video_store/token"
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
)

func NewVideoHandler(videoUpload domain.VideoUploader, metrics Metrics) *UploadHandler {
	return &UploadHandler{
		videoUpload: videoUpload,
		metrics:     metrics,
	}
}

func JSONError(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{"error": message})
}

func JSONSucess(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{"message": message})
}

const ContextUserID = "userID"
const ContextUserPlan = "userPlan"

func (u *UploadHandler) AuthMiddleware(tokenMaker *token.JWTMaker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return JSONError(c, http.StatusUnauthorized, "missing authorization header")
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
				return JSONError(c, http.StatusUnauthorized, "invalid authorization format")
			}
			accessToken := fields[1]
			claims, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				return JSONError(c, http.StatusUnauthorized, "invalid or expired access token")
			}

			userPlan := fmt.Sprintf("%v", claims.Plan)
			c.Set(ContextUserID, claims.ID)
			c.Set(ContextUserPlan, userPlan)
			return next(c)
		}
	}
}

func (v *UploadHandler) Register(e *echo.Group, tokenMaker *token.JWTMaker) {
	protected := e.Group("")
	protected.Use(v.AuthMiddleware(tokenMaker))

	protected.POST("/videos", v.HandleVideoUpload)
	protected.GET("/videos/:id/:filename", v.HandleVideoStreaming)
}

func (v *UploadHandler) HandleVideoUpload(c echo.Context) error {
	start := time.Now()
	ctx := c.Request().Context()
	req := &VideoRequest{}

	loggedInUserID, ok := c.Get(ContextUserID).(string)

	if !ok || loggedInUserID == "" {
		return JSONError(c, http.StatusUnauthorized, "user ID not available")
	}

	loggedInUserPlan, ok := c.Get(ContextUserPlan).(string)
	fmt.Print("user plan no handle: ")
	fmt.Println(loggedInUserPlan)
	if !ok || loggedInUserPlan == "" {
		return JSONError(c, http.StatusUnauthorized, "user Plan not available")
	}

	v.metrics.DevicesInc()
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	if err := v.videoUpload.Store(ctx, req.Title, req.Description, file, loggedInUserID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upload video")
	}

	duration := time.Since(start).Seconds()
	v.metrics.VideoUploadTime().Observe(duration)
	v.metrics.UploadsInc()
	return echo.NewHTTPError(http.StatusCreated, "video uploaded successfully")
}

func (v *UploadHandler) HandleVideoStreaming(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	filename := c.Param("filename")
	data, contentType, err := v.videoUpload.GetStream(ctx, id, filename)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusNotFound, "file not found")
	}
	defer data.Close()
	c.Response().Header().Set("Content-Type", contentType)
	_, err = io.Copy(c.Response().Writer, data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to stream video")
	}
	return nil
}

type Metrics interface {
	VideoUploadTime() prometheus.Histogram
	DevicesInc()
	UploadsInc()
}
