package infrastructure

import (
	"context"
	"fmt"
	"io"
	"os"

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

func (o *ObjectStore) DownloadVideo(ctx context.Context, id int, filename string) (string, error) {
	key := fmt.Sprintf("videos/%d-%s", id, filename)

	resp, err := o.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	localPath := fmt.Sprintf("/tmp/%d-%s", id, filename)
	outFile, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to copy S3 object to local file: %w", err)
	}

	return localPath, nil
}
