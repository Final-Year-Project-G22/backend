// Package storage provides MinIO storage implementation.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIO implements Storage interface for MinIO/S3-compatible storage.
type MinIO struct {
	client    *minio.Client
	bucket    string
	presigned bool
}

// NewMinIO creates a new MinIO storage instance.
func NewMinIO(config MinIOConfig) (*MinIO, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}

	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("minio access key and secret key are required")
	}

	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	bucket := config.Bucket
	if bucket == "" {
		bucket = "default"
	}

	m := &MinIO{
		client:    client,
		bucket:    bucket,
		presigned: true,
	}

	if err := m.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}

	return m, nil
}

// CreateUploadIntent generates a direct-upload contract.
// MinIO support is intentionally not implemented in this phase.
func (m *MinIO) CreateUploadIntent(_ context.Context, _ UploadIntentOptions) (*UploadIntent, error) {
	return nil, fmt.Errorf("create upload intent is not implemented for minio")
}

// ensureBucket creates the bucket if it doesn't exist.
func (m *MinIO) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		if err := m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// Upload uploads a file to MinIO.
func (m *MinIO) Upload(ctx context.Context, opts UploadOptions) (*FileInfo, error) {
	if opts.Content == nil {
		return nil, fmt.Errorf("content is required")
	}

	return m.UploadFromReader(ctx, opts, bytes.NewReader(opts.Content))
}

// UploadFromReader uploads a file from an io.Reader.
func (m *MinIO) UploadFromReader(ctx context.Context, opts UploadOptions, reader io.Reader) (*FileInfo, error) {
	if opts.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	contentType := opts.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	metadata := make(map[string]string)
	for k, v := range opts.Metadata {
		metadata[k] = v
	}

	info, err := m.client.PutObject(ctx, m.bucket, opts.Key, reader, -1, minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
		CacheControl: "no-cache",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	fileInfo := &FileInfo{
		Key:         opts.Key,
		Size:        info.Size,
		ContentType: contentType,
		Metadata:    opts.Metadata,
		CreatedAt:   time.Now(),
		ETag:        info.ETag,
	}

	return fileInfo, nil
}

// Download retrieves a file from MinIO.
func (m *MinIO) Download(ctx context.Context, key string) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	if err := obj.Close(); err != nil {
		return nil, fmt.Errorf("failed to close object: %w", err)
	}

	return data, nil
}

// Delete removes a file from MinIO.
func (m *MinIO) Delete(ctx context.Context, key string) error {
	err := m.client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil
		}
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// GetInfo returns metadata about a file.
func (m *MinIO) GetInfo(ctx context.Context, key string) (*FileInfo, error) {
	stat, err := m.client.StatObject(ctx, m.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	metadata := make(map[string]string)
	for k, v := range stat.Metadata {
		if len(v) > 0 {
			metadata[k] = v[0]
		}
	}

	return &FileInfo{
		Key:         key,
		Size:        stat.Size,
		ContentType: stat.ContentType,
		Metadata:    metadata,
		CreatedAt:   stat.LastModified,
		ETag:        stat.ETag,
	}, nil
}

// GetPresignedURL generates a presigned URL for file access.
func (m *MinIO) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if !m.presigned {
		return "", fmt.Errorf("presigned URLs are disabled")
	}

	url, err := m.client.PresignedGetObject(ctx, m.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// List returns a list of files.
func (m *MinIO) List(ctx context.Context, opts ListOptions) ([]FileInfo, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 20
	}
	if opts.PageSize > 1000 {
		opts.PageSize = 1000
	}

	objects := m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    opts.Prefix,
		Recursive: false,
	})

	var files []FileInfo
	offset := (opts.Page - 1) * opts.PageSize
	count := 0
	skipped := 0

	for obj := range objects {
		if obj.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", obj.Err)
		}

		if skipped < offset {
			skipped++
			continue
		}

		if count >= opts.PageSize {
			break
		}

		files = append(files, FileInfo{
			Key:         obj.Key,
			Size:        obj.Size,
			ContentType: obj.ContentType,
			CreatedAt:   obj.LastModified,
			ETag:        obj.ETag,
		})
		count++
	}

	return files, nil
}

// Exists checks if a file exists.
func (m *MinIO) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.GetInfo(ctx, key)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBucket returns the bucket name.
func (m *MinIO) GetBucket() string {
	return m.bucket
}
