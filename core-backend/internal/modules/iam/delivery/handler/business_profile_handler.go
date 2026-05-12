package handler

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BusinessProfileHandler struct {
	profileUsecase iamusecase.BusinessProfileUsecase
	accountUsecase iamusecase.AccountUsecase
	userUsecase    iamusecase.UserUsecase
	sectorRepo     repository.SectorRepository
	tagRepo        repository.TagRepository
}

func NewBusinessProfileHandler(
	profileUsecase iamusecase.BusinessProfileUsecase,
	accountUsecase iamusecase.AccountUsecase,
	userUsecase iamusecase.UserUsecase,
	sectorRepo repository.SectorRepository,
	tagRepo repository.TagRepository,
) *BusinessProfileHandler {
	return &BusinessProfileHandler{
		profileUsecase: profileUsecase,
		accountUsecase: accountUsecase,
		userUsecase:    userUsecase,
		sectorRepo:     sectorRepo,
		tagRepo:        tagRepo,
	}
}

// HandleGetBusinessProfile handles GET /api/v1/users/business-profile
func (h *BusinessProfileHandler) HandleGetBusinessProfile(ctx context.Context, _ *dto.GetBusinessProfileInput) (*dto.GetBusinessProfileOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	profile, err := h.profileUsecase.GetBusinessProfileByAccount(ctx, accountID)
	if err != nil {
		if err == iamerror.ErrBusinessProfileNotFound {
			return nil, apperrors.NotFoundErrorWithKey("iam.errors.businessProfileNotFound")
		}
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetBusinessProfileOutput{Body: h.toResponse(profile)}, nil
}

// HandleCreateBusinessProfile handles POST /api/v1/users/business-profile
func (h *BusinessProfileHandler) HandleCreateBusinessProfile(ctx context.Context, input *dto.CreateBusinessProfileInput) (*dto.CreateBusinessProfileOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	// Fetch account and user for auto-fill
	account, err := h.accountUsecase.GetAccount(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	user, err := h.userUsecase.GetUser(ctx, account.UserID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	// Resolve sector slug to ID
	var sectorID *uuid.UUID
	if input.Body.SectorSlug != nil && *input.Body.SectorSlug != "" {
		sid, err := h.resolveSectorSlug(ctx, *input.Body.SectorSlug)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
		sectorID = sid
	}

	// Resolve tag slugs to IDs
	tagIDs, err := h.resolveTagSlugs(ctx, input.Body.TagSlugs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	// Auto-fill defaults
	companyName := input.Body.CompanyName
	if companyName == "" {
		companyName = user.FirstName + " " + user.LastName + "'s Business"
	}
	companyEmail := input.Body.CompanyEmail
	if companyEmail == "" {
		companyEmail = account.Email
	}
	companyPhone := input.Body.CompanyPhoneNumber
	if companyPhone == "" {
		if account.PhoneNumber != nil {
			companyPhone = *account.PhoneNumber
		}
	}

	socialLinks := input.Body.SocialLinks
	if socialLinks == nil {
		socialLinks = datatypes.JSONMap{}
	}

	profile, err := h.profileUsecase.CreateBusinessProfile(ctx, accountID, iamusecase.CreateBusinessProfileInput{
		CompanyName:             companyName,
		CompanyEmail:            companyEmail,
		CompanyPhoneNumber:      companyPhone,
		PhysicalAddress:         input.Body.PhysicalAddress,
		Description:             input.Body.Description,
		LogoURL:                 input.Body.LogoURL,
		BannerURL:               input.Body.BannerURL,
		SocialLinks:             socialLinks,
		RegistrationNumber:      input.Body.RegistrationNumber,
		RegistrationDate:        input.Body.RegistrationDate,
		TaxIdentificationNumber: input.Body.TaxIdentificationNumber,
		TradeLicenseNumber:      input.Body.TradeLicenseNumber,
		SectorID:                sectorID,
		TagIDs:                  tagIDs,
		Region:                  input.Body.Region,
		Stage:                   input.Body.Stage,
	})
	if err != nil {
		if err == iamerror.ErrBusinessProfileAlreadyExists {
			return nil, apperrors.ConflictError("iam.errors.businessProfileAlreadyExists")
		}
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.CreateBusinessProfileOutput{Body: h.toResponse(profile)}, nil
}

// HandleUpdateBusinessProfile handles PUT /api/v1/users/business-profile
func (h *BusinessProfileHandler) HandleUpdateBusinessProfile(ctx context.Context, input *dto.UpdateBusinessProfileInput) (*dto.UpdateBusinessProfileOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	// Resolve sector slug to ID if provided
	var sectorID *uuid.UUID
	if input.Body.SectorSlug != nil && *input.Body.SectorSlug != "" {
		sid, err := h.resolveSectorSlug(ctx, *input.Body.SectorSlug)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
		sectorID = sid
	}

	// Resolve tag slugs to IDs if provided
	var tagIDs []uuid.UUID
	if input.Body.TagSlugs != nil {
		ids, err := h.resolveTagSlugs(ctx, input.Body.TagSlugs)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
		tagIDs = ids
	}

	profile, err := h.profileUsecase.UpdateBusinessProfile(ctx, accountID, iamusecase.UpdateBusinessProfileInput{
		CompanyName:             input.Body.CompanyName,
		CompanyEmail:            input.Body.CompanyEmail,
		CompanyPhoneNumber:      input.Body.CompanyPhoneNumber,
		PhysicalAddress:         input.Body.PhysicalAddress,
		Description:             input.Body.Description,
		LogoURL:                 input.Body.LogoURL,
		BannerURL:               input.Body.BannerURL,
		SocialLinks:             input.Body.SocialLinks,
		RegistrationNumber:      input.Body.RegistrationNumber,
		RegistrationDate:        input.Body.RegistrationDate,
		TaxIdentificationNumber: input.Body.TaxIdentificationNumber,
		TradeLicenseNumber:      input.Body.TradeLicenseNumber,
		SectorID:                sectorID,
		TagIDs:                  tagIDs,
		Region:                  input.Body.Region,
		Stage:                   input.Body.Stage,
	})
	if err != nil {
		if err == iamerror.ErrBusinessProfileNotFound {
			return nil, apperrors.NotFoundErrorWithKey("iam.errors.businessProfileNotFound")
		}
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateBusinessProfileOutput{Body: h.toResponse(profile)}, nil
}

// resolveSectorSlug looks up a sector by slug and returns its ID.
func (h *BusinessProfileHandler) resolveSectorSlug(ctx context.Context, slug string) (*uuid.UUID, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, nil
	}

	// Query only the sector ID. Avoid scanning all columns (including ancestor_ids),
	// which can fail on some drivers when mapped into []uuid.UUID.
	var sector entity.Sector
	err := h.sectorRepo.GetDB().WithContext(ctx).
		Model(&entity.Sector{}).
		Select("id").
		Where("slug = ?", slug).
		First(&sector).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.BadRequestError("iam.errors.invalidSectorSlug")
		}
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	if sector.ID == uuid.Nil {
		return nil, apperrors.BadRequestError("iam.errors.invalidSectorSlug")
	}
	return &sector.ID, nil
}

// resolveTagSlugs looks up tags by slugs and returns their IDs.
func (h *BusinessProfileHandler) resolveTagSlugs(ctx context.Context, slugs []string) ([]uuid.UUID, error) {
	if len(slugs) == 0 {
		return nil, nil
	}

	var tagIDs []uuid.UUID
	for _, slug := range slugs {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}
		tags, err := h.tagRepo.Find(ctx, query.QueryOptions{
			Filters: map[string]interface{}{"slug": slug},
		})
		if err != nil {
			return nil, err
		}
		if len(tags) == 0 {
			return nil, apperrors.BadRequestError("iam.errors.invalidTagSlug")
		}
		tagIDs = append(tagIDs, tags[0].ID)
	}
	return tagIDs, nil
}

func (h *BusinessProfileHandler) toResponse(profile *entity.BusinessProfile) dto.BusinessProfileResponse {
	var sectorRef *dto.SectorRef
	if profile.Sector != nil {
		sectorRef = &dto.SectorRef{
			ID:   profile.Sector.ID,
			Slug: profile.Sector.Slug,
		}
	}

	tagRefs := make([]dto.TagRef, 0, len(profile.Tags))
	for _, t := range profile.Tags {
		tagRefs = append(tagRefs, dto.TagRef{
			ID:            t.ID,
			Slug:          t.Slug,
			Group:         t.Group,
			IsMultiSelect: t.IsMultiSelect,
		})
	}

	return dto.BusinessProfileResponse{
		ID:                      profile.ID,
		CompanyName:             profile.CompanyName,
		CompanyEmail:            profile.CompanyEmail,
		CompanyPhoneNumber:      profile.CompanyPhoneNumber,
		PhysicalAddress:         profile.PhysicalAddress,
		Description:             profile.Description,
		LogoURL:                 profile.LogoURL,
		BannerURL:               profile.BannerURL,
		SocialLinks:             profile.SocialLinks,
		RegistrationNumber:      profile.RegistrationNumber,
		RegistrationDate:        profile.RegistrationDate,
		TaxIdentificationNumber: profile.TaxIdentificationNumber,
		TradeLicenseNumber:      profile.TradeLicenseNumber,
		Region:                  profile.Region,
		Stage:                   profile.Stage,
		Sector:                  sectorRef,
		Tags:                    tagRefs,
		CreatedAt:               profile.CreatedAt,
		UpdatedAt:               profile.UpdatedAt,
	}
}
