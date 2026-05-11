package chapa

// Package chapa provides a custom HTTP client for the Chapa ET payment gateway API.
// It wraps the Chapa REST endpoints needed for subscription payment processing:
//
//   - InitializeTransaction: POST /v1/transaction/initialize
//   - VerifyTransaction:     GET /v1/transaction/verify/{tx_ref}
//
// The client is intentionally minimal — it covers only the endpoints required
// by the payment module. Webhook signature verification is provided via
// VerifySignature.
//
// Usage:
//
//	client := chapa.NewClient(chapa.Config{
//	    SecretKey: "CHASECK-xxxxxxxxxxxxxxxx",
//	    BaseURL:   "https://api.chapa.co/v1",
//	})
//	resp, err := client.InitializeTransaction(ctx, &chapa.InitRequest{...})
//
// All methods accept context.Context for cancellation and timeout control.
//
