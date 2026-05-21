package appusecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifevent "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type businessProfileUsecase struct {
	profileRepo     repository.BusinessProfileRepository
	tagRepo         repository.TagRepository
	sectorRepo      repository.SectorRepository
	notifOutboxRepo notifrepo.NotificationOutboxRepository
}

func NewBusinessProfileUsecase(
	profileRepo repository.BusinessProfileRepository,
	tagRepo repository.TagRepository,
	sectorRepo repository.SectorRepository,
	notifOutboxRepo notifrepo.NotificationOutboxRepository,
) iamusecase.BusinessProfileUsecase {
	return &businessProfileUsecase{
		profileRepo:     profileRepo,
		tagRepo:         tagRepo,
		sectorRepo:      sectorRepo,
		notifOutboxRepo: notifOutboxRepo,
	}
}

func (u *businessProfileUsecase) CreateBusinessProfile(ctx context.Context, accountID uuid.UUID, input iamusecase.CreateBusinessProfileInput) (*entity.BusinessProfile, error) {
	exists, err := u.profileRepo.ExistsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, iamerror.ErrBusinessProfileAlreadyExists
	}

	if err := u.validateTags(ctx, input.TagIDs); err != nil {
		return nil, err
	}

	profile := &entity.BusinessProfile{
		AccountID:               accountID,
		CompanyName:             input.CompanyName,
		CompanyEmail:            input.CompanyEmail,
		CompanyPhoneNumber:      input.CompanyPhoneNumber,
		RegistrationNumber:      input.RegistrationNumber,
		RegistrationDate:        input.RegistrationDate,
		TaxIdentificationNumber: input.TaxIdentificationNumber,
		TradeLicenseNumber:      input.TradeLicenseNumber,
		PhysicalAddress:         input.PhysicalAddress,
		Description:             input.Description,
		LogoURL:                 input.LogoURL,
		BannerURL:               input.BannerURL,
		SocialLinks:             input.SocialLinks,
		SectorID:                input.SectorID,
		Region:                  input.Region,
		Stage:                   input.Stage,
	}

	if len(input.TagIDs) > 0 {
		tagPtrs, err := u.tagRepo.FindByIDs(ctx, input.TagIDs)
		if err != nil {
			return nil, err
		}
		profile.Tags = make([]entity.Tag, len(tagPtrs))
		for i, t := range tagPtrs {
			profile.Tags[i] = *t
		}
	}

	if err := u.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	if err := u.publishBPEvent(ctx, accountID); err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *businessProfileUsecase) GetBusinessProfileByAccount(ctx context.Context, accountID uuid.UUID) (*entity.BusinessProfile, error) {
	return u.profileRepo.GetByAccountID(ctx, accountID)
}

func (u *businessProfileUsecase) UpdateBusinessProfile(ctx context.Context, accountID uuid.UUID, input iamusecase.UpdateBusinessProfileInput) (*entity.BusinessProfile, error) {
	profile, err := u.profileRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if err := u.validateTags(ctx, input.TagIDs); err != nil {
		return nil, err
	}

	if input.CompanyName != nil {
		profile.CompanyName = *input.CompanyName
	}
	if input.CompanyEmail != nil {
		profile.CompanyEmail = *input.CompanyEmail
	}
	if input.CompanyPhoneNumber != nil {
		profile.CompanyPhoneNumber = *input.CompanyPhoneNumber
	}
	if input.RegistrationNumber != nil {
		profile.RegistrationNumber = input.RegistrationNumber
	}
	if input.RegistrationDate != nil {
		profile.RegistrationDate = input.RegistrationDate
	}
	if input.TaxIdentificationNumber != nil {
		profile.TaxIdentificationNumber = input.TaxIdentificationNumber
	}
	if input.TradeLicenseNumber != nil {
		profile.TradeLicenseNumber = input.TradeLicenseNumber
	}
	if input.PhysicalAddress != nil {
		profile.PhysicalAddress = input.PhysicalAddress
	}
	if input.Description != nil {
		profile.Description = input.Description
	}
	if input.LogoURL != nil {
		profile.LogoURL = input.LogoURL
	}
	if input.BannerURL != nil {
		profile.BannerURL = input.BannerURL
	}
	if input.SocialLinks != nil {
		profile.SocialLinks = input.SocialLinks
	}
	if input.SectorID != nil {
		profile.SectorID = input.SectorID
	}
	if input.Region != nil {
		profile.Region = input.Region
	}
	if input.Stage != nil {
		profile.Stage = input.Stage
	}

	if input.TagIDs != nil {
		if len(input.TagIDs) > 0 {
			tagPtrs, err := u.tagRepo.FindByIDs(ctx, input.TagIDs)
			if err != nil {
				return nil, err
			}
			profile.Tags = make([]entity.Tag, len(tagPtrs))
			for i, t := range tagPtrs {
				profile.Tags[i] = *t
			}
		} else {
			profile.Tags = nil
		}
	}

	if err := u.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	if err := u.publishBPEvent(ctx, accountID); err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *businessProfileUsecase) UpdateSocialLinks(ctx context.Context, accountID uuid.UUID, socialLinks datatypes.JSONMap) (*entity.BusinessProfile, error) {
	profile, err := u.profileRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	profile.SocialLinks = socialLinks
	if err := u.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *businessProfileUsecase) publishBPEvent(ctx context.Context, accountID uuid.UUID) error {
	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        notifevent.BusinessProfileUpdated,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "iam",
		AccountID:        accountID,
		NotificationType: "",
		ChannelPolicy:    notificationevent.ChannelPolicySingle,
		Channel:          strPtr("in_app"),
		Variables:        map[string]string{},
		Metadata: notificationevent.Metadata{
			IdempotencyKey: "bp-updated:" + accountID.String() + ":" + uuid.New().String(),
		},
	}
	return u.writeToOutbox(ctx, &env)
}

func (u *businessProfileUsecase) writeToOutbox(ctx context.Context, env *notificationevent.Envelope) error {
	payload := make(map[string]interface{})
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	outbox := &notifentity.NotificationOutbox{
		EventType:      env.EventType,
		SchemaVersion:  env.SchemaVersion,
		SourceModule:   env.SourceModule,
		AccountID:      env.AccountID,
		IdempotencyKey: env.Metadata.IdempotencyKey,
		Payload:        payload,
		Status:         notifentity.NotificationOutboxStatusPending,
	}
	return u.notifOutboxRepo.Create(ctx, outbox)
}

func strPtr(s string) *string {
	return &s
}

// validateTags enforces tag group rules:
// - For single-select groups (IsMultiSelect=false), only one tag per group is allowed.
// - For multi-select groups, any number of tags is allowed.
func (u *businessProfileUsecase) validateTags(ctx context.Context, tagIDs []uuid.UUID) error {
	if len(tagIDs) == 0 {
		return nil
	}

	tagPtrs, err := u.tagRepo.FindByIDs(ctx, tagIDs)
	if err != nil {
		return err
	}

	if len(tagPtrs) != len(tagIDs) {
		return errors.BadRequestError("errors.invalidInput")
	}

	groupCount := make(map[entity.TagGroup]int)
	for _, tag := range tagPtrs {
		groupCount[tag.Group]++
	}

	for _, tag := range tagPtrs {
		if !tag.IsMultiSelect && groupCount[tag.Group] > 1 {
			return errors.BadRequestError("errors.invalidInput")
		}
	}

	return nil
}
