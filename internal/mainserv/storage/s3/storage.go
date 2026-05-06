package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Storage struct {
	client *Client
}

func NewStorage(client *Client) *Storage {
	return &Storage{client: client}
}

func (s *Storage) UploadFile(
	ctx context.Context,
	key string,
	data []byte,
	contentType string,
) error {

	_, err := s.client.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.client.Bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})

	if err != nil {
		return fmt.Errorf("upload file to s3: %w", err)
	}

	return nil
}

func (s *Storage) DownloadFile(
	ctx context.Context,
	key string,
) ([]byte, error) {

	out, err := s.client.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.client.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("get file from s3: %w", err)
	}
	defer out.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(out.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *Storage) DeleteFile(
	ctx context.Context,
	key string,
) error {

	_, err := s.client.S3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.client.Bucket,
		Key:    &key,
	})

	return err
}

func GenerateKey(filename string) string {
	return uuid.New().String() + "-" + filename
}
