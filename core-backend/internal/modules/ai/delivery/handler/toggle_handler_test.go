package handler

import (
	"context"
	"testing"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
)

type mockIngestControl struct {
	enabled   bool
	exists    bool
	getErr    error
	setCalled bool
	setVal    bool
	setErr    error
}

func (m *mockIngestControl) IsEnabled(ctx context.Context) bool {
	return m.enabled
}

func (m *mockIngestControl) SetEnabled(ctx context.Context, enabled bool) error {
	m.setCalled = true
	m.setVal = enabled
	return m.setErr
}

func (m *mockIngestControl) GetToggleState(ctx context.Context) (bool, bool, error) {
	return m.enabled, m.exists, m.getErr
}

func newToggleHandler(ctrl *mockIngestControl) *ToggleHandler {
	return &ToggleHandler{ingestControl: ctrl}
}

var _ port.IngestControl = (*mockIngestControl)(nil)

func TestToggleHandler_HandleGetIngestToggle(t *testing.T) {
	ctrl := &mockIngestControl{enabled: true, exists: true}
	h := newToggleHandler(ctrl)

	out, err := h.HandleGetIngestToggle(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Body.Enabled {
		t.Error("expected enabled=true")
	}
}

func TestToggleHandler_HandleGetIngestToggle_Disabled(t *testing.T) {
	ctrl := &mockIngestControl{enabled: false, exists: true}
	h := newToggleHandler(ctrl)

	out, err := h.HandleGetIngestToggle(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.Enabled {
		t.Error("expected enabled=false")
	}
}

func TestToggleHandler_HandleSetIngestToggle(t *testing.T) {
	ctrl := &mockIngestControl{}
	h := newToggleHandler(ctrl)

	input := dto.SetIngestToggleInput{}
	input.Body.Enabled = true

	out, err := h.HandleSetIngestToggle(context.Background(), &input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Body.Enabled {
		t.Error("expected enabled=true")
	}
	if !ctrl.setCalled {
		t.Error("SetEnabled not called")
	}
	if !ctrl.setVal {
		t.Error("SetEnabled got false, want true")
	}
}

func TestToggleHandler_HandleSetIngestToggle_Error(t *testing.T) {
	ctrl := &mockIngestControl{setErr: context.DeadlineExceeded}
	h := newToggleHandler(ctrl)

	input := dto.SetIngestToggleInput{}
	input.Body.Enabled = true

	_, err := h.HandleSetIngestToggle(context.Background(), &input)
	if err == nil {
		t.Error("expected error")
	}
}
