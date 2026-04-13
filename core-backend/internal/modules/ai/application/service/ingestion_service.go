package service

import (
	"context"
)

type IngestionService struct{}

func NewIngestionService() *IngestionService {
	return &IngestionService{}
}

func (s *IngestionService) Ping(_ context.Context) error {
	return nil
}
