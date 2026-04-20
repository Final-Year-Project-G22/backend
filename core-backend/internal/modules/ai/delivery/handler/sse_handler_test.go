package handler

import (
	"context"
	"testing"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/google/uuid"
)

type mockSSEGateway struct {
	called  bool
	account uuid.UUID
	lastID  string
	sendErr error
	callSeq int
}

func (m *mockSSEGateway) StreamStatusByAccount(ctx context.Context, accountID uuid.UUID, lastEventID string, sendFunc service.SSEDeliveryFunc) error {
	m.called = true
	m.account = accountID
	m.lastID = lastEventID
	m.callSeq++
	if m.sendErr != nil {
		return m.sendErr
	}
	_ = sendFunc("0", []byte(`{"status":"completed"}`))
	return nil
}

var _ sseStreamer = (*mockSSEGateway)(nil)

func TestSSEHandler_StreamAccountStatus(t *testing.T) {
	accountID := uuid.New()
	ctx := context.WithValue(context.Background(), contextkeys.AccountID, accountID)

	mock := &mockSSEGateway{}
	h := &SSEHandler{gateway: mock}
	_ = h.StreamAccountStatus(ctx, "100", func(event string, payload any) error {
		return nil
	})

	if !mock.called {
		t.Error("expected gateway.StreamStatusByAccount to be called")
	}
	if mock.account != accountID {
		t.Errorf("account = %v, want %v", mock.account, accountID)
	}
	if mock.lastID != "100" {
		t.Errorf("lastEventID = %v, want %v", mock.lastID, "100")
	}
}

func TestSSEHandler_StreamAccountStatus_NoAuth(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.AccountID, contextkeys.NilUUID)

	mock := &mockSSEGateway{}
	h := &SSEHandler{gateway: mock}

	err := h.StreamAccountStatus(ctx, "", func(event string, payload any) error {
		return nil
	})
	if err == nil {
		t.Error("expected unauthorized error")
	}
}

func TestSSEHandler_StreamAccountStatus_GatewayError(t *testing.T) {
	accountID := uuid.New()
	ctx := context.WithValue(context.Background(), contextkeys.AccountID, accountID)

	mock := &mockSSEGateway{sendErr: context.DeadlineExceeded}
	h := &SSEHandler{gateway: mock}

	err := h.StreamAccountStatus(ctx, "", func(event string, payload any) error {
		return nil
	})
	if err == nil {
		t.Error("expected error from gateway")
	}
}
