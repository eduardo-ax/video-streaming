package api

import "github.com/eduardo-ax/video-streaming/services/video_store/domain"

type VideoRequest struct {
	Title       string `form:"title"`
	Description string `form:"description"`
}

type UploadHandler struct {
	videoUpload domain.VideoUploader
	metrics     Metrics
}


