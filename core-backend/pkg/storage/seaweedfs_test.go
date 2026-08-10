package storage

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewSeaweedFS_NormalizesBaseURLs(t *testing.T) {
	t.Run("filer URL without scheme gains http prefix and public URL falls back to normalized filer URL", func(t *testing.T) {
		fs, err := NewSeaweedFS(SeaweedConfig{FilerURL: "localhost:8888"})
		if err != nil {
			t.Fatalf("NewSeaweedFS returned error: %v", err)
		}
		if fs.filerURL != "http://localhost:8888" {
			t.Errorf("filerURL = %q, want %q", fs.filerURL, "http://localhost:8888")
		}
		if fs.publicURL != "http://localhost:8888" {
			t.Errorf("publicURL = %q, want %q", fs.publicURL, "http://localhost:8888")
		}
	})

	t.Run("explicit https public URL is preserved unchanged", func(t *testing.T) {
		fs, err := NewSeaweedFS(SeaweedConfig{
			FilerURL:  "localhost:8888",
			PublicURL: "https://files.johna.me",
		})
		if err != nil {
			t.Fatalf("NewSeaweedFS returned error: %v", err)
		}
		if fs.publicURL != "https://files.johna.me" {
			t.Errorf("publicURL = %q, want %q", fs.publicURL, "https://files.johna.me")
		}
	})

	t.Run("bare public URL gains http prefix", func(t *testing.T) {
		fs, err := NewSeaweedFS(SeaweedConfig{
			FilerURL:  "http://localhost:8888",
			PublicURL: "localhost:8888",
		})
		if err != nil {
			t.Fatalf("NewSeaweedFS returned error: %v", err)
		}
		if fs.publicURL != "http://localhost:8888" {
			t.Errorf("publicURL = %q, want %q", fs.publicURL, "http://localhost:8888")
		}
	})
}

func TestCreateUploadIntent_ReturnsAbsoluteUploadURL(t *testing.T) {
	fs, err := NewSeaweedFS(SeaweedConfig{FilerURL: "localhost:8888"})
	if err != nil {
		t.Fatalf("NewSeaweedFS returned error: %v", err)
	}

	intent, err := fs.CreateUploadIntent(context.Background(), UploadIntentOptions{
		Key:         "avatars/user123.jpg",
		ContentType: "image/jpeg",
		Expiry:      5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("CreateUploadIntent returned error: %v", err)
	}

	if !strings.HasPrefix(intent.UploadURL, "http://") && !strings.HasPrefix(intent.UploadURL, "https://") {
		t.Errorf("UploadURL = %q, want scheme http:// or https://", intent.UploadURL)
	}
}
