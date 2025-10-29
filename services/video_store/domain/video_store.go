package domain

import (
	"context"
	"fmt"
	"mime/multipart"
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
	Store(ctx context.Context, title string, description string, file *multipart.FileHeader) error
}

type VideoManager struct {
	db          Storage
	pub         MessagePublisher
	ObjectStore ObjectStore
}

func NewVideoManager(db Storage, pub MessagePublisher, objectStore ObjectStore) *VideoManager {
	return &VideoManager{
		db:          db,
		pub:         pub,
		ObjectStore: objectStore,
	}
}

func (v *VideoManager) Store(ctx context.Context, title string, description string, file *multipart.FileHeader) error {
	id, err := v.db.Persist(ctx, title, description)
	if err != nil {
		return err
	}
	fmt.Printf("Video saved with ID: %d\n", id)
	testMessage := fmt.Sprintf("ID do Video: %d", id)
	testContent := []byte(testMessage)
	v.pub.SendMessage(fmt.Sprintf("%d", id), testContent)
	videoURL, err := v.ObjectStore.UploadVideo(ctx, file, id)
	if err != nil {
		return err
	}
	fmt.Printf("Video uploaded to: %s\n", videoURL)
	return nil
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) (int, error)
}

type MessagePublisher interface {
	SendMessage(key string, value []byte) error
}

type ObjectStore interface {
	UploadVideo(ctx context.Context, file *multipart.FileHeader, id int) (string, error)
}
