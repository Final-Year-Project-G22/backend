# Chapa Integration — `pkg/chapa/` Client Design

## Purpose

The `pkg/chapa/` package is a thin, custom HTTP client wrapping the Chapa ET REST API. It follows the same pattern as existing packages (`pkg/email/`, `pkg/storage/`) with an interface, a real implementation, and a no-op fallback.

## Directory Structure

```
pkg/chapa/
├── client.go       # Client interface + HTTP implementation
├── types.go        # Request/response DTOs
├── webhook.go      # Signature verification helper
└── doc.go          # Package documentation
```

## Interface

```go
package chapa

type Client interface {
    // InitializeTransaction creates a new payment and returns a checkout URL.
    InitializeTransaction(ctx context.Context, req *InitRequest) (*InitResponse, error)

    // VerifyTransaction checks the current status of a payment by its tx_ref.
    VerifyTransaction(ctx context.Context, txRef string) (*VerifyResponse, error)
}
```

## Types

```go
// InitRequest is the payload for POST /v1/transaction/initialize
type InitRequest struct {
    Amount        int64                  // Amount in minor unit (satcker)
    Currency      string                 // "ETB"
    Email         string                 // Customer email
    FirstName     string                 // Customer first name
    LastName      string                 // Customer last name
    Phone         string                 // Customer phone (e.g., "0912345678")
    TxRef         string                 // Backend-generated transaction reference
    CallbackURL   string                 // Chapa callback URL (webhook)
    ReturnURL     string                 // Redirect URL after payment (for web checkout)
    Customization map[string]interface{} // Optional: title, description, logo
}

// InitResponse mirrors Chapa's response to POST /v1/transaction/initialize
type InitResponse struct {
    Message  string           `json:"message"`
    Status   string           `json:"status"`   // "success" or "error"
    Data     InitResponseData `json:"data"`
}

type InitResponseData struct {
    CheckoutURL string `json:"checkout_url"`
}

// VerifyResponse mirrors Chapa's response to GET /v1/transaction/verify/<tx_ref>
type VerifyResponse struct {
    Message    string `json:"message"`
    Status     string `json:"status"`
    Data       VerifyData `json:"data"`
}

type VerifyData struct {
    TxRef         string    `json:"tx_ref"`
    Reference     string    `json:"reference"`
    Status        string    `json:"status"`        // "success", "pending", "failed"
    Amount        string    `json:"amount"`        // String from Chapa
    Charge        string    `json:"charge"`
    Currency      string    `json:"currency"`
    PaymentMethod string    `json:"payment_method"` // "telebirr", "cbebirr", etc.
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

## HTTP Client Implementation

**Base URL:** `https://api.chapa.co/v1` (configurable via `ChapaConfig.BaseURL`)

**Authentication:** Bearer token using the secret key:
```
Authorization: Bearer CHASECK-xxxxxxxxxxxxxxxx
```

**Implementation approach:**
- Uses Go's standard `net/http` package (no external HTTP client library)
- `NewClient(cfg chapaConfig) Client` creates the client
- Each method constructs an `*http.Request`, sets headers, and decodes the JSON response
- All methods accept a `context.Context` for timeout/cancellation
- Errors return a custom `ErrChapa` type wrapping the HTTP status code and Chapa's error message

### Error handling

```go
type Error struct {
    HTTPStatus  int
    Code        string   // Chapa error code if available
    Message     string   // Human-readable message
    RawResponse []byte   // Raw response body for debugging
}
```

## Webhook Verification

Located in `webhook.go`:

```go
package chapa

// VerifySignature checks that the HMAC-SHA256 signature in the
// `x-chapa-signature` header matches the payload signed with the secret key.
func VerifySignature(payload []byte, signature string, secret string) bool
```

**Verification method:**
1. Compute `HMAC-SHA256(secret, rawRequestBody)` → hex string
2. Compare with the value of the `x-chapa-signature` header
3. Either `x-chapa-signature` or `chapa-signature` header is acceptable

**Usage in HTTP handler:**
```go
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    sig := r.Header.Get("x-chapa-signature")
    if !chapa.VerifySignature(body, sig, h.secret) {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }
    // Parse event, handle idempotency, process...
}
```

## Integration with the Payment Module

The payment module publishes `payment.confirmation` events via the existing **Notification Outbox** pattern (not directly to RabbitMQ). This is the same pattern used by the IAM module.

```
internal/modules/payment/application/usecase/payment_usecase.go
    │
    ├── Depends on: chapa.Client (interface)
    │   Depends on: payment repository
    │   Depends on: notification outbox repository
    │
    ├── InitiatePayment(ctx, req):
    │   1. Validate plan exists (query Plan from DB)
    │   2. Generate tx_ref: fmt.Sprintf("tx_%s_%s_%s_%d", accountID, planName, period, timestamp)
    │   3. Create Payment record in DB (status: pending)
    │   4. Call chapaClient.InitializeTransaction(ctx, initReq)
    │   5. Return checkout URL to caller
    │
    ├── VerifyPayment(ctx, txRef):
    │   1. Find Payment by txRef in DB
    │   2. Call chapaClient.VerifyTransaction(ctx, txRef)
    │   3. Update Payment status based on Chapa response
    │   4. If success: create/update Subscription, write notification outbox entry
    │   5. Return verification result
    │
    └── HandleWebhook(ctx, payload):
        1. Verify signature via chapa.VerifySignature()
        2. Parse webhook event
        3. Find Payment by txRef
        4. If already terminal (success/failed): return (idempotent)
        5. Call chapaClient.VerifyTransaction() (re-verify per best practice)
        6. Update Payment, Subscription
        7. Write notification outbox entry (payment.confirmation event)
```

### Event Publishing (Outbox Pattern)

The payment module **does not publish directly to RabbitMQ**. Instead, it writes a `NotificationOutbox` row, which the existing `NotificationOutboxDispatcher` worker picks up and publishes to the message bus:

```
Payment Use Case
    │
    ├── 1. Write NotificationOutbox row:
    │      eventType: "payment.confirmation"
    │      accountID: <account>
    │      payload: { "tx_ref": "...", "plan": "Pro", "amount": 19900, ... }
    │      idempotencyKey: "payment:<txRef>"
    │
    ├── 2. NotificationOutboxDispatcher (worker, runs every 5s):
    │      Reads pending outbox rows
    │      Publishes to RabbitMQ (eventType = routing key)
    │      Marks row as published
    │
    └── 3. Notification module subscriber receives event:
           Matches "payment.confirmation" → NotificationTypePaymentConfirmation
           Resolves template → dispatches in-app + email
```

Using the outbox pattern ensures notifications are **reliable** (retry with backoff) and **decoupled** (payment module doesn't need to import the notification module).