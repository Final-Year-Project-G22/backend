package core

import (
	"os"
	"reflect"
	"testing"
)

func TestResolvePlaceholdersInString(t *testing.T) {
	t.Setenv("TEST_HOST", "localhost")
	t.Setenv("TEST_PORT", "5432")

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "plain var",
			in:   "postgres://${TEST_HOST}:${TEST_PORT}",
			want: "postgres://localhost:5432",
		},
		{
			name: "with default when missing",
			in:   "${MISSING_VAR:-fallback}",
			want: "fallback",
		},
		{
			name: "missing without default becomes empty",
			in:   "x-${MISSING_VAR}-y",
			want: "x--y",
		},
		{
			name: "no placeholders",
			in:   "static-value",
			want: "static-value",
		},
		{
			name: "default with env set",
			in:   "${TEST_HOST:-default_host}",
			want: "localhost",
		},
		{
			name: "empty default",
			in:   "${MISSING_VAR:-}",
			want: "",
		},
		{
			name:    "strict mode with missing var",
			in:      "${MISSING_VAR:?this is required}",
			want:    "${MISSING_VAR:?this is required}",
			wantErr: true,
		},
		{
			name:    "strict mode with env set",
			in:      "${TEST_HOST:?this is required}",
			want:    "localhost",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePlaceholdersInString(tt.in, false)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

type testNestedConfig struct {
	Provider string
	Nested   struct {
		URL string
	}
	List []string
	Meta map[string]string
}

func TestResolveConfigPlaceholders(t *testing.T) {
	t.Setenv("PROVIDER", "seaweedfs")
	t.Setenv("FILER_URL", "localhost:8888")

	cfg := &testNestedConfig{
		Provider: "${PROVIDER:-s3}",
		List: []string{
			"${FILER_URL:-default:8888}",
		},
		Meta: map[string]string{
			"url": "${FILER_URL:-fallback}",
		},
	}
	cfg.Nested.URL = "http://${FILER_URL}"

	if err := resolveConfigPlaceholders(cfg, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Provider != "seaweedfs" {
		t.Fatalf("provider got %q", cfg.Provider)
	}
	if cfg.Nested.URL != "http://localhost:8888" {
		t.Fatalf("nested.url got %q", cfg.Nested.URL)
	}
	if !reflect.DeepEqual(cfg.List, []string{"localhost:8888"}) {
		t.Fatalf("list got %#v", cfg.List)
	}
	if cfg.Meta["url"] != "localhost:8888" {
		t.Fatalf("meta[url] got %q", cfg.Meta["url"])
	}

	_ = os.Getenv
}
