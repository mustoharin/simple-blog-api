package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Uploader handles S3 upload and deletion.
type Uploader struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

// NewUploader creates an S3 Uploader. endpoint may be empty for AWS S3.
func NewUploader(ctx context.Context, region, bucket, endpoint, accessKey, secretKey string) (*Uploader, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}

	var clientOpts []func(*s3.Options)
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(cfg, clientOpts...)

	return &Uploader{client: client, bucket: bucket, endpoint: endpoint}, nil
}

// Upload stores the content at key and returns the public URL.
func (u *Uploader) Upload(ctx context.Context, key, contentType string, data []byte) (string, error) {
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        io.Reader(bytes.NewReader(data)),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}

	var url string
	if u.endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", u.endpoint, u.bucket, key)
	} else {
		url = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.bucket, key)
	}

	return url, nil
}

// Delete removes an object from S3 by key.
func (u *Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete object: %w", err)
	}
	return nil
}
