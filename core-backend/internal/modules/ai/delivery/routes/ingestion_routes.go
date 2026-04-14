package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const (
	aiIngestionBase = "/api/v1/ai/ingestion"
)

func registerIngestionRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "createIngestionUploadIntent",
		Method:      "POST",
		Path:        aiIngestionBase + "/uploads/intents",
		Summary:     "Create direct upload intent",
		Description: "Generates a SeaweedFS direct-upload URL and required headers.",
		Tags:        []string{"AI Ingestion"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.IngestionHandler.HandleCreateUploadIntent)

	huma.Register(api, huma.Operation{
		OperationID: "finalizeIngestionUpload",
		Method:      "POST",
		Path:        aiIngestionBase + "/uploads/finalize",
		Summary:     "Finalize uploaded document ingestion",
		Description: "Validates uploaded object metadata, writes ingestion metadata and outbox entry atomically.",
		Tags:        []string{"AI Ingestion"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.IngestionHandler.HandleFinalizeUpload)
}
