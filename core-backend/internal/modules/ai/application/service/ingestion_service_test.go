package service

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
)

type fakeStorage struct {
	intent      *storage.UploadIntent
	intentErr   error
	exists      bool
	existsErr   error
	fileInfo    *storage.FileInfo
	fileInfoErr error
}

func (f *fakeStorage) CreateUploadIntent(context.Context, storage.UploadIntentOptions) (*storage.UploadIntent, error) {
	if f.intentErr != nil {
		return nil, f.intentErr
	}
	return f.intent, nil
}
func (f *fakeStorage) Upload(context.Context, storage.UploadOptions) (*storage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStorage) UploadFromReader(context.Context, storage.UploadOptions, io.Reader) (*storage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStorage) Download(context.Context, string) ([]byte, error) { return nil, nil }
func (f *fakeStorage) Delete(context.Context, string) error             { return nil }
func (f *fakeStorage) GetInfo(context.Context, string) (*storage.FileInfo, error) {
	if f.fileInfoErr != nil {
		return nil, f.fileInfoErr
	}
	return f.fileInfo, nil
}
func (f *fakeStorage) GetPresignedURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (f *fakeStorage) List(context.Context, storage.ListOptions) ([]storage.FileInfo, error) {
	return nil, nil
}
func (f *fakeStorage) Exists(context.Context, string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.exists, nil
}

type fakeDocRepo struct {
	existing *entity.IngestionDocument
	created  *entity.IngestionDocument
	getErr   error
}

func (f *fakeDocRepo) Create(_ context.Context, doc *entity.IngestionDocument) error {
	f.created = doc
	return nil
}
func (f *fakeDocRepo) GetByID(context.Context, uuid.UUID) (*entity.IngestionDocument, error) {
	return nil, nil
}
func (f *fakeDocRepo) GetByIdempotencyKey(context.Context, uuid.UUID, string) (*entity.IngestionDocument, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.existing, nil
}
func (f *fakeDocRepo) UpdateStatus(context.Context, uuid.UUID, entity.IngestionDocumentStatus, *string) error {
	return nil
}

type fakeOutboxRepo2 struct {
	created *entity.IngestionOutbox
}

func (f *fakeOutboxRepo2) Create(_ context.Context, item *entity.IngestionOutbox) error {
	f.created = item
	return nil
}
func (f *fakeOutboxRepo2) GetByEventID(context.Context, uuid.UUID) (*entity.IngestionOutbox, error) {
	return nil, nil
}
func (f *fakeOutboxRepo2) ListPending(context.Context, time.Time, int) ([]*entity.IngestionOutbox, error) {
	return nil, nil
}
func (f *fakeOutboxRepo2) MarkPublished(context.Context, uuid.UUID, time.Time, []byte) error {
	return nil
}
func (f *fakeOutboxRepo2) MarkRetryScheduled(context.Context, uuid.UUID, int, time.Time, int32, string) error {
	return nil
}
func (f *fakeOutboxRepo2) MarkDeadLetter(context.Context, uuid.UUID, int, int32, string) error {
	return nil
}

type fakeTransactor struct{}

func (f *fakeTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newEnabledService(fs *fakeStorage, docRepo *fakeDocRepo, outboxRepo *fakeOutboxRepo2) *IngestionService {
	return NewIngestionService(true, fs, docRepo, outboxRepo, &fakeTransactor{})
}

func newDisabledService(fs *fakeStorage, docRepo *fakeDocRepo, outboxRepo *fakeOutboxRepo2) *IngestionService {
	return NewIngestionService(false, fs, docRepo, outboxRepo, &fakeTransactor{})
}

func TestCreateUploadIntent_ValidatesContentType(t *testing.T) {
	svc := newEnabledService(&fakeStorage{}, &fakeDocRepo{}, &fakeOutboxRepo2{})
	_, err := svc.CreateUploadIntent(context.Background(), CreateUploadIntentInput{ContentType: ""})
	if err == nil {
		t.Fatalf("expected error for missing content type")
	}
}

func TestCreateUploadIntent_ReturnsStorageIntent(t *testing.T) {
	intent := &storage.UploadIntent{Key: "k", UploadURL: "u", Method: "PUT", ExpiresAt: time.Now().Add(10 * time.Minute)}
	svc := newEnabledService(&fakeStorage{intent: intent}, &fakeDocRepo{}, &fakeOutboxRepo2{})

	out, err := svc.CreateUploadIntent(context.Background(), CreateUploadIntentInput{ContentType: "application/pdf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Key != intent.Key || out.UploadURL != intent.UploadURL {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestCreateUploadIntent_RejectsWhenIngestionDisabled(t *testing.T) {
	svc := newDisabledService(&fakeStorage{}, &fakeDocRepo{}, &fakeOutboxRepo2{})
	_, err := svc.CreateUploadIntent(context.Background(), CreateUploadIntentInput{ContentType: "application/pdf"})
	if err == nil {
		t.Fatalf("expected forbidden error when ingestion is disabled")
	}
}

func TestFinalizeUpload_ReturnsExistingForDuplicateIdempotencyKey(t *testing.T) {
	id := uuid.New()
	eventID := uuid.New()
	docRepo := &fakeDocRepo{existing: &entity.IngestionDocument{BaseModel: entity.IngestionDocument{}.BaseModel, EventID: eventID, Status: entity.IngestionDocumentStatusQueued}}
	docRepo.existing.ID = id

	svc := newEnabledService(&fakeStorage{}, docRepo, &fakeOutboxRepo2{})

	out, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/a.pdf",
		ContentType:    "application/pdf",
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.IngestionID != id || out.EventID != eventID {
		t.Fatalf("expected existing IDs returned")
	}
}

func TestFinalizeUpload_PersistsDocumentAndOutbox(t *testing.T) {
	docRepo := &fakeDocRepo{}
	outboxRepo := &fakeOutboxRepo2{}
	svc := newEnabledService(
		&fakeStorage{
			exists: true,
			fileInfo: &storage.FileInfo{
				Size:        128,
				ContentType: "application/pdf",
			},
		},
		docRepo,
		outboxRepo,
	)

	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/a.pdf",
		ContentType:    "application/pdf",
		SizeBytes:      128,
		ChecksumSHA256: "abc",
		IdempotencyKey: "idem-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docRepo.created == nil {
		t.Fatalf("expected document to be created")
	}
	if outboxRepo.created == nil {
		t.Fatalf("expected outbox row to be created")
	}
}

func TestFinalizeUpload_FailsWhenObjectMissing(t *testing.T) {
	svc := newEnabledService(&fakeStorage{exists: false}, &fakeDocRepo{}, &fakeOutboxRepo2{})
	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/missing.pdf",
		ContentType:    "application/pdf",
		IdempotencyKey: "idem-missing",
	})
	if err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestFinalizeUpload_RequiresIdempotencyKey(t *testing.T) {
	svc := newEnabledService(&fakeStorage{exists: true}, &fakeDocRepo{}, &fakeOutboxRepo2{})
	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/file.pdf",
		ContentType:    "application/pdf",
		IdempotencyKey: "",
	})
	if err == nil {
		t.Fatalf("expected idempotency key validation error")
	}
}

func TestFinalizeUpload_PropagatesRepositoryLookupError(t *testing.T) {
	docRepo := &fakeDocRepo{getErr: errors.New("db down")}
	svc := newEnabledService(&fakeStorage{exists: true}, docRepo, &fakeOutboxRepo2{})
	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/file.pdf",
		ContentType:    "application/pdf",
		IdempotencyKey: "idem-x",
	})
	if err == nil {
		t.Fatalf("expected repository lookup error")
	}
}

func TestCreateUploadIntent_PropagatesStorageFailure(t *testing.T) {
	svc := newEnabledService(&fakeStorage{intentErr: errors.New("boom")}, &fakeDocRepo{}, &fakeOutboxRepo2{})
	_, err := svc.CreateUploadIntent(context.Background(), CreateUploadIntentInput{ContentType: "application/pdf"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestFinalizeUpload_FailsWhenSizeMismatch(t *testing.T) {
	svc := newEnabledService(
		&fakeStorage{
			exists: true,
			fileInfo: &storage.FileInfo{
				Size:        256,
				ContentType: "application/pdf",
			},
		},
		&fakeDocRepo{},
		&fakeOutboxRepo2{},
	)
	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/size.pdf",
		ContentType:    "application/pdf",
		SizeBytes:      128,
		IdempotencyKey: "idem-size",
	})
	if err == nil {
		t.Fatalf("expected size mismatch error")
	}
}

func TestFinalizeUpload_FailsWhenContentTypeMismatch(t *testing.T) {
	svc := newEnabledService(
		&fakeStorage{
			exists: true,
			fileInfo: &storage.FileInfo{
				Size:        128,
				ContentType: "image/png",
			},
		},
		&fakeDocRepo{},
		&fakeOutboxRepo2{},
	)
	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/type.pdf",
		ContentType:    "application/pdf",
		SizeBytes:      128,
		IdempotencyKey: "idem-type",
	})
	if err == nil {
		t.Fatalf("expected content type mismatch error")
	}
}

func TestFinalizeUpload_RejectsWhenIngestionDisabled(t *testing.T) {
	svc := newDisabledService(&fakeStorage{exists: true}, &fakeDocRepo{}, &fakeOutboxRepo2{})
	_, err := svc.FinalizeUpload(context.Background(), FinalizeUploadInput{
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		StorageKey:     "docs/file.pdf",
		ContentType:    "application/pdf",
		IdempotencyKey: "idem-disabled",
	})
	if err == nil {
		t.Fatalf("expected forbidden error when ingestion is disabled")
	}
}
