package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type ComplianceHandler struct {
	complianceUC usecase.ComplianceEntryUsecase
	ctRepo       repository.ComplianceTypeRepository
}

func NewComplianceHandler(complianceUC usecase.ComplianceEntryUsecase, ctRepo repository.ComplianceTypeRepository) *ComplianceHandler {
	return &ComplianceHandler{complianceUC: complianceUC, ctRepo: ctRepo}
}

func (h *ComplianceHandler) HandleList(ctx context.Context, input *dto.ListComplianceEntriesInput) (*dto.ListComplianceEntriesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	entries, err := h.complianceUC.ListByBusinessProfile(ctx, accountID, input.BusinessProfileID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ListComplianceEntriesOutput{Body: dto.ListComplianceEntriesResponseBody{
		Data: dto.ToComplianceEntryResponses(entries),
	}}, nil
}

func (h *ComplianceHandler) HandleCreate(ctx context.Context, input *dto.CreateComplianceEntryInput) (*dto.CreateComplianceEntryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	entry, err := h.complianceUC.Create(ctx, accountID, dto.ToCreateComplianceEntryInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateComplianceEntryOutput{Body: dto.CreateComplianceEntryResponseBody{
		ID:      entry.ID,
		Message: "Compliance entry created",
	}}, nil
}

func (h *ComplianceHandler) HandleUpdate(ctx context.Context, input *dto.UpdateComplianceEntryInput) (*dto.UpdateComplianceEntryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.complianceUC.Update(ctx, accountID, input.ID, dto.ToUpdateComplianceEntryInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateComplianceEntryOutput{Body: dto.UpdateComplianceEntryResponseBody{
		Message: "Compliance entry updated",
	}}, nil
}

func (h *ComplianceHandler) HandleDelete(ctx context.Context, input *dto.DeleteComplianceEntryInput) (*dto.DeleteComplianceEntryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.complianceUC.Delete(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteComplianceEntryOutput{Body: dto.DeleteComplianceEntryResponseBody{
		Message: "Compliance entry deleted",
	}}, nil
}

func (h *ComplianceHandler) HandleGetCalendar(ctx context.Context, input *struct{}) (*dto.GetCalendarOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	calendar, err := h.complianceUC.GetCalendar(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetCalendarOutput{Body: dto.GetCalendarResponseBody{
		Entries: dto.ToCalendarEntryResponses(calendar.Entries),
	}}, nil
}

type listTypesInput struct {
	Locale string `query:"locale"`
}

func (h *ComplianceHandler) HandleListTypes(ctx context.Context, input *listTypesInput) (*dto.ListComplianceTypesOutput, error) {
	locale := input.Locale
	if locale == "" {
		locale = "en"
	}
	types, err := h.ctRepo.ListWithLabels(ctx, locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	resp := make([]dto.ComplianceTypeResponse, 0, len(types))
	for _, t := range types {
		resp = append(resp, dto.ComplianceTypeResponse{Slug: t.Slug, Label: t.Label})
	}
	return &dto.ListComplianceTypesOutput{Body: dto.ListComplianceTypesResponseBody{
		Data: resp,
	}}, nil
}
