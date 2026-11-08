package usecase_test

import (
	"context"
	"testing"

	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	"github.com/VH288/tc-template/internal/domain/sample/usecase"
	"github.com/VH288/tc-template/internal/infrastructure/telemetry"
	apperrors "github.com/VH288/tc-template/internal/pkg/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type memRepo struct {
	items map[string]*entity.Sample
}

func newMemRepo() *memRepo {
	return &memRepo{items: make(map[string]*entity.Sample)}
}

var _ sample.Repository = (*memRepo)(nil)

func (m *memRepo) Create(ctx context.Context, sample *entity.Sample) error {
	if sample.ID == "" {
		sample.ID = uuid.NewString()
	}
	if sample.Version == 0 {
		sample.Version = 1
	}
	cp := *sample
	m.items[sample.ID] = &cp
	return nil
}

func (m *memRepo) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *memRepo) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	out := make([]entity.Sample, 0, len(m.items))
	for _, s := range m.items {
		out = append(out, *s)
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) Update(ctx context.Context, sample *entity.Sample) error {
	existing, ok := m.items[sample.ID]
	if !ok {
		return apperrors.ErrNotFound
	}
	if existing.Version != sample.Version {
		return apperrors.ErrConflict
	}
	sample.Version++
	cp := *sample
	m.items[sample.ID] = &cp
	return nil
}

func (m *memRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func TestSampleUsecase_Create(t *testing.T) {
	repo := newMemRepo()
	uc := usecase.New(repo, telemetry.NoopLogger{})
	ctx := context.Background()

	// Default status
	s1, err := uc.Create(ctx, "Sample 1", "Desc", "")
	require.NoError(t, err)
	require.Equal(t, "Sample 1", s1.Name)
	require.Equal(t, entity.StatusActive, s1.Status)

	// Explicit status
	s2, err := uc.Create(ctx, "Sample 2", "Desc", entity.StatusInactive)
	require.NoError(t, err)
	require.Equal(t, entity.StatusInactive, s2.Status)
}

func TestSampleUsecase_GetByID(t *testing.T) {
	repo := newMemRepo()
	uc := usecase.New(repo, telemetry.NoopLogger{})
	ctx := context.Background()

	created, err := uc.Create(ctx, "Sample", "Desc", "")
	require.NoError(t, err)

	got, err := uc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	_, err = uc.GetByID(ctx, "unknown")
	require.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestSampleUsecase_List(t *testing.T) {
	repo := newMemRepo()
	uc := usecase.New(repo, telemetry.NoopLogger{})
	ctx := context.Background()

	_, _ = uc.Create(ctx, "Sample 1", "Desc", "")
	_, _ = uc.Create(ctx, "Sample 2", "Desc", "")

	items, total, err := uc.List(ctx, 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)
}

func TestSampleUsecase_Update(t *testing.T) {
	repo := newMemRepo()
	uc := usecase.New(repo, telemetry.NoopLogger{})
	ctx := context.Background()

	created, err := uc.Create(ctx, "Old Name", "Old Desc", entity.StatusActive)
	require.NoError(t, err)

	updated, err := uc.Update(ctx, created.ID, "New Name", "New Desc", entity.StatusInactive)
	require.NoError(t, err)
	require.Equal(t, "New Name", updated.Name)
	require.Equal(t, "New Desc", updated.Description)
	require.Equal(t, entity.StatusInactive, updated.Status)

	// Update non-existent
	_, err = uc.Update(ctx, "non-existent", "N", "D", entity.StatusActive)
	require.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestSampleUsecase_Delete(t *testing.T) {
	repo := newMemRepo()
	uc := usecase.New(repo, telemetry.NoopLogger{})
	ctx := context.Background()

	created, err := uc.Create(ctx, "To Delete", "Desc", "")
	require.NoError(t, err)

	err = uc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = uc.GetByID(ctx, created.ID)
	require.ErrorIs(t, err, apperrors.ErrNotFound)

	// Delete non-existent
	err = uc.Delete(ctx, "non-existent")
	require.ErrorIs(t, err, apperrors.ErrNotFound)
}
