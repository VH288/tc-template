package usecase

import (
	"context"

	"github.com/VH288/tc-template/internal/domain/ports"
	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	"github.com/VH288/tc-template/internal/pkg/utils"
)

type SampleUsecase struct {
	repo sample.Repository
	log  ports.Logger
}

func New(repo sample.Repository, log ports.Logger) *SampleUsecase {
	return &SampleUsecase{
		repo: repo,
		log:  log,
	}
}

func (u *SampleUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	if status == "" {
		status = entity.StatusActive
	}

	s := &entity.Sample{
		Name:        name,
		Description: description,
		Status:      status,
	}

	if err := u.repo.Create(ctx, s); err != nil {
		if u.log != nil {
			u.log.Error(ctx, "failed to create sample", "error", err)
		}
		return nil, err
	}

	if u.log != nil {
		u.log.Info(ctx, "sample created", "id", s.ID)
	}
	return s, nil
}

func (u *SampleUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *SampleUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	page, perPage, offset := utils.NormalizePagination(page, perPage)
	return u.repo.List(ctx, perPage, offset)
}

func (u *SampleUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = name
	existing.Description = description
	existing.Status = status

	if err := u.repo.Update(ctx, existing); err != nil {
		if u.log != nil {
			u.log.Error(ctx, "failed to update sample", "id", id, "error", err)
		}
		return nil, err
	}

	if u.log != nil {
		u.log.Info(ctx, "sample updated", "id", existing.ID)
	}
	return existing, nil
}

func (u *SampleUsecase) Delete(ctx context.Context, id string) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		if u.log != nil {
			u.log.Error(ctx, "failed to delete sample", "id", id, "error", err)
		}
		return err
	}

	if u.log != nil {
		u.log.Info(ctx, "sample deleted", "id", id)
	}
	return nil
}
