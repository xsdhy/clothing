package storage

import (
	"bytes"
	"clothing/internal/config"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type s3ClientOptions struct {
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ForcePathStyle  bool
}

type remoteS3Storage struct {
	client *s3.Client
	bucket string
	prefix string
}

func (s *remoteS3Storage) Save(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty payload")
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	key := buildObjectPath(opts.Category, opts.BaseName, opts.Extension)
	if s.prefix != "" {
		key = joinPrefix(s.prefix, key)
	}

	if opts.SkipIfExists {
		_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
		if err == nil {
			return key, nil
		}
		if !isS3NotFound(err) {
			return "", fmt.Errorf("head object: %w", err)
		}
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	}

	if ct := detectContentType(opts.Extension); ct != "" {
		input.ContentType = aws.String(ct)
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return key, nil
}

var _ Storage = (*remoteS3Storage)(nil)

func isS3NotFound(err error) bool {
	if err == nil {
		return false
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := strings.ToLower(apiErr.ErrorCode())
		if code == "notfound" || code == "nosuchkey" || code == "404" {
			return true
		}
	}
	if strings.Contains(strings.ToLower(err.Error()), "status code: 404") {
		return true
	}
	return false
}

func NewS3Storage(cfg config.Config) (Storage, error) {
	bucket := strings.TrimSpace(cfg.StorageS3Bucket)
	if bucket == "" {
		return nil, errors.New("storage: missing S3 bucket")
	}
	region := strings.TrimSpace(cfg.StorageS3Region)
	if region == "" {
		return nil, errors.New("storage: missing S3 region")
	}
	accessKey := strings.TrimSpace(cfg.StorageS3AccessKeyID)
	secretKey := strings.TrimSpace(cfg.StorageS3SecretAccessKey)
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("storage: missing S3 credentials")
	}

	client, err := newS3Client(s3ClientOptions{
		Region:          region,
		Endpoint:        strings.TrimSpace(cfg.StorageS3Endpoint),
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		SessionToken:    strings.TrimSpace(cfg.StorageS3SessionToken),
		ForcePathStyle:  cfg.StorageS3ForcePathStyle,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: create S3 client: %w", err)
	}

	return &remoteS3Storage{
		client: client,
		bucket: bucket,
		prefix: trimPrefix(cfg.StorageS3Prefix),
	}, nil
}

func newS3Client(opts s3ClientOptions) (*s3.Client, error) {
	region := strings.TrimSpace(opts.Region)
	if region == "" {
		return nil, errors.New("storage: missing S3 region")
	}
	accessKey := strings.TrimSpace(opts.AccessKeyID)
	secretKey := strings.TrimSpace(opts.SecretAccessKey)
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("storage: missing S3 credentials")
	}

	credentialsProvider := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(accessKey, secretKey, strings.TrimSpace(opts.SessionToken)),
	)

	awsCfg := aws.Config{
		Region:      region,
		Credentials: credentialsProvider,
	}

	endpoint := strings.TrimSpace(opts.Endpoint)
	if endpoint != "" {
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "https://" + endpoint
		}
		awsCfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, _ string, _ ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("storage: no endpoint for service %s", service)
		})
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = opts.ForcePathStyle
	})

	return client, nil
}
