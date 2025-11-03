package domain

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type QueueContent struct {
	id      string
	content string
}

func NewQueueContent(id string, content string) (QueueContent, error) {
	integerID, _ := strconv.Atoi(id)
	if integerID < 1 {
		return QueueContent{}, fmt.Errorf("Queue ID error")
	}
	if content < "" {
		return QueueContent{}, fmt.Errorf("Queue Content error")
	}

	return QueueContent{
		id:      id,
		content: content,
	}, nil
}

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

func (v *VideoTranscoder) TranscodeVideo(ctx context.Context, id string, content string) error {

	queueContent, err := NewQueueContent(id, content)

	if err != nil {
		return fmt.Errorf("Queue content error")
	}

	localPath, err := v.ObjectStore.DownloadFile(ctx, queueContent.content)
	if err != nil {
		fmt.Println("Error Download:", err)
		return err
	}

	defer func() {
		if rErr := os.Remove(localPath); rErr != nil {
			fmt.Printf("Warning: failed to delete original file %s: %v\n", localPath, rErr)
		}
	}()

	m3u8Path, err := TranscodeToHLS(ctx, localPath)
	if err != nil {
		fmt.Println("Error Transcode:", err)
		return err
	}

	hlsDir := filepath.Dir(m3u8Path)
	if err := v.ObjectStore.UploadHLSFiles(ctx, hlsDir, queueContent.id); err != nil {
		return err
	}

	fmt.Printf("Transcoding complete: %s\n", m3u8Path)
	return nil
}

func TranscodeToHLS(ctx context.Context, inputPath string) (string, error) {
	dir := filepath.Dir(inputPath)
	outputM3U8 := filepath.Join(dir, "index.m3u8")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-strict", "-2",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-pix_fmt", "yuv420p",
		"-start_number", "0",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(dir, "index%d.ts"),
		"-f", "hls",
		outputM3U8,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %w", err)
	}
	return outputM3U8, nil
}

type Storage interface {
	Persist(ctx context.Context, title string, description string) (int, error)
}

type MessageQueue interface {
	SendMessage(ctx context.Context, key string) error
	ReceiveMessage(ctx context.Context, handler func(id string, msg string)) error
}

type ObjectStore interface {
	DownloadFile(ctx context.Context, filename string) (string, error)
	UploadHLSFiles(ctx context.Context, hlsDir, s3Prefix string) error
}
