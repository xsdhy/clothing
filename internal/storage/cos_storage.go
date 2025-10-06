package storage

import (
	"bytes"
	"clothing/internal/config"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type cosStorage struct {
	client *cos.Client
	prefix string
}

func NewCOSStorage(cfg config.Config) (Storage, error) {
	baseURL := strings.TrimSpace(cfg.StorageCOSBucketURL)
	if baseURL == "" {
		return nil, errors.New("storage: missing COS bucket URL")
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("storage: parse COS bucket URL: %w", err)
	}

	secretID := strings.TrimSpace(cfg.StorageCOSSecretID)
	secretKey := strings.TrimSpace(cfg.StorageCOSSecretKey)
	if secretID == "" || secretKey == "" {
		return nil, errors.New("storage: missing COS credentials")
	}

	transport := &cos.AuthorizationTransport{
		SecretID:  secretID,
		SecretKey: secretKey,
	}

	client := cos.NewClient(&cos.BaseURL{BucketURL: parsedURL}, &http.Client{Transport: transport})

	return &cosStorage{
		client: client,
		prefix: trimPrefix(cfg.StorageCOSPrefix),
	}, nil
}

func (s *cosStorage) Save(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
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
		resp, err := s.client.Object.Head(ctx, key, nil)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if err == nil {
			return key, nil
		}
		if !cos.IsNotFoundError(err) {
			return "", fmt.Errorf("head object: %w", err)
		}
	}

	options := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{},
	}
	if ct := detectContentType(opts.Extension); ct != "" {
		options.ObjectPutHeaderOptions.ContentType = ct
	}

	resp, err := s.client.Object.Put(
		ctx,
		key,
		bytes.NewReader(data),
		options,
	)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return key, nil
}

var _ Storage = (*cosStorage)(nil)
