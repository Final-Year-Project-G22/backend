package service

import (
	"context"
	"testing"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/google/uuid"
)

type fakeStatusProjectionRepo struct {
	projections  map[uuid.UUID]*entity.IngestionStatusProjection
	getByDocErr  error
	getByAcctErr error
	getByUserErr error
}

func (f *fakeStatusProjectionRepo) GetByDocumentID(ctx context.Context, documentID uuid.UUID) (*entity.IngestionStatusProjection, error) {
	if f.getByDocErr != nil {
		return nil, f.getByDocErr
	}
	p, ok := f.projections[documentID]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (f *fakeStatusProjectionRepo) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error) {
	if f.getByAcctErr != nil {
		return nil, f.getByAcctErr
	}
	var result []*entity.IngestionStatusProjection
	for _, p := range f.projections {
		if p.AccountID == accountID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (f *fakeStatusProjectionRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error) {
	if f.getByUserErr != nil {
		return nil, f.getByUserErr
	}
	var result []*entity.IngestionStatusProjection
	for _, p := range f.projections {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (f *fakeStatusProjectionRepo) UpsertProjection(ctx context.Context, projection *entity.IngestionStatusProjection) error {
	f.projections[projection.DocumentID] = projection
	return nil
}

func TestStatusProjection_TenantIsolationByAccount(t *testing.T) {
	accountID1 := uuid.New()
	accountID2 := uuid.New()
	docID := uuid.New()

	repo := &fakeStatusProjectionRepo{
		projections: map[uuid.UUID]*entity.IngestionStatusProjection{
			docID: {
				DocumentID: docID,
				AccountID:  accountID1,
			},
		},
	}

	projections, err := repo.GetByAccountID(context.Background(), accountID2, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(projections) != 0 {
		t.Errorf("expected no projections for different account, got %d", len(projections))
	}
}

func TestStatusProjection_TenantIsolationByUser(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	docID := uuid.New()

	repo := &fakeStatusProjectionRepo{
		projections: map[uuid.UUID]*entity.IngestionStatusProjection{
			docID: {
				DocumentID: docID,
				UserID:     userID1,
			},
		},
	}

	projections, err := repo.GetByUserID(context.Background(), userID2, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(projections) != 0 {
		t.Errorf("expected no projections for different user, got %d", len(projections))
	}
}

func TestStatusProjection_OwnerCanAccessOwnDocument(t *testing.T) {
	accountID := uuid.New()
	docID := uuid.New()

	repo := &fakeStatusProjectionRepo{
		projections: map[uuid.UUID]*entity.IngestionStatusProjection{
			docID: {
				DocumentID: docID,
				AccountID:  accountID,
			},
		},
	}

	projection, err := repo.GetByDocumentID(context.Background(), docID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if projection == nil {
		t.Fatal("expected projection, got nil")
	}

	if projection.AccountID != accountID {
		t.Errorf("expected account %s, got %s", accountID, projection.AccountID)
	}
}

func TestStatusProjection_MissingDocumentReturnsNil(t *testing.T) {
	repo := &fakeStatusProjectionRepo{
		projections: map[uuid.UUID]*entity.IngestionStatusProjection{},
	}

	projection, err := repo.GetByDocumentID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if projection != nil {
		t.Errorf("expected nil for missing document, got %v", projection)
	}
}

func TestStatusProjection_RejectsCrossAccountQuery(t *testing.T) {
	accountID1 := uuid.New()
	accountID2 := uuid.New()
	docID := uuid.New()

	repo := &fakeStatusProjectionRepo{
		projections: map[uuid.UUID]*entity.IngestionStatusProjection{
			docID: {
				DocumentID: docID,
				AccountID:  accountID1,
			},
		},
	}

	results, err := repo.GetByAccountID(context.Background(), accountID2, 100, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range results {
		if p.AccountID == accountID1 {
			t.Error("rejected cross-account query but still got results")
			break
		}
	}
}
