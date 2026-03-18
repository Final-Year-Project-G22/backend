// Package storage provides SeaweedFS storage implementation.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SeaweedFS implements Storage interface for SeaweedFS.
type SeaweedFS struct {
	filerURL    string
	volumeURL   string
	masterURL   string
	replication string
	collection  string
	httpClient  *http.Client
}

// SeaweedFSOption configures the SeaweedFS storage.
type SeaweedFSOption func(*SeaweedFS)

// WithSeaweedFSReplication sets the replication strategy.
func WithSeaweedFSReplication(replication string) SeaweedFSOption {
	return func(s *SeaweedFS) {
		s.replication = replication
	}
}

// WithSeaweedFSCollection sets the collection name.
func WithSeaweedFSCollection(collection string) SeaweedFSOption {
	return func(s *SeaweedFS) {
		s.collection = collection
	}
}

// NewSeaweedFS creates a new SeaweedFS storage instance.
func NewSeaweedFS(config SeaweedConfig, opts ...SeaweedFSOption) (*SeaweedFS, error) {
	if config.FilerURL == "" {
		return nil, fmt.Errorf("seaweedfs filer URL is required")
	}

	s := &SeaweedFS{
		filerURL:    config.FilerURL,
		volumeURL:   config.VolumeURL,
		masterURL:   config.MasterURL,
		replication: config.Replication,
		collection:  config.Collection,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.volumeURL == "" {
		s.volumeURL = "http://" + strings.ReplaceAll(s.filerURL, "8888", "8080")
	}

	if s.masterURL == "" {
		s.masterURL = strings.ReplaceAll(s.filerURL, "8888", "9333")
	}

	return s, nil
}

// Upload uploads a file to SeaweedFS.
func (s *SeaweedFS) Upload(ctx context.Context, opts UploadOptions) (*FileInfo, error) {
	if opts.Content == nil {
		return nil, fmt.Errorf("content is required")
	}

	return s.UploadFromReader(ctx, opts, bytes.NewReader(opts.Content))
}

// UploadFromReader uploads a file from an io.Reader.
func (s *SeaweedFS) UploadFromReader(ctx context.Context, opts UploadOptions, reader io.Reader) (*FileInfo, error) {
	if opts.Key == "" {
		opts.Key = generateKey(opts.ContentType)
	}

	assignURL, err := s.getAssignURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get assign URL: %w", err)
	}

	if assignURL.Fid == "" {
		return nil, fmt.Errorf("missing file id from assign response")
	}

	uploadURL := s.volumeURL + "/" + assignURL.Fid

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if opts.ContentType != "" {
		req.Header.Set("Content-Type", opts.ContentType)
	} else {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	for k, v := range opts.Metadata {
		req.Header.Set("X-Meta-"+k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	fileInfo := &FileInfo{
		Key:         opts.Key,
		Size:        0,
		ContentType: opts.ContentType,
		Metadata:    opts.Metadata,
		CreatedAt:   time.Now(),
		URL:         s.volumeURL + "/" + assignURL.Fid,
	}

	if fileInfo.ContentType == "" {
		fileInfo.ContentType = "application/octet-stream"
	}

	return fileInfo, nil
}

// Download retrieves a file from SeaweedFS.
func (s *SeaweedFS) Download(ctx context.Context, key string) ([]byte, error) {
	downloadURL := s.filerURL + "/" + key

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Delete removes a file from SeaweedFS.
func (s *SeaweedFS) Delete(ctx context.Context, key string) error {
	deleteURL := s.filerURL + "/" + key

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetInfo returns metadata about a file.
func (s *SeaweedFS) GetInfo(ctx context.Context, key string) (*FileInfo, error) {
	statURL := s.filerURL + "/stat" + key

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stat failed with status %d", resp.StatusCode)
	}

	var statResp SeaweedStatResponse
	if err := json.NewDecoder(resp.Body).Decode(&statResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	createdAt := time.Now()
	if statResp.Mtime > 0 {
		createdAt = time.Unix(statResp.Mtime, 0)
	}

	metadata := make(map[string]string)
	for k, v := range statResp.Meta {
		metadata[k] = v
	}

	return &FileInfo{
		Key:         key,
		Size:        int64(statResp.Size),
		ContentType: statResp.Mime,
		Metadata:    metadata,
		CreatedAt:   createdAt,
		URL:         s.volumeURL + "/" + key,
	}, nil
}

// GetPresignedURL generates a presigned URL for file access.
func (s *SeaweedFS) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignURL := fmt.Sprintf("%s/%s?ttl=%s", s.filerURL, key, formatExpiry(expiry))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, presignURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("presign failed with status %d", resp.StatusCode)
	}

	var presignResp SeaweedPresignResponse
	if err := json.NewDecoder(resp.Body).Decode(&presignResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return presignResp.Url, nil
}

// List returns a list of files.
func (s *SeaweedFS) List(ctx context.Context, opts ListOptions) ([]FileInfo, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 20
	}

	listURL := s.filerURL + "/list"
	if opts.Prefix != "" {
		listURL = s.filerURL + "/dir/" + opts.Prefix
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("pageLimit", fmt.Sprintf("%d", opts.PageSize))
	q.Add("pageOffset", fmt.Sprintf("%d", (opts.Page-1)*opts.PageSize))
	req.URL.RawQuery = q.Encode()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list failed with status %d", resp.StatusCode)
	}

	var listResp SeaweedListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var files []FileInfo
	for _, entry := range listResp.Entries {
		files = append(files, FileInfo{
			Key:         path.Join(opts.Prefix, entry.Name),
			Name:        entry.Name,
			Size:        entry.Size,
			ContentType: entry.Mime,
			CreatedAt:   time.Unix(entry.Mtime, 0),
		})
	}

	return files, nil
}

// Exists checks if a file exists.
func (s *SeaweedFS) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.GetInfo(ctx, key)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getAssignURL requests a file ID and upload URL from the master.
func (s *SeaweedFS) getAssignURL(ctx context.Context) (*SeaweedAssignResponse, error) {
	assignURL := "http://" + s.masterURL + "/dir/assign"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assignURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if s.replication != "" {
		q.Add("replication", s.replication)
	}
	if s.collection != "" {
		q.Add("collection", s.collection)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("assign failed with status %d", resp.StatusCode)
	}

	var assignResp SeaweedAssignResponse
	if err := json.NewDecoder(resp.Body).Decode(&assignResp); err != nil {
		return nil, err
	}

	return &assignResp, nil
}

// generateKey generates a unique key for a file.
func generateKey(contentType string) string {
	ext := getExtension(contentType)
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}

// getExtension returns the file extension for a content type.
func getExtension(contentType string) string {
	extMap := map[string]string{
		"image/jpeg":               ".jpg",
		"image/png":                ".png",
		"image/gif":                ".gif",
		"image/webp":               ".webp",
		"image/svg+xml":            ".svg",
		"application/pdf":          ".pdf",
		"application/json":         ".json",
		"application/xml":          ".xml",
		"text/plain":               ".txt",
		"text/html":                ".html",
		"text/css":                 ".css",
		"application/javascript":   ".js",
		"application/zip":          ".zip",
		"application/octet-stream": "",
	}

	if ext, ok := extMap[contentType]; ok {
		return ext
	}
	return ""
}

// formatExpiry formats duration as SeaweedFS TTL.
func formatExpiry(d time.Duration) string {
	if d <= 0 {
		return "1h"
	}
	hours := int(d.Hours())
	if hours < 1 {
		return "30m"
	}
	return fmt.Sprintf("%dh", hours)
}

func closeResponseBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	if err := body.Close(); err != nil {
		fmt.Printf("failed to close response body: %v\n", err)
	}
}

// Response types for SeaweedFS API.

type SeaweedAssignResponse struct {
	Fid       string `json:"fid"`
	URL       string `json:"url"`
	Count     int    `json:"count"`
	Auth      string `json:"auth,omitempty"`
	Token     string `json:"token,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

type SeaweedStatResponse struct {
	Name     string            `json:"Name"`
	Mime     string            `json:"Mime"`
	Size     int               `json:"Size"`
	Mtime    int64             `json:"mtime"`
	Meta     map[string]string `json:"Meta"`
	ETag     string            `json:"etag,omitempty"`
	FileSize int64             `json:"fileSize,omitempty"`
}

type SeaweedListResponse struct {
	Entries []SeaweedEntry `json:"entries"`
}

type SeaweedEntry struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mime  string `json:"mime"`
	Mtime int64  `json:"mtime"`
}

type SeaweedPresignResponse struct {
	Url string `json:"Url"`
}
