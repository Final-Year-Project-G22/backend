package chapa

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// VerifySignature checks that the HMAC-SHA256 signature in the header matches
// the payload signed with the secret key. Chapa may send the signature in either
// the "x-chapa-signature" or "chapa-signature" header.
//
// The signature is computed as: hex(HMAC-SHA256(secret, rawBody))
func VerifySignature(payload []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

// VerifySignatureFromHeaders is a convenience wrapper that checks both header names.
func VerifySignatureFromHeaders(payload []byte, headers map[string]string, secret string) error {
	if secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	// Check both header names (Chapa sends either)
	sig := headers["x-chapa-signature"]
	if sig == "" {
		sig = headers["chapa-signature"]
	}

	if sig == "" {
		return fmt.Errorf("missing chapa signature header")
	}

	if !VerifySignature(payload, sig, secret) {
		return fmt.Errorf("invalid chapa signature")
	}

	return nil
}
