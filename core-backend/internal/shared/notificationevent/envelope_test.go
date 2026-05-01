package notificationevent

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validAccountID() uuid.UUID {
	return uuid.MustParse("d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9")
}

func validSingleEnvelopeJSON() string {
	return `{
		"schemaVersion": "1.0.0",
		"eventType": "user.email_otp_requested",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "account_verification",
		"channelPolicy": "single",
		"channel": "email",
		"variables": {
			"platformName": "TestPlatform",
			"verificationMessage": "Click the link to verify",
			"verificationUrl": "https://example.com/verify",
			"expiryMinutes": "10"
		},
		"metadata": {
			"idempotencyKey": "verify-email:d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9:otp-123",
			"locale": "en",
			"traceId": "trace-abc"
		}
	}`
}

func validAllEnabledEnvelopeJSON() string {
	return `{
		"schemaVersion": "1.0.0",
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {
			"platformName": "TestPlatform",
			"accountName": "Jane",
			"gettingStartedUrl": "https://example.com/start"
		},
		"metadata": {
			"idempotencyKey": "welcome:d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
			"locale": "en"
		}
	}`
}

func TestParse_SingleChannel(t *testing.T) {
	env, err := Parse([]byte(validSingleEnvelopeJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.SchemaVersion != SchemaVersionV1 {
		t.Fatalf("expected schemaVersion %s, got %s", SchemaVersionV1, env.SchemaVersion)
	}
	if env.ChannelPolicy != ChannelPolicySingle {
		t.Fatalf("expected channelPolicy single, got %s", env.ChannelPolicy)
	}
	if env.Channel == nil || *env.Channel != "email" {
		t.Fatalf("expected channel email, got %v", env.Channel)
	}
	if env.NotificationType != "account_verification" {
		t.Fatalf("expected account_verification, got %s", env.NotificationType)
	}
	if env.AccountID != validAccountID() {
		t.Fatalf("expected accountId %s, got %s", validAccountID(), env.AccountID)
	}
	if env.Metadata.IdempotencyKey != "verify-email:d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9:otp-123" {
		t.Fatalf("unexpected idempotencyKey: %s", env.Metadata.IdempotencyKey)
	}
	if env.Metadata.Locale == nil || *env.Metadata.Locale != "en" {
		t.Fatalf("expected locale en, got %v", env.Metadata.Locale)
	}
	if len(env.Variables) != 4 {
		t.Fatalf("expected 4 variables, got %d", len(env.Variables))
	}
}

func TestParse_AllEnabled(t *testing.T) {
	env, err := Parse([]byte(validAllEnabledEnvelopeJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.ChannelPolicy != ChannelPolicyAllEnabled {
		t.Fatalf("expected channelPolicy all_enabled, got %s", env.ChannelPolicy)
	}
	if env.Channel != nil {
		t.Fatalf("expected nil channel for all_enabled, got %v", *env.Channel)
	}
	if env.NotificationType != "welcome_message" {
		t.Fatalf("expected welcome_message, got %s", env.NotificationType)
	}
	if env.SourceModule != "iam" {
		t.Fatalf("expected sourceModule iam, got %s", env.SourceModule)
	}
}

func TestParse_VariablesDefaultsToEmptyMap(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.alert",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "account_alert",
		"channelPolicy": "all_enabled",
		"metadata": {
			"idempotencyKey": "alert-1"
		}
	}`
	env, err := Parse([]byte(j))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env.Variables) != 0 {
		t.Fatalf("expected empty variables, got %v", env.Variables)
	}
}

func TestParse_MetadataExtraFields(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.alert",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "account_alert",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "alert-1",
			"alertCode": "new_device_login",
			"severity": "critical"
		}
	}`
	env, err := Parse([]byte(j))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := env.Metadata.Extra["alertCode"]; !ok || v != "new_device_login" {
		t.Fatalf("expected alertCode=new_device_login in extras, got %v", env.Metadata.Extra)
	}
	if v, ok := env.Metadata.Extra["severity"]; !ok || v != "critical" {
		t.Fatalf("expected severity=critical in extras, got %v", env.Metadata.Extra)
	}
}

func TestParse_MetadataRoundTrip(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.alert",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "account_alert",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "alert-1",
			"locale": "am",
			"alertCode": "new_device_login"
		}
	}`
	env, err := Parse([]byte(j))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metadataMap := env.MetadataMap()
	if metadataMap["idempotencyKey"] != "alert-1" {
		t.Fatalf("idempotencyKey mismatch in metadata map")
	}
	if metadataMap["locale"] != "am" {
		t.Fatalf("locale mismatch in metadata map")
	}
	if metadataMap["alertCode"] != "new_device_login" {
		t.Fatalf("alertCode mismatch in metadata map")
	}
}

func TestParse_MissingSchemaVersion(t *testing.T) {
	j := `{
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for missing schemaVersion")
	}
}

func TestParse_InvalidSchemaVersion(t *testing.T) {
	j := `{
		"schemaVersion": "0.0.1",
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for invalid schemaVersion")
	}
}

func TestParse_NilAccountID(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "00000000-0000-0000-0000-000000000000",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for nil accountId")
	}
}

func TestParse_SingleMissingChannel(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "user.email_otp_requested",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "account_verification",
		"channelPolicy": "single",
		"variables": {},
		"metadata": {
			"idempotencyKey": "otp-1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for single policy without channel")
	}
}

func TestParse_AllEnabledWithChannel(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"channel": "email",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for all_enabled with channel")
	}
}

func TestParse_InvalidChannelPolicy(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "broadcast",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for invalid channelPolicy")
	}
}

func TestParse_MissingIdempotencyKey(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.registered",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"locale": "en"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for missing idempotencyKey")
	}
}

func TestParse_MissingEventType(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"occurredAt": "2026-04-30T12:34:56Z",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for missing eventType")
	}
}

func TestParse_MissingOccurredAt(t *testing.T) {
	j := `{
		"schemaVersion": "1.0.0",
		"eventType": "account.registered",
		"sourceModule": "iam",
		"accountId": "d4e5f6a7-b8c9-40d1-a2b3-c4d5e6f7a8b9",
		"notificationType": "welcome_message",
		"channelPolicy": "all_enabled",
		"variables": {},
		"metadata": {
			"idempotencyKey": "welcome:1"
		}
	}`
	_, err := Parse([]byte(j))
	if err == nil {
		t.Fatalf("expected error for missing occurredAt")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestEnvelope_MarshalJSON_PreservesExtraMetadata(t *testing.T) {
	loc := "en"
	env := Envelope{
		SchemaVersion:    SchemaVersionV1,
		EventType:        "account.alert",
		OccurredAt:       time.Date(2026, 4, 30, 12, 34, 56, 0, time.UTC),
		SourceModule:     "iam",
		AccountID:        validAccountID(),
		NotificationType: "account_alert",
		ChannelPolicy:    ChannelPolicyAllEnabled,
		Variables:        map[string]string{"key": "val"},
		Metadata: Metadata{
			IdempotencyKey: "alert-1",
			Locale:         &loc,
			Extra:          map[string]string{"alertCode": "new_device_login"},
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("failed to parse marshaled envelope: %v", err)
	}
	if v, ok := parsed.Metadata.Extra["alertCode"]; !ok || v != "new_device_login" {
		t.Fatalf("extra metadata field lost in round-trip")
	}
}

func TestEnvelope_MetadataMap_NoNilPanic(t *testing.T) {
	emailCh := "email"
	env := Envelope{
		SchemaVersion:    SchemaVersionV1,
		EventType:        "test.event",
		OccurredAt:       time.Now(),
		SourceModule:     "iam",
		AccountID:        validAccountID(),
		NotificationType: "welcome_message",
		ChannelPolicy:    ChannelPolicySingle,
		Channel:          &emailCh,
		Variables:        map[string]string{},
		Metadata: Metadata{
			IdempotencyKey: "test-key",
		},
	}

	m := env.MetadataMap()
	if m["idempotencyKey"] != "test-key" {
		t.Fatalf("idempotencyKey not in metadata map")
	}
	if _, ok := m["locale"]; ok {
		t.Fatalf("locale should not be in metadata map when nil")
	}
}
