package api

import (
	"context"
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

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open file"})
	}
	defer src.Close()

	fileContent, err := io.ReadAll(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to read file"})
	}
	if err := v.videoUpload.Store(ctx, req.Title, req.Description, fileContent); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to upload video"})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "video uploaded successfully"})
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) error
}
