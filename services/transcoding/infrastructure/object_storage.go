package infrastructure

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ObjectStore struct {
	client *s3.Client
	bucket string
}

func NewObjectStore(client *s3.Client, bucket string) *ObjectStore {
	return &ObjectStore{
		client: client,
		bucket: bucket,
	}
}

func (o *ObjectStore) DownloadFile(ctx context.Context, filename string) (string, error) {
	fmt.Println("Downloading file from S3:")
	fmt.Printf("videos/%s\n", filename)

	key := fmt.Sprintf("videos/%s", filename)

	out, err := o.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer out.Body.Close()

	basePath := "/var/videos"
	localPath := filepath.Join(basePath, filename)
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return "", err
	}

	f, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err = io.Copy(f, out.Body); err != nil {
		return "", err
	}

	return localPath, nil
}

func (o *ObjectStore) UploadHLSFiles(ctx context.Context, hlsDir, id string) error {
	bucketPath := fmt.Sprintf("videos/%s/", id)
	files, err := os.ReadDir(hlsDir)
	if err != nil {
		return fmt.Errorf("error reading HLS directory %s: %w", hlsDir, err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			filePath := filepath.Join(hlsDir, fileName)

			s3Key := bucketPath + fileName

			fmt.Printf("Uploading %s to S3 key %s\n", fileName, s3Key)

			if err := o.UploadLocalFile(ctx, filePath, s3Key); err != nil {
				return fmt.Errorf("error uploading file %s to S3: %w", fileName, err)
			}

			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Warning: failed to delete local file %s: %v\n", filePath, err)
			}
		}
	}
	if err := os.Remove(hlsDir); err != nil {
		fmt.Printf("Warning: failed to delete HLS directory %s: %v\n", hlsDir, err)
	}
	return nil
}

func (o *ObjectStore) UploadLocalFile(ctx context.Context, localPath string, key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = o.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(o.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String(contentType),
	})

	return err
}
