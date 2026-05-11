package chapa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.chapa.co/v1"
	contentType    = "application/json"
)

// client implements the Client interface.
type client struct {
	httpClient *http.Client
	config     Config
}

// NewClient creates a new Chapa API client.
func NewClient(cfg Config) Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config: Config{
			SecretKey: cfg.SecretKey,
			BaseURL:   baseURL,
		},
	}
}

// Client defines the Chapa API operations.
type Client interface {
	// InitializeTransaction creates a new payment and returns a checkout URL.
	InitializeTransaction(ctx context.Context, req *InitRequest) (*InitResponse, error)

	// VerifyTransaction checks the current status of a payment by its tx_ref.
	VerifyTransaction(ctx context.Context, txRef string) (*VerifyResponse, error)
}

// InitializeTransaction calls POST /v1/transaction/initialize.
func (c *client) InitializeTransaction(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	url := c.config.BaseURL + "/transaction/initialize"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal init request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create init request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.SecretKey)
	httpReq.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chapa init request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read init response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp.StatusCode, respBody)
	}

	var initResp InitResponse
	if err := json.Unmarshal(respBody, &initResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal init response: %w", err)
	}

	if initResp.Status != "success" {
		return &initResp, &Error{
			HTTPStatus:  resp.StatusCode,
			Message:     initResp.Message,
			RawResponse: respBody,
		}
	}

	return &initResp, nil
}

// VerifyTransaction calls GET /v1/transaction/verify/{tx_ref}.
func (c *client) VerifyTransaction(ctx context.Context, txRef string) (*VerifyResponse, error) {
	url := fmt.Sprintf("%s/transaction/verify/%s", c.config.BaseURL, txRef)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create verify request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.SecretKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chapa verify request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verify response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp.StatusCode, respBody)
	}

	var verifyResp VerifyResponse
	if err := json.Unmarshal(respBody, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify response: %w", err)
	}

	if verifyResp.Status != "success" {
		return &verifyResp, &Error{
			HTTPStatus:  resp.StatusCode,
			Message:     verifyResp.Message,
			RawResponse: respBody,
		}
	}

	return &verifyResp, nil
}

// parseError attempts to extract a structured error from Chapa's response.
func parseError(statusCode int, body []byte) *Error {
	err := &Error{
		HTTPStatus:  statusCode,
		RawResponse: body,
	}

	// Try to parse Chapa's error format
	var chapaErr struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &chapaErr) == nil && chapaErr.Message != "" {
		err.Message = chapaErr.Message
	} else {
		err.Message = fmt.Sprintf("chapa API returned HTTP %d", statusCode)
	}

	return err
}
