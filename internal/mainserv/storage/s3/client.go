package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appcfg "github.com/drobyshevv/doc-service/internal/mainserv/config"
)

type Client struct {
	S3     *s3.Client
	Bucket string
}

func NewClient(cfg appcfg.S3Config) (*Client, error) {
	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true

		o.BaseEndpoint = &cfg.Endpoint
	})

	return &Client{
		S3:     s3Client,
		Bucket: cfg.Bucket,
	}, nil
}
