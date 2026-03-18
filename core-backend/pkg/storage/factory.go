// Package storage provides a factory for creating storage instances.
package storage

import (
	"fmt"
)

// New creates a Storage instance based on the provided configuration.
//
// The provider is determined from config.Provider:
//   - "seaweedfs" - SeaweedFS storage (default)
//   - "minio" - MinIO/S3-compatible storage
//
// Example:
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
func New(config Config) (Storage, error) {
	switch config.Provider {
	case "seaweedfs":
		return NewSeaweedFS(config.SeaweedFS)
	case "minio":
		return NewMinIO(config.MinIO)
	case "":
		return NewSeaweedFS(config.SeaweedFS)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}
}

// NewStorage creates a Storage instance and panics on error.
// Use this for application initialization where failure is fatal.
func NewStorage(config Config) Storage {
	s, err := New(config)
	if err != nil {
		panic(fmt.Errorf("failed to create storage: %w", err))
	}
	return s
}
