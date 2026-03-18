// Package storage provides configuration structures for storage providers.
package storage

// Config contains the configuration for the storage system.
// Use the Provider field to select which storage backend to use.
type Config struct {
	// Provider specifies which storage backend to use.
	// Supported values: "seaweedfs", "minio"
	// Default: "seaweedfs"
	Provider string `mapstructure:"provider"`

	// SeaweedFS contains configuration for SeaweedFS storage.
	SeaweedFS SeaweedConfig `mapstructure:"seaweedfs"`

	// MinIO contains configuration for MinIO storage.
	MinIO MinIOConfig `mapstructure:"minio"`
}

// SeaweedConfig contains configuration for SeaweedFS storage.
//
// SeaweedFS is a distributed file system and S3-compatible object store.
// See: https://github.com/seaweedfs/seaweedfs
type SeaweedConfig struct {
	// FilerURL is the address of the SeaweedFS filer service.
	// Format: "host:port" (e.g., "localhost:8888")
	// Required.
	FilerURL string `mapstructure:"filer_url"`

	// VolumeURL is the base URL for volume servers.
	// Format: "http://host:port" (e.g., "http://localhost:8080")
	// Optional. If empty, derived from FilerURL.
	VolumeURL string `mapstructure:"volume_url"`

	// MasterURL is the address of the SeaweedFS master service.
	// Format: "host:port" (e.g., "localhost:9333")
	// Optional. If empty, derived from FilerURL.
	MasterURL string `mapstructure:"master_url"`

	// Replication specifies the replication strategy.
	// Format: "<dc><rack><data_center>" (e.g., "000" for no replication)
	// Common values:
	//   - "000" - No replication
	//   - "001" - 1 replica
	//   - "010" - Replica across racks
	//   - "100" - Replica across data centers
	// Default: "000"
	Replication string `mapstructure:"replication"`

	// Collection is the name of the collection to store files in.
	// Collections help organize files and can have different storage policies.
	// Optional.
	Collection string `mapstructure:"collection"`

	// Username for authentication (if enabled).
	// Optional.
	Username string `mapstructure:"username"`

	// Password for authentication (if enabled).
	// Optional.
	Password string `mapstructure:"password"`
}

// MinIOConfig contains configuration for MinIO storage.
//
// MinIO is an S3-compatible object storage server.
// See: https://min.io/
type MinIOConfig struct {
	// Endpoint is the MinIO server address.
	// Format: "host:port" (e.g., "localhost:9000")
	// Required.
	Endpoint string `mapstructure:"endpoint"`

	// AccessKey is the MinIO access key (username).
	// Required.
	AccessKey string `mapstructure:"access_key"`

	// SecretKey is the MinIO secret key (password).
	// Required.
	SecretKey string `mapstructure:"secret_key"`

	// Bucket is the default bucket name.
	// Optional. Can be overridden per operation.
	Bucket string `mapstructure:"bucket"`

	// UseSSL enables SSL/TLS connections.
	// Default: false
	UseSSL bool `mapstructure:"use_ssl"`

	// Region specifies the MinIO region.
	// Default: "us-east-1"
	Region string `mapstructure:"region"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		Provider: "seaweedfs",
		SeaweedFS: SeaweedConfig{
			FilerURL:    "localhost:8888",
			VolumeURL:   "http://localhost:8080",
			Replication: "000",
		},
		MinIO: MinIOConfig{
			Endpoint: "localhost:9000",
			Region:   "us-east-1",
			UseSSL:   false,
		},
	}
}
