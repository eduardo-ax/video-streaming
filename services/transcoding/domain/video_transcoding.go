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

func (v *VideoTranscoder) TranscodeVideo(ctx context.Context, content string) error {

	id := 1
	fmt.Println("Transcoding video id:", content)
	_,err := v.ObjectStore.DownloadFile(ctx, id, content)
	if err != nil {
		fmt.Println("Error Download")
		return err
	}
	return nil
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) (int, error)
}

type MessageQueue interface {
	SendMessage(ctx context.Context, key string) error
	ReceiveMessage(ctx context.Context, handler func(msg string)) error
}

type ObjectStore interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, id int) error
	DownloadFile(ctx context.Context, id int, filename string) (string, error)
}
