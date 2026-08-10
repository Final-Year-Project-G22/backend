package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestGetPresignedURL_ToleratesRawFilerResponse(t *testing.T) {
	t.Run("falls back to public URL when filer returns raw bytes", func(t *testing.T) {
		filer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.RawQuery, "ttl=") {
				t.Errorf("expected ttl query param, got %q", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj"))
		}))
		defer filer.Close()

		fs, err := NewSeaweedFS(SeaweedConfig{FilerURL: filer.URL, PublicURL: filer.URL})
		if err != nil {
			t.Fatalf("NewSeaweedFS returned error: %v", err)
		}

		url, err := fs.GetPresignedURL(context.Background(), "docs/sample.pdf", 15*time.Minute)
		if err != nil {
			t.Fatalf("GetPresignedURL returned error: %v", err)
		}
		want := fs.publicURL + "/docs/sample.pdf"
		if url != want {
			t.Errorf("url = %q, want fallback %q", url, want)
		}
	})

	t.Run("returns presigned URL from JSON response when available", func(t *testing.T) {
		filer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"Url":"http://volume.example:8080/docs/sample.pdf?expires=1"}`))
		}))
		defer filer.Close()

		fs, err := NewSeaweedFS(SeaweedConfig{FilerURL: filer.URL, PublicURL: filer.URL})
		if err != nil {
			t.Fatalf("NewSeaweedFS returned error: %v", err)
		}

		url, err := fs.GetPresignedURL(context.Background(), "docs/sample.pdf", 15*time.Minute)
		if err != nil {
			t.Fatalf("GetPresignedURL returned error: %v", err)
		}
		want := "http://volume.example:8080/docs/sample.pdf?expires=1"
		if url != want {
			t.Errorf("url = %q, want %q", url, want)
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
