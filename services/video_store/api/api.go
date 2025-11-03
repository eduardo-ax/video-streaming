package api

import (
	"io"
	"net/http"

	"github.com/eduardo-ax/video-streaming/services/video_store/domain"
	"github.com/labstack/echo"
)

type VideoRequest struct {
	Title       string `form:"title"`
	Description string `form:"description"`
}

type UploadHandler struct {
	videoUpload domain.VideoUploader
}

func NewVideoHandler(videoUpload domain.VideoUploader) *UploadHandler {
	return &UploadHandler{
		videoUpload: videoUpload,
	}
}

func (v *UploadHandler) Register(e *echo.Group) {
	e.POST("/videos", v.HandleVideoUpload)
	e.GET("/videos/:id/:filename", v.HandleVideoStreaming)
}

func (v *UploadHandler) HandleVideoUpload(c echo.Context) error {
	ctx := c.Request().Context()
	req := &VideoRequest{}

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

	c.Response().Header().Set("Content-Type", contentType)
	_, err = io.Copy(c.Response().Writer, data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to stream video")
	}
	return nil
}
