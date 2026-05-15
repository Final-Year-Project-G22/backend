package service

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// mockAttachmentRepo implements communityrepo.AttachmentRepository
type mockAttachmentRepo struct {
	createFunc       func(ctx context.Context, attachment *entity.Attachment) error
	findByIDsFunc    func(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error)
	updatePostIDFunc func(ctx context.Context, attachmentIDs []uuid.UUID, postID uuid.UUID) error
	deleteByIDsFunc  func(ctx context.Context, ids []uuid.UUID) error
	findByPostIDFunc func(ctx context.Context, postID uuid.UUID) ([]*entity.Attachment, error)
	findPendingFunc  func(ctx context.Context, olderThan time.Time) ([]*entity.Attachment, error)
	bulkCreateFunc   func(ctx context.Context, entities []*entity.Attachment) error
}

func (m *mockAttachmentRepo) Create(ctx context.Context, attachment *entity.Attachment) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, attachment)
	}
	return nil
}

func (m *mockAttachmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Attachment, error) {
	return nil, nil
}

func (m *mockAttachmentRepo) Update(ctx context.Context, entity *entity.Attachment) error {
	return nil
}

func (m *mockAttachmentRepo) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

func (m *mockAttachmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAttachmentRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAttachmentRepo) FindAll(ctx context.Context, opts query.QueryOptions) repository.PaginatedResult[entity.Attachment] {
	return repository.PaginatedResult[entity.Attachment]{}
}

func (m *mockAttachmentRepo) FindAllArchived(ctx context.Context, opts query.QueryOptions) repository.PaginatedResult[entity.Attachment] {
	return repository.PaginatedResult[entity.Attachment]{}
}

func (m *mockAttachmentRepo) First(ctx context.Context, opts query.QueryOptions) (*entity.Attachment, error) {
	return nil, nil
}

func (m *mockAttachmentRepo) Find(ctx context.Context, opts query.QueryOptions) ([]*entity.Attachment, error) {
	return nil, nil
}

func (m *mockAttachmentRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
	if m.findByIDsFunc != nil {
		return m.findByIDsFunc(ctx, ids)
	}
	return nil, nil
}

func (m *mockAttachmentRepo) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockAttachmentRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockAttachmentRepo) Transaction(ctx context.Context, fn func(repo repository.GenericRepository[entity.Attachment]) error) error {
	return nil
}

func (m *mockAttachmentRepo) GetDB() *gorm.DB {
	return nil
}

func (m *mockAttachmentRepo) UpdatePostID(ctx context.Context, attachmentIDs []uuid.UUID, postID uuid.UUID) error {
	if m.updatePostIDFunc != nil {
		return m.updatePostIDFunc(ctx, attachmentIDs, postID)
	}
	return nil
}

func (m *mockAttachmentRepo) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if m.deleteByIDsFunc != nil {
		return m.deleteByIDsFunc(ctx, ids)
	}
	return nil
}

func (m *mockAttachmentRepo) FindByPostID(ctx context.Context, postID uuid.UUID) ([]*entity.Attachment, error) {
	if m.findByPostIDFunc != nil {
		return m.findByPostIDFunc(ctx, postID)
	}
	return nil, nil
}

func (m *mockAttachmentRepo) FindPendingOlderThan(ctx context.Context, olderThan time.Time) ([]*entity.Attachment, error) {
	if m.findPendingFunc != nil {
		return m.findPendingFunc(ctx, olderThan)
	}
	return nil, nil
}

func (m *mockAttachmentRepo) BulkCreate(ctx context.Context, entities []*entity.Attachment) error {
	if m.bulkCreateFunc != nil {
		return m.bulkCreateFunc(ctx, entities)
	}
	return nil
}

// mockStorage implements storage.Storage
type mockStorage struct {
	uploadFunc func(ctx context.Context, opts storage.UploadOptions) (*storage.FileInfo, error)
	deleteFunc func(ctx context.Context, key string) error
}

func (m *mockStorage) Upload(ctx context.Context, opts storage.UploadOptions) (*storage.FileInfo, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, opts)
	}
	return &storage.FileInfo{URL: "/api/v1/files/" + opts.Key}, nil
}

func (m *mockStorage) UploadFromReader(ctx context.Context, opts storage.UploadOptions, reader io.Reader) (*storage.FileInfo, error) {
	return nil, nil
}

func (m *mockStorage) Download(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return nil
}

func (m *mockStorage) GetInfo(ctx context.Context, key string) (*storage.FileInfo, error) {
	return nil, nil
}

func (m *mockStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", nil
}
func (m *mockStorage) GetPresignedURLLocal(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", nil
}

func (m *mockStorage) List(ctx context.Context, opts storage.ListOptions) ([]storage.FileInfo, error) {
	return nil, nil
}

func (m *mockStorage) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *mockStorage) CreateUploadIntent(ctx context.Context, opts storage.UploadIntentOptions) (*storage.UploadIntent, error) {
	return nil, nil
}

// mockValidator implements AttachmentValidator
type mockValidator struct {
	validateFunc func(fileBytes []byte) (*ValidatedCommunityAttachment, error)
}

func (m *mockValidator) Validate(fileBytes []byte) (*ValidatedCommunityAttachment, error) {
	if m.validateFunc != nil {
		return m.validateFunc(fileBytes)
	}
	return &ValidatedCommunityAttachment{ContentType: "image/jpeg", Extension: ".jpg"}, nil
}

func TestAttachmentService_Upload(t *testing.T) {
	repo := &mockAttachmentRepo{
		createFunc: func(ctx context.Context, attachment *entity.Attachment) error {
			// Simulate successful creation
			return nil
		},
	}
	stor := &mockStorage{
		uploadFunc: func(ctx context.Context, opts storage.UploadOptions) (*storage.FileInfo, error) {
			return &storage.FileInfo{URL: "/api/v1/files/test-key"}, nil
		},
	}
	validator := &mockValidator{
		validateFunc: func(fileBytes []byte) (*ValidatedCommunityAttachment, error) {
			return &ValidatedCommunityAttachment{ContentType: "image/jpeg", Extension: ".jpg", Content: fileBytes}, nil
		},
	}

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        stor,
		validator:      validator,
	}

	accountID := uuid.New()
	inputs := []usecase.AttachmentUploadInput{
		{FileBytes: []byte("fake-image"), Filename: "test.jpg"},
	}

	attachments, err := svc.Upload(context.Background(), accountID, inputs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(attachments))
	assert.Equal(t, "image/jpeg", attachments[0].FileType)
	assert.Equal(t, accountID, attachments[0].UploadedBy)
	assert.Equal(t, entity.AttachmentStatusPending, attachments[0].Status)
}

func TestAttachmentService_Upload_ValidationFails(t *testing.T) {
	repo := &mockAttachmentRepo{}
	stor := &mockStorage{}
	validator := NewCommunityAttachmentValidator()

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        stor,
		validator:      validator,
	}

	accountID := uuid.New()
	inputs := []usecase.AttachmentUploadInput{
		{FileBytes: []byte{}, Filename: "test.jpg"},
	}

	attachments, err := svc.Upload(context.Background(), accountID, inputs)
	assert.Error(t, err)
	assert.Nil(t, attachments)
}

func TestAttachmentService_LinkToPost(t *testing.T) {
	attachmentID := uuid.New()
	postID := uuid.New()
	accountID := uuid.New()

	repo := &mockAttachmentRepo{
		findByIDsFunc: func(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
			return []*entity.Attachment{
				{BaseModel: model.BaseModel{ID: attachmentID}, Status: entity.AttachmentStatusPending, UploadedBy: accountID},
			}, nil
		},
		updatePostIDFunc: func(ctx context.Context, attachmentIDs []uuid.UUID, pID uuid.UUID) error {
			assert.Equal(t, postID, pID)
			return nil
		},
	}

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        &mockStorage{},
		validator:      NewCommunityAttachmentValidator(),
	}

	err := svc.LinkToPost(context.Background(), postID, []uuid.UUID{attachmentID}, accountID)
	assert.NoError(t, err)
}

func TestAttachmentService_LinkToPost_NotOwned(t *testing.T) {
	attachmentID := uuid.New()
	postID := uuid.New()
	accountID := uuid.New()
	otherAccountID := uuid.New()

	repo := &mockAttachmentRepo{
		findByIDsFunc: func(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
			return []*entity.Attachment{
				{BaseModel: model.BaseModel{ID: attachmentID}, Status: entity.AttachmentStatusPending, UploadedBy: otherAccountID},
			}, nil
		},
	}

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        &mockStorage{},
		validator:      NewCommunityAttachmentValidator(),
	}

	err := svc.LinkToPost(context.Background(), postID, []uuid.UUID{attachmentID}, accountID)
	assert.Error(t, err)
}

func TestAttachmentService_UnlinkFromPost(t *testing.T) {
	attachmentID := uuid.New()
	postID := uuid.New()
	accountID := uuid.New()

	repo := &mockAttachmentRepo{
		findByIDsFunc: func(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
			pid := postID
			return []*entity.Attachment{
				{BaseModel: model.BaseModel{ID: attachmentID}, PostID: &pid, UploadedBy: accountID, StorageKey: "test-key"},
			}, nil
		},
		deleteByIDsFunc: func(ctx context.Context, ids []uuid.UUID) error {
			assert.Equal(t, attachmentID, ids[0])
			return nil
		},
	}

	stor := &mockStorage{
		deleteFunc: func(ctx context.Context, key string) error {
			assert.Equal(t, "test-key", key)
			return nil
		},
	}

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        stor,
		validator:      NewCommunityAttachmentValidator(),
	}

	err := svc.UnlinkFromPost(context.Background(), postID, []uuid.UUID{attachmentID}, accountID)
	assert.NoError(t, err)
}

func TestAttachmentService_DeleteOrphan(t *testing.T) {
	attachmentID := uuid.New()
	accountID := uuid.New()

	repo := &mockAttachmentRepo{
		findByIDsFunc: func(ctx context.Context, ids []uuid.UUID) ([]*entity.Attachment, error) {
			return []*entity.Attachment{
				{BaseModel: model.BaseModel{ID: attachmentID}, Status: entity.AttachmentStatusPending, UploadedBy: accountID, StorageKey: "test-key"},
			}, nil
		},
		deleteByIDsFunc: func(ctx context.Context, ids []uuid.UUID) error {
			return nil
		},
	}

	stor := &mockStorage{
		deleteFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        stor,
		validator:      NewCommunityAttachmentValidator(),
	}

	err := svc.DeleteOrphan(context.Background(), attachmentID, accountID)
	assert.NoError(t, err)
}

func TestAttachmentService_CleanupPending(t *testing.T) {
	attachmentID := uuid.New()

	repo := &mockAttachmentRepo{
		findPendingFunc: func(ctx context.Context, olderThan time.Time) ([]*entity.Attachment, error) {
			return []*entity.Attachment{
				{BaseModel: model.BaseModel{ID: attachmentID}, StorageKey: "test-key"},
			}, nil
		},
		deleteByIDsFunc: func(ctx context.Context, ids []uuid.UUID) error {
			return nil
		},
	}

	stor := &mockStorage{
		deleteFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}

	svc := &attachmentService{
		attachmentRepo: repo,
		storage:        stor,
		validator:      NewCommunityAttachmentValidator(),
	}

	err := svc.CleanupPending(context.Background(), time.Now().UTC())
	assert.NoError(t, err)
}
