package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
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

func (o *ObjectStore) DownloadFile(ctx context.Context, id int, filename string) (string, error) {
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
	localPath := filepath.Join(basePath, fmt.Sprintf("%d", id), filename)
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



func (o *ObjectStore) UploadFile(ctx context.Context, file *multipart.FileHeader, id int) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return err
	}

	key := fmt.Sprintf("videos/%s", fmt.Sprintf("%d-", id)+file.Filename)
	_, err = o.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(o.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return err
	}
	return nil
}
