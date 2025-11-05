package api

import (
	"io"
	"net/http"
	"time"

	"github.com/eduardo-ax/video-streaming/services/video_store/domain"
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
)

type VideoRequest struct {
	Title       string `form:"title"`
	Description string `form:"description"`
}

type UploadHandler struct {
	videoUpload domain.VideoUploader
	metrics     Metrics
}

func NewVideoHandler(videoUpload domain.VideoUploader, metrics Metrics) *UploadHandler {
	return &UploadHandler{
		videoUpload: videoUpload,
		metrics:     metrics,
	}
}

func (v *UploadHandler) Register(e *echo.Group) {
	e.POST("/videos", v.HandleVideoUpload)
	e.GET("/videos/:id/:filename", v.HandleVideoStreaming)
}

func (v *UploadHandler) HandleVideoUpload(c echo.Context) error {
	start := time.Now()
	ctx := c.Request().Context()
	req := &VideoRequest{}

	v.metrics.DevicesInc()
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	if err := v.videoUpload.Store(ctx, req.Title, req.Description, file); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upload video")
	}

	duration := time.Since(start).Seconds() // em segundos
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
