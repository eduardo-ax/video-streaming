package domain

import (
	"context"
	"fmt"
)

type Video struct {
	Content     []byte
	Title       string
	Description string
}

func NewVideo(title string, description string, content []byte) *Video {
	return &Video{
		Content:     content,
		Title:       title,
		Description: description,
	}
}

type VideoUploader interface {
	Store(ctx context.Context, title string, description string, content []byte) error
}

type VideoManager struct{}

func (v *VideoManager) Store(ctx context.Context, title string, description string, content []byte) error {

	fmt.Printf("Saving video: %s, Description: %s, Size: %d bytes\n", title, description, len(content))
	return nil
}
