package chapa

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Test secrets — not real credentials.
// #nosec G101
const (
	testSecret        = "whsec_test_secret_key"
	testCorrectSecret = "whsec_correct_secret"
	testWrongSecret   = "whsec_wrong_secret"
)

func TestVerifySignature_Valid(t *testing.T) {
	payload := []byte(`{"event":"charge.success","tx_ref":"tx_abc123"}`)

	// Generate valid signature
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	if !VerifySignature(payload, validSig, testSecret) {
		t.Fatal("expected valid signature to be accepted")
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	payload := []byte(`{"event":"charge.success","tx_ref":"tx_abc123"}`)

	if VerifySignature(payload, "invalid_signature", testSecret) {
		t.Fatal("expected invalid signature to be rejected")
	}
}

func TestVerifySignature_EmptySignature(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)

	if VerifySignature(payload, "", testSecret) {
		t.Fatal("expected empty signature to be rejected")
	}
}

func TestVerifySignature_EmptySecret(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)
	mac := hmac.New(sha256.New, []byte(""))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	if VerifySignature(payload, sig, "") {
		t.Fatal("expected empty secret to be rejected")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"event":"charge.success","tx_ref":"tx_abc123"}`)

	mac := hmac.New(sha256.New, []byte(testCorrectSecret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	if VerifySignature(payload, validSig, testWrongSecret) {
		t.Fatal("expected signature with wrong secret to be rejected")
	}
}

func TestVerifySignature_TamperedPayload(t *testing.T) {
	originalPayload := []byte(`{"event":"charge.success","tx_ref":"tx_abc123"}`)
	tamperedPayload := []byte(`{"event":"charge.success","tx_ref":"tx_tampered"}`)

	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(originalPayload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	if VerifySignature(tamperedPayload, validSig, testSecret) {
		t.Fatal("expected tampered payload to be rejected")
	}
}

func TestVerifySignatureFromHeaders_Valid(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)

	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	headers := map[string]string{
		"x-chapa-signature": sig,
	}

	if err := VerifySignatureFromHeaders(payload, headers, testSecret); err != nil {
		t.Fatalf("expected valid headers to pass, got: %v", err)
	}
}

func TestVerifySignatureFromHeaders_ChapaSignatureHeader(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)

	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	headers := map[string]string{
		"chapa-signature": sig,
	}

	if err := VerifySignatureFromHeaders(payload, headers, testSecret); err != nil {
		t.Fatalf("expected chapa-signature header to be accepted, got: %v", err)
	}
}

func TestVerifySignatureFromHeaders_MissingHeader(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)

	headers := map[string]string{}

	if err := VerifySignatureFromHeaders(payload, headers, testSecret); err == nil {
		t.Fatal("expected missing header to fail")
	}
}

func TestVerifySignatureFromHeaders_MissingSecret(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)
	headers := map[string]string{
		"x-chapa-signature": "some_sig",
	}

	if err := VerifySignatureFromHeaders(payload, headers, ""); err == nil {
		t.Fatal("expected missing secret to fail")
	}
}
