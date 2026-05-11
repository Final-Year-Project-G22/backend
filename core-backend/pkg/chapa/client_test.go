package chapa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInitializeTransaction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/transaction/initialize" {
			t.Errorf("expected path /transaction/initialize, got %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test_secret" {
			t.Errorf("expected Authorization: Bearer test_secret, got %s", auth)
		}

		resp := InitResponse{
			Message: "Hosted Link",
			Status:  "success",
			Data: InitResponseData{
				CheckoutURL: "https://checkout.chapa.co/checkout/test-123",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{SecretKey: "test_secret", BaseURL: server.URL})
	req := &InitRequest{
		Amount:   19900,
		Currency: "ETB",
		TxRef:    "tx_test_123",
		Email:    "test@example.com",
	}

	resp, err := client.InitializeTransaction(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.CheckoutURL != "https://checkout.chapa.co/checkout/test-123" {
		t.Errorf("expected checkout URL, got %s", resp.Data.CheckoutURL)
	}
}

func TestInitializeTransaction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid currency"})
	}))
	defer server.Close()

	client := NewClient(Config{SecretKey: "test_secret", BaseURL: server.URL})
	req := &InitRequest{Amount: 100, Currency: "USD", TxRef: "tx_test"}

	_, err := client.InitializeTransaction(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
	chapaErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *chapa.Error, got %T", err)
	}
	if chapaErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", chapaErr.HTTPStatus)
	}
}

func TestVerifyTransaction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		expectedPath := "/transaction/verify/tx_test_verify"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		resp := VerifyResponse{
			Message: "Payment details",
			Status:  "success",
			Data: VerifyData{
				TxRef:         "tx_test_verify",
				Reference:     "AP123456",
				Status:        "success",
				Amount:        "199.00",
				Currency:      "ETB",
				PaymentMethod: "telebirr",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{SecretKey: "test_secret", BaseURL: server.URL})
	resp, err := client.VerifyTransaction(context.Background(), "tx_test_verify")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Data.Status)
	}
	if resp.Data.Reference != "AP123456" {
		t.Errorf("expected reference AP123456, got %s", resp.Data.Reference)
	}
}

func TestVerifyTransaction_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "transaction not found"})
	}))
	defer server.Close()

	client := NewClient(Config{SecretKey: "test_secret", BaseURL: server.URL})
	_, err := client.VerifyTransaction(context.Background(), "tx_nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewClient_DefaultBaseURL(t *testing.T) {
	c, ok := NewClient(Config{SecretKey: "test_secret"}).(*client)
	if !ok {
		t.Fatal("expected *client type")
	}
	if c.config.BaseURL != defaultBaseURL {
		t.Errorf("expected default base URL %s, got %s", defaultBaseURL, c.config.BaseURL)
	}
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	customURL := "https://custom.chapa.co/v1"
	c, ok := NewClient(Config{SecretKey: "test_secret", BaseURL: customURL}).(*client)
	if !ok {
		t.Fatal("expected *client type")
	}
	if c.config.BaseURL != customURL {
		t.Errorf("expected custom base URL %s, got %s", customURL, c.config.BaseURL)
	}
}

func TestError_ErrorString(t *testing.T) {
	err := &Error{
		HTTPStatus:  400,
		Message:     "invalid currency",
		RawResponse: []byte(`{"message":"invalid currency"}`),
	}
	expected := "chapa error: status=400, message=invalid currency"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsError(t *testing.T) {
	chapaErr := &Error{HTTPStatus: 400, Message: "test"}
	if !IsError(chapaErr) {
		t.Error("expected IsError to return true for *chapa.Error")
	}
	if IsError(fmt.Errorf("regular error")) {
		t.Error("expected IsError to return false for regular error")
	}
}

func TestErrorStatus(t *testing.T) {
	chapaErr := &Error{HTTPStatus: 404, Message: "not found"}
	if ErrorStatus(chapaErr) != 404 {
		t.Errorf("expected status 404, got %d", ErrorStatus(chapaErr))
	}
	if ErrorStatus(fmt.Errorf("regular")) != 0 {
		t.Error("expected status 0 for non-chapa error")
	}
}

func TestInitializeTransaction_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{SecretKey: "test_secret", BaseURL: server.URL})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.InitializeTransaction(ctx, &InitRequest{TxRef: "tx_test"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
