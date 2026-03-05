// Package storage provides a generic storage interface for file operations.
// It supports multiple storage providers (SeaweedFS, MinIO, etc.) through a unified API.
//
// # Usage
//
// Create a storage instance using the factory:
//
//	config := storage.Config{
//	    Provider: "seaweedfs",
//	    SeaweedFS: storage.SeaweedConfig{
//	        FilerURL:  "localhost:8888",
//	        VolumeURL: "http://localhost:8080",
//	    },
//	}
//
//	s, err := storage.New(config)
//
// Then use the storage interface:
//
//	info, err := s.Upload(ctx, storage.UploadOptions{
//	    Key:         "avatars/user123.jpg",
//	    Content:     fileBytes,
//	    ContentType: "image/jpeg",
//	})
package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage operations.
// Implement this interface to add support for different storage backends.
type Storage interface {
	// Upload uploads a file from byte slice.
	// The key is the unique identifier for the file (e.g., "avatars/user123.jpg").
	Upload(ctx context.Context, opts UploadOptions) (*FileInfo, error)

	// UploadFromReader uploads a file from an io.Reader.
	// Useful for handling multipart form file uploads.
	UploadFromReader(ctx context.Context, opts UploadOptions, reader io.Reader) (*FileInfo, error)

	// Download retrieves a file by its key.
	// Returns the file content as byte slice.
	Download(ctx context.Context, key string) ([]byte, error)

	// Delete removes a file by its key.
	Delete(ctx context.Context, key string) error

	// GetInfo returns metadata about a file.
	// Returns an error if the file doesn't exist.
	GetInfo(ctx context.Context, key string) (*FileInfo, error)

	// GetPresignedURL generates a temporary URL for file access.
	// The URL expires after the specified duration.
	// Returns empty string if the storage provider doesn't support presigned URLs.
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// List returns a list of files matching the specified options.
	// Supports pagination through ListOptions.
	List(ctx context.Context, opts ListOptions) ([]FileInfo, error)

	// Exists checks if a file exists.
	Exists(ctx context.Context, key string) (bool, error)
}

// UploadOptions contains options for file upload.
type UploadOptions struct {
	// Key is the unique identifier for the file.
	// Examples: "avatars/user123.jpg", "documents/invoice-2024.pdf"
	Key string

	// Content is the file content as byte slice.
	// Use this or Reader, not both.
	Content []byte

	// Reader is an optional io.Reader for streaming upload.
	// Use this or Content, not both.
	Reader io.Reader

	// ContentType is the MIME type of the file.
	// Examples: "image/jpeg", "application/pdf"
	ContentType string

	// Metadata is optional custom metadata stored with the file.
	Metadata map[string]string
}

// FileInfo contains metadata about a stored file.
type FileInfo struct {
	// Key is the unique identifier for the file.
	Key string `json:"key"`

	// Name is the original filename if provided.
	Name string `json:"name,omitempty"`

	// Size is the file size in bytes.
	Size int64 `json:"size"`

	// ContentType is the MIME type of the file.
	ContentType string `json:"content_type"`

	// Metadata is custom metadata stored with the file.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is the time when the file was uploaded.
	CreatedAt time.Time `json:"created_at"`

	// URL is the direct access URL (if available).
	URL string `json:"url,omitempty"`

	// ETag is the entity tag for caching.
	ETag string `json:"etag,omitempty"`
}

// ListOptions contains options for listing files.
type ListOptions struct {
	// Prefix filters files by key prefix.
	// Example: "avatars/" lists all files in the avatars folder.
	Prefix string

	// Page is the page number (1-indexed).
	// Default: 1
	Page int

	// PageSize is the number of items per page.
	// Default: 20, Max: 100
	PageSize int
}

// DefaultListOptions returns default options for listing.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Page:     1,
		PageSize: 20,
	}
}

// DefaultUploadOptions returns default options for upload.
func DefaultUploadOptions() UploadOptions {
	return UploadOptions{
		Metadata: make(map[string]string),
	}
}
