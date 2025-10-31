package domain

import (
	"context"
	"fmt"
	"mime/multipart"
)

type Video struct {
	Content     *multipart.FileHeader
	Title       string
	Description string
}

func NewVideo(title string, description string, content *multipart.FileHeader) (Video, error) {

	if content == nil {
		return Video{}, fmt.Errorf("video content cannot be nil")
	}

	if title == "" || description == "" || len(title) > 100 || len(description) > 500 {
		return Video{}, fmt.Errorf("invalid video data")
	}

	return Video{
		Content:     content,
		Title:       title,
		Description: description,
	}, nil
}

type VideoUploader interface {
	Store(ctx context.Context, title string, description string, file *multipart.FileHeader) error
}

type VideoManager struct {
	db          Storage
	pub         MessagePublisher
	objectStore ObjectStore
}

func NewVideoManager(db Storage, pub MessagePublisher, objectStore ObjectStore) *VideoManager {
	return &VideoManager{
		db:          db,
		pub:         pub,
		objectStore: objectStore,
	}
}

func (v *VideoManager) Store(ctx context.Context, title string, description string, file *multipart.FileHeader) error {
	src, err := NewVideo(title, description, file)
	if err != nil {
		fmt.Printf("Error creating video entity: %v", err)
		return err
	}

	id, err := v.db.Persist(ctx, src.Title, src.Description)
	if err != nil {
		fmt.Printf("Error persisting video metadata: %v", err)
		return err
	}

	err = v.objectStore.UploadVideo(ctx, file, id)
	if err != nil {
		fmt.Printf("Error uploading video to object store: %v", err)	
		return err
	}

	fmt.Printf("Video saved with ID: %d\n", id)
	err = v.pub.SendMessage(ctx, fmt.Sprintf("%d", id), file.Filename)
	if err != nil {
		return err
	}
	return nil
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) (int, error)
}

type MessagePublisher interface {
	SendMessage(ctx context.Context, id string, filename string) error
}

type ObjectStore interface {
	UploadVideo(ctx context.Context, file *multipart.FileHeader, id int) error
}
