package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// SSEWriter writes Server-Sent Events (SSE) with proper framing and flushing.
type SSEWriter struct {
	ctx     huma.Context
	writer  io.Writer
	flusher http.Flusher
}

// NewSSEWriter creates a new SSE writer bound to the provided Huma context.
func NewSSEWriter(ctx huma.Context) *SSEWriter {
	bodyWriter := ctx.BodyWriter()

	var flusher http.Flusher
	if f, ok := bodyWriter.(http.Flusher); ok {
		flusher = f
	}

	return &SSEWriter{
		ctx:     ctx,
		writer:  bodyWriter,
		flusher: flusher,
	}
}

// WriteHeaders configures standard SSE response headers.
func (w *SSEWriter) WriteHeaders() {
	w.ctx.SetStatus(http.StatusOK)
	w.ctx.SetHeader("Content-Type", "text/event-stream")
	w.ctx.SetHeader("Cache-Control", "no-cache")
	w.ctx.SetHeader("Connection", "keep-alive")
	w.ctx.SetHeader("X-Accel-Buffering", "no")
}

// WriteEvent writes a single SSE event using the provided event type and payload.
func (w *SSEWriter) WriteEvent(eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal sse event payload: %w", err)
	}

	if _, err := fmt.Fprintf(w.writer, "event: %s\n", eventType); err != nil {
		return fmt.Errorf("write sse event type: %w", err)
	}
	if _, err := fmt.Fprintf(w.writer, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("write sse event payload: %w", err)
	}

	w.Flush()
	return nil
}

// WriteComment writes an SSE comment line.
func (w *SSEWriter) WriteComment(comment string) error {
	if _, err := fmt.Fprintf(w.writer, ": %s\n\n", comment); err != nil {
		return fmt.Errorf("write sse comment: %w", err)
	}
	w.Flush()
	return nil
}

// Flush flushes buffered data to the client if flushing is supported.
func (w *SSEWriter) Flush() {
	if w.flusher != nil {
		w.flusher.Flush()
	}
}
