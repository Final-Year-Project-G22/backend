package server

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/domain/service"
	coreaitoolv1 "github.com/Final-Year-Project-G22/backend/core/pb/core/ai_tool/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AIToolService struct {
	coreaitoolv1.UnimplementedAIToolServiceServer

	registry *service.ToolRegistry
}

func NewAIToolService(registry *service.ToolRegistry) *AIToolService {
	return &AIToolService{registry: registry}
}

func (s *AIToolService) ListTools(
	ctx context.Context,
	req *coreaitoolv1.ListToolsRequest,
) (*coreaitoolv1.ListToolsResponse, error) {
	handlers := s.registry.List()
	tools := make([]*coreaitoolv1.ToolDefinition, 0, len(handlers))
	for _, h := range handlers {
		tools = append(tools, &coreaitoolv1.ToolDefinition{
			Name:                h.Name(),
			Description:         h.Description(),
			ParameterSchemaJson: h.ParameterSchema(),
		})
	}
	return &coreaitoolv1.ListToolsResponse{Tools: tools}, nil
}

func (s *AIToolService) ExecuteTool(
	ctx context.Context,
	req *coreaitoolv1.ExecuteToolRequest,
) (*coreaitoolv1.ExecuteToolResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	handler, ok := s.registry.Get(req.GetTool())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "tool %q not found", req.GetTool())
	}

	accountID, err := uuid.Parse(req.GetAccountId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := handler.Execute(ctx, req.GetArgumentsJson(), accountID, userID)
	if err != nil {
		return &coreaitoolv1.ExecuteToolResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &coreaitoolv1.ExecuteToolResponse{
		Success:    true,
		ResultJson: result,
	}, nil
}
