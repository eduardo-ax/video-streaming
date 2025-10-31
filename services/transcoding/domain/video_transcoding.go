package domain

import (
	"context"
	"fmt"
	"mime/multipart"
)

type VideoTranscoder struct {
	db          Storage
	queue       MessageQueue
	ObjectStore ObjectStore
}

func NewVideoTranscoder(db Storage, queue MessageQueue, objectStore ObjectStore) *VideoTranscoder {
	return &VideoTranscoder{
		db:          db,
		queue:       queue,
		ObjectStore: objectStore,
	}
}

func TranscodeVideo(videoID string) error {
	fmt.Println("Transcoding video id:", videoID)
	// lógica real de transcoding ficaria aqui
	return nil
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) (int, error)
}

type MessageQueue interface {
	SendMessage(ctx context.Context, key string) error
	ReceiveMessage(ctx context.Context) (string, error)
}

type ObjectStore interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, id int) error
	DownloadFile(ctx context.Context, id int, filename string) (string, error)
}
