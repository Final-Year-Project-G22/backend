package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type BusinessProfileUsecase interface {
	CreateBusinessProfile(ctx context.Context, accountID uuid.UUID, input CreateBusinessProfileInput) (*entity.BusinessProfile, error)
	GetBusinessProfileByAccount(ctx context.Context, accountID uuid.UUID) (*entity.BusinessProfile, error)
	UpdateBusinessProfile(ctx context.Context, accountID uuid.UUID, input UpdateBusinessProfileInput) (*entity.BusinessProfile, error)
	UpdateSocialLinks(ctx context.Context, accountID uuid.UUID, socialLinks datatypes.JSONMap) (*entity.BusinessProfile, error)
}

type CreateBusinessProfileInput struct {
	CompanyName             string
	CompanyEmail            string
	CompanyPhoneNumber      string
	RegistrationNumber      *string
	RegistrationDate        *time.Time
	TaxIdentificationNumber *string
	TradeLicenseNumber      *string
	PhysicalAddress         *string
	Description             *string
	LogoURL                 *string
	BannerURL               *string
	SocialLinks             datatypes.JSONMap
	SectorID                *uuid.UUID
	TagIDs                  []uuid.UUID
	Region                  *entity.Region
	Stage                   *entity.BusinessStage
}

type UpdateBusinessProfileInput struct {
	CompanyName             *string
	CompanyEmail            *string
	CompanyPhoneNumber      *string
	RegistrationNumber      *string
	RegistrationDate        *time.Time
	TaxIdentificationNumber *string
	TradeLicenseNumber      *string
	PhysicalAddress         *string
	Description             *string
	LogoURL                 *string
	BannerURL               *string
	SocialLinks             datatypes.JSONMap
	SectorID                *uuid.UUID
	TagIDs                  []uuid.UUID
	Region                  *entity.Region
	Stage                   *entity.BusinessStage
}
