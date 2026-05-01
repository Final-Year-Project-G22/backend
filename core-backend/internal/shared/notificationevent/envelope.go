package notificationevent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

const SchemaVersionV1 = "1.0.0"

type ChannelPolicy string

const (
	ChannelPolicySingle     ChannelPolicy = "single"
	ChannelPolicyAllEnabled ChannelPolicy = "all_enabled"
)

type Metadata struct {
	IdempotencyKey string            `json:"idempotencyKey"`
	Locale         *string           `json:"locale,omitempty"`
	TraceID        *string           `json:"traceId,omitempty"`
	Extra          map[string]string `json:"-"`
}

func (m *Metadata) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Extra = make(map[string]string)
	for key, val := range raw {
		switch key {
		case "idempotencyKey":
			if s, ok := val.(string); ok {
				m.IdempotencyKey = s
			}
		case "locale":
			if s, ok := val.(string); ok {
				m.Locale = &s
			}
		case "traceId":
			if s, ok := val.(string); ok {
				m.TraceID = &s
			}
		default:
			if s, ok := val.(string); ok {
				m.Extra[key] = s
			}
		}
	}
	return nil
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	out := make(map[string]interface{})
	out["idempotencyKey"] = m.IdempotencyKey
	if m.Locale != nil {
		out["locale"] = *m.Locale
	}
	if m.TraceID != nil {
		out["traceId"] = *m.TraceID
	}
	for k, v := range m.Extra {
		out[k] = v
	}
	return json.Marshal(out)
}

type Envelope struct {
	SchemaVersion    string                  `json:"schemaVersion"`
	EventType        string                  `json:"eventType"`
	OccurredAt       time.Time               `json:"occurredAt"`
	SourceModule     string                  `json:"sourceModule"`
	AccountID        uuid.UUID               `json:"accountId"`
	NotificationType entity.NotificationType `json:"notificationType"`
	ChannelPolicy    ChannelPolicy           `json:"channelPolicy"`
	Channel          *entity.Channel         `json:"channel,omitempty"`
	Variables        map[string]string       `json:"variables"`
	Metadata         Metadata                `json:"metadata"`
}

func Parse(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("invalid canonical envelope: %w", err)
	}
	if err := env.Validate(); err != nil {
		return nil, err
	}
	return &env, nil
}

func (e *Envelope) Validate() error {
	if strings.TrimSpace(e.SchemaVersion) == "" {
		return fmt.Errorf("missing_required_field: schemaVersion")
	}
	if e.SchemaVersion != SchemaVersionV1 {
		return fmt.Errorf("invalid_schema_version: %s", e.SchemaVersion)
	}
	if strings.TrimSpace(e.EventType) == "" {
		return fmt.Errorf("missing_required_field: eventType")
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("missing_required_field: occurredAt")
	}
	if strings.TrimSpace(e.SourceModule) == "" {
		return fmt.Errorf("missing_required_field: sourceModule")
	}
	if e.AccountID == uuid.Nil {
		return fmt.Errorf("invalid_account_id")
	}
	if strings.TrimSpace(string(e.NotificationType)) == "" {
		return fmt.Errorf("missing_required_field: notificationType")
	}
	if strings.TrimSpace(e.Metadata.IdempotencyKey) == "" {
		return fmt.Errorf("invalid_idempotency_key")
	}

	switch e.ChannelPolicy {
	case ChannelPolicySingle:
		if e.Channel == nil || *e.Channel == "" {
			return fmt.Errorf("missing_single_channel")
		}
	case ChannelPolicyAllEnabled:
		if e.Channel != nil {
			return fmt.Errorf("forbidden_channel_for_all_enabled")
		}
	default:
		return fmt.Errorf("invalid_channel_policy")
	}

	return nil
}

func (e *Envelope) MetadataMap() map[string]interface{} {
	out := make(map[string]interface{})
	out["idempotencyKey"] = e.Metadata.IdempotencyKey
	if e.Metadata.Locale != nil {
		out["locale"] = *e.Metadata.Locale
	}
	if e.Metadata.TraceID != nil {
		out["traceId"] = *e.Metadata.TraceID
	}
	for k, v := range e.Metadata.Extra {
		out[k] = v
	}
	return out
}
