package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"

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

func (o *ObjectStore) UploadVideo(ctx context.Context, file *multipart.FileHeader, id int) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return err
	}

	key := fmt.Sprintf("videos/%s/%s", fmt.Sprintf("%d", id), file.Filename)
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
