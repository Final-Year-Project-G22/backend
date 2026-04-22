package handler

import (
	"testing"
	"time"
)

func TestParseTimeFromRFC3339(t *testing.T) {
	t.Run("valid RFC3339 timestamp", func(t *testing.T) {
		got := parseTimeFromRFC3339("2026-01-01T10:00:00Z")
		if got.IsZero() {
			t.Fatal("expected parsed timestamp, got zero time")
		}
		if got.UTC().Format(time.RFC3339) != "2026-01-01T10:00:00Z" {
			t.Fatalf("unexpected parsed time: %s", got.UTC().Format(time.RFC3339))
		}
	})

	t.Run("invalid timestamp returns zero time", func(t *testing.T) {
		got := parseTimeFromRFC3339("not-a-time")
		if !got.IsZero() {
			t.Fatal("expected zero time for invalid timestamp")
		}
	})
}
