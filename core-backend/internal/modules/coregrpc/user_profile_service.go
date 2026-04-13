package coregrpc

import (
	"context"
	"errors"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	coreuserv1 "github.com/Final-Year-Project-G22/backend/core/pb/core/user/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	placeholderTier              = "pro"
	placeholderPreferredLanguage = "en"
)

type UserProfileService struct {
	coreuserv1.UnimplementedCoreUserServiceServer

	db *core.Database
}

func NewUserProfileService(db *core.Database) *UserProfileService {
	return &UserProfileService{db: db}
}

func (s *UserProfileService) GetUserProfile(
	ctx context.Context,
	req *coreuserv1.GetUserProfileRequest,
) (*coreuserv1.GetUserProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var account entity.Account
	err = s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to load user profile")
	}

	return &coreuserv1.GetUserProfileResponse{
		UserId:            userID.String(),
		AccountId:         account.ID.String(),
		Tier:              placeholderTier,
		PreferredLanguage: placeholderPreferredLanguage,
	}, nil
}
