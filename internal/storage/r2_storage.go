package storage

import (
	"clothing/internal/config"
	"errors"
	"fmt"
	"strings"
)

func NewR2Storage(cfg config.Config) (Storage, error) {
	bucket := strings.TrimSpace(cfg.StorageR2Bucket)
	if bucket == "" {
		return nil, errors.New("storage: missing R2 bucket")
	}
	accessKey := strings.TrimSpace(cfg.StorageR2AccessKeyID)
	secretKey := strings.TrimSpace(cfg.StorageR2SecretAccessKey)
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("storage: missing R2 credentials")
	}

	endpoint := strings.TrimSpace(cfg.StorageR2Endpoint)
	accountID := strings.TrimSpace(cfg.StorageR2AccountID)
	if endpoint == "" {
		if accountID == "" {
			return nil, errors.New("storage: missing R2 endpoint or account id")
		}
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	}

	region := strings.TrimSpace(cfg.StorageR2Region)
	if region == "" {
		region = "auto"
	}

	client, err := newS3Client(s3ClientOptions{
		Region:          region,
		Endpoint:        endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		ForcePathStyle:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: create R2 client: %w", err)
	}

	return &remoteS3Storage{
		client: client,
		bucket: bucket,
		prefix: trimPrefix(cfg.StorageR2Prefix),
	}, nil
}
