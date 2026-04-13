package event

const (
	DocumentIngestionRequestedV1     = "document.ingestion.requested.v1"
	DocumentIngestionStatusUpdatedV1 = "document.ingestion.status.updated.v1"
	DocumentLifecycleArchivedV1      = "document.lifecycle.archived.v1"
	DocumentLifecycleRemovedV1       = "document.lifecycle.removed.v1"
)

const (
	EnvelopeSchemaVersion = "1.0.0"

	IngestionRequestedPayloadSchemaVersion     = "1.0.0"
	IngestionStatusUpdatedPayloadSchemaVersion = "1.0.0"
	LifecycleArchivedPayloadSchemaVersion      = "1.0.0"
	LifecycleRemovedPayloadSchemaVersion       = "1.0.0"

	ProducerCoreBackend = "core-backend"
	ProducerAIService   = "ai-service"
)
