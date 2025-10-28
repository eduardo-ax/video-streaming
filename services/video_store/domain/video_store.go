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

type VideoManager struct {
	db  Storage
	pub MessagePublisher
}

func NewVideoManager(db Storage, pub MessagePublisher) *VideoManager {
	return &VideoManager{
		db: db,
	}
}

func (v *VideoManager) Store(ctx context.Context, title string, description string, content []byte) error {

	fmt.Printf("Saving video: %s, Description: %s, Size: %d bytes\n", title, description, len(content))
	id, err := v.db.Persist(ctx, title, description)

	if err != nil {
		return err
	}
	fmt.Printf("Video saved with ID: %d\n", id)
	v.pub.SendMessage(fmt.Sprintf("teste", "teste"), content)
	return nil
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) (int, error)
}

type MessagePublisher interface {
	SendMessage(key string, value []byte) error
}
