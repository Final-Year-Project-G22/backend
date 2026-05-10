package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	aievent "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/event"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
)

type StatusEventSubscriber struct {
	bus     rabbitmq.Bus
	logger  core.Logger
	service *service.StatusProjectionService
}

func NewStatusEventSubscriber(
	bus rabbitmq.Bus,
	logger core.Logger,
	service *service.StatusProjectionService,
) *StatusEventSubscriber {
	return &StatusEventSubscriber{
		bus:     bus,
		logger:  logger,
		service: service,
	}
}

func (s *StatusEventSubscriber) Subscribe() error {
	return s.bus.Subscribe(
		aievent.DocumentIngestionStatusUpdatedV1,
		s.handleEvent,
	)
}

func (s *StatusEventSubscriber) handleEvent(ctx context.Context, body []byte) error {
	var envelope service.StatusEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		s.logger.Error("failed to unmarshal status event envelope", core.Error(err))
		return fmt.Errorf("invalid envelope: %w", err)
	}

	if err := s.service.ConsumeStatusEvent(ctx, envelope); err != nil {
		s.logger.Error(
			"failed to consume status event",
			core.Error(err),
			core.String("document_id", envelope.Payload.DocumentID),
		)
		return fmt.Errorf("consume failed: %w", err)
	}

	s.logger.Info(
		"status event consumed",
		core.String("document_id", envelope.Payload.DocumentID),
		core.String("to_stage", envelope.Payload.ToStage),
	)
	return nil
}
