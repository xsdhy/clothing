package storage

import (
	"bytes"
	"clothing/internal/config"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type ossStorage struct {
	bucket *oss.Bucket
	prefix string
}

func NewOSSStorage(cfg config.Config) (Storage, error) {
	endpoint := strings.TrimSpace(cfg.StorageOSSEndpoint)
	if endpoint == "" {
		return nil, errors.New("storage: missing OSS endpoint")
	}
	bucketName := strings.TrimSpace(cfg.StorageOSSBucket)
	if bucketName == "" {
		return nil, errors.New("storage: missing OSS bucket")
	}
	accessKey := strings.TrimSpace(cfg.StorageOSSAccessKeyID)
	secretKey := strings.TrimSpace(cfg.StorageOSSAccessKeySecret)
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("storage: missing OSS credentials")
	}

	client, err := oss.New(endpoint, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("storage: create OSS client: %w", err)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("storage: open OSS bucket: %w", err)
	}

	return &ossStorage{
		bucket: bucket,
		prefix: trimPrefix(cfg.StorageOSSPrefix),
	}, nil
}

func (s *ossStorage) Save(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
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
		exists, err := s.bucket.IsObjectExist(key)
		if err != nil {
			return "", fmt.Errorf("check object: %w", err)
		}
		if exists {
			return key, nil
		}
	}

	options := []oss.Option{oss.WithContext(ctx)}
	if ct := detectContentType(opts.Extension); ct != "" {
		options = append(options, oss.ContentType(ct))
	}

	if err := s.bucket.PutObject(key, bytes.NewReader(data), options...); err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return key, nil
}

var _ Storage = (*ossStorage)(nil)
