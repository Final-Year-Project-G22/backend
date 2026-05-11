package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SectorRef is a lightweight sector reference in business profile responses.
type SectorRef struct {
	ID   uuid.UUID `json:"id" doc:"Sector identifier"`
	Slug string    `json:"slug" doc:"Sector slug"`
}

// TagRef is a lightweight tag reference in business profile responses.
type TagRef struct {
	ID            uuid.UUID       `json:"id" doc:"Tag identifier"`
	Slug          string          `json:"slug" doc:"Tag slug"`
	Group         entity.TagGroup `json:"group" doc:"Tag group"`
	IsMultiSelect bool            `json:"isMultiSelect" doc:"Whether the tag allows multiple selections"`
}

// BusinessProfileResponse is the response body for business profile operations.
type BusinessProfileResponse struct {
	ID                      uuid.UUID             `json:"id" doc:"Business profile identifier"`
	CompanyName             string                `json:"companyName" doc:"Company name"`
	CompanyEmail            string                `json:"companyEmail" doc:"Company email"`
	CompanyPhoneNumber      string                `json:"companyPhoneNumber" doc:"Company phone number"`
	PhysicalAddress         *string               `json:"physicalAddress,omitempty" doc:"Physical address"`
	Description             *string               `json:"description,omitempty" doc:"Business description"`
	LogoURL                 *string               `json:"logoUrl,omitempty" doc:"Logo URL"`
	BannerURL               *string               `json:"bannerUrl,omitempty" doc:"Banner URL"`
	SocialLinks             datatypes.JSONMap     `json:"socialLinks" doc:"Social media links"`
	RegistrationNumber      *string               `json:"registrationNumber,omitempty" doc:"Business registration number"`
	RegistrationDate        *time.Time            `json:"registrationDate,omitempty" doc:"Registration date"`
	TaxIdentificationNumber *string               `json:"taxIdentificationNumber,omitempty" doc:"TIN"`
	TradeLicenseNumber      *string               `json:"tradeLicenseNumber,omitempty" doc:"Trade license number"`
	Region                  *entity.Region        `json:"region,omitempty" doc:"Business region"`
	Stage                   *entity.BusinessStage `json:"stage,omitempty" doc:"Business lifecycle stage"`
	Sector                  *SectorRef            `json:"sector,omitempty" doc:"Business sector"`
	Tags                    []TagRef              `json:"tags" doc:"Associated tags"`
	CreatedAt               *time.Time            `json:"createdAt" doc:"Creation timestamp"`
	UpdatedAt               *time.Time            `json:"updatedAt" doc:"Last update timestamp"`
}

// CreateBusinessProfileRequest is the input for creating a business profile.
// Slugs are accepted for sector and tags; they are resolved server-side.
type CreateBusinessProfileRequest struct {
	CompanyName             string                `json:"companyName,omitempty" doc:"Company name (auto-filled if omitted)"`
	CompanyEmail            string                `json:"companyEmail,omitempty" doc:"Company email (auto-filled if omitted)"`
	CompanyPhoneNumber      string                `json:"companyPhoneNumber,omitempty" doc:"Company phone (auto-filled if omitted)"`
	PhysicalAddress         *string               `json:"physicalAddress,omitempty" doc:"Physical address"`
	Description             *string               `json:"description,omitempty" doc:"Business description"`
	LogoURL                 *string               `json:"logoUrl,omitempty" doc:"Logo URL"`
	BannerURL               *string               `json:"bannerUrl,omitempty" doc:"Banner URL"`
	SocialLinks             datatypes.JSONMap     `json:"socialLinks,omitempty" doc:"Social media links"`
	RegistrationNumber      *string               `json:"registrationNumber,omitempty" doc:"Business registration number"`
	RegistrationDate        *time.Time            `json:"registrationDate,omitempty" doc:"Registration date"`
	TaxIdentificationNumber *string               `json:"taxIdentificationNumber,omitempty" doc:"TIN"`
	TradeLicenseNumber      *string               `json:"tradeLicenseNumber,omitempty" doc:"Trade license number"`
	Region                  *entity.Region        `json:"region,omitempty" doc:"Business region"`
	Stage                   *entity.BusinessStage `json:"stage,omitempty" doc:"Business lifecycle stage"`
	SectorSlug              *string               `json:"sectorSlug,omitempty" doc:"Sector slug (e.g., trade, manufacturing)"`
	TagSlugs                []string              `json:"tagSlugs,omitempty" doc:"Tag slugs (e.g., sole-proprietor, tax-vat)"`
}

// CreateBusinessProfileInput wraps the create request body.
type CreateBusinessProfileInput struct {
	Body CreateBusinessProfileRequest
}

// CreateBusinessProfileOutput is the response for creating a business profile.
type CreateBusinessProfileOutput struct {
	Body BusinessProfileResponse
}

// UpdateBusinessProfileRequest is the input for updating a business profile.
type UpdateBusinessProfileRequest struct {
	CompanyName             *string               `json:"companyName,omitempty" doc:"Company name"`
	CompanyEmail            *string               `json:"companyEmail,omitempty" doc:"Company email"`
	CompanyPhoneNumber      *string               `json:"companyPhoneNumber,omitempty" doc:"Company phone"`
	PhysicalAddress         *string               `json:"physicalAddress,omitempty" doc:"Physical address"`
	Description             *string               `json:"description,omitempty" doc:"Business description"`
	LogoURL                 *string               `json:"logoUrl,omitempty" doc:"Logo URL"`
	BannerURL               *string               `json:"bannerUrl,omitempty" doc:"Banner URL"`
	SocialLinks             datatypes.JSONMap     `json:"socialLinks,omitempty" doc:"Social media links"`
	RegistrationNumber      *string               `json:"registrationNumber,omitempty" doc:"Business registration number"`
	RegistrationDate        *time.Time            `json:"registrationDate,omitempty" doc:"Registration date"`
	TaxIdentificationNumber *string               `json:"taxIdentificationNumber,omitempty" doc:"TIN"`
	TradeLicenseNumber      *string               `json:"tradeLicenseNumber,omitempty" doc:"Trade license number"`
	Region                  *entity.Region        `json:"region,omitempty" doc:"Business region"`
	Stage                   *entity.BusinessStage `json:"stage,omitempty" doc:"Business lifecycle stage"`
	SectorSlug              *string               `json:"sectorSlug,omitempty" doc:"Sector slug"`
	TagSlugs                []string              `json:"tagSlugs,omitempty" doc:"Tag slugs"`
}

// UpdateBusinessProfileInput wraps the update request body.
type UpdateBusinessProfileInput struct {
	Body UpdateBusinessProfileRequest
}

// UpdateBusinessProfileOutput is the response for updating a business profile.
type UpdateBusinessProfileOutput struct {
	Body BusinessProfileResponse
}

// GetBusinessProfileInput has no body — account is taken from middleware context.
type GetBusinessProfileInput struct{}

// GetBusinessProfileOutput is the response for fetching a business profile.
type GetBusinessProfileOutput struct {
	Body BusinessProfileResponse
}
