package usecase

import (
	"context"

	"github.com/VH288/tc-template/internal/domain/ports"
	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	"github.com/VH288/tc-sdk/pagination"
)

type SampleUsecase struct {
	repo    sample.Repository
	log     ports.Logger
	trace   ports.Tracer
	metrics ports.SampleMetrics
}

func New(
	repo sample.Repository,
	log ports.Logger,
	trace ports.Tracer,
	metrics ports.SampleMetrics,
) *SampleUsecase {
	return &SampleUsecase{
		repo:    repo,
		log:     log,
		trace:   trace,
		metrics: metrics,
	}
}

func (u *SampleUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	if u.trace != nil {
		var end func()
		ctx, end = u.trace.Start(ctx, "sample.usecase.Create")
		defer end()
	}

	if status == "" {
		status = entity.StatusActive
	}

	s := &entity.Sample{
		Name:        name,
		Description: description,
		Status:      status,
	}

	if err := u.repo.Create(ctx, s); err != nil {
		if u.metrics != nil {
			u.metrics.IncOperation("create", "error")
		}
		if u.log != nil {
			u.log.Error(ctx, "failed to create sample", "error", err)
		}
		return nil, err
	}

	if u.metrics != nil {
		u.metrics.IncOperation("create", "success")
	}
	if u.log != nil {
		u.log.Info(ctx, "sample created", "id", s.ID)
	}
	return s, nil
}

func (u *SampleUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	if u.trace != nil {
		var end func()
		ctx, end = u.trace.Start(ctx, "sample.usecase.GetByID")
		defer end()
	}

	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if u.metrics != nil {
			u.metrics.IncOperation("get", "error")
		}
		return nil, err
	}

	if u.metrics != nil {
		u.metrics.IncOperation("get", "success")
	}
	return s, nil
}

func (u *SampleUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	if u.trace != nil {
		var end func()
		ctx, end = u.trace.Start(ctx, "sample.usecase.List")
		defer end()
	}

	page, perPage, offset := pagination.NormalizePagination(page, perPage)
	items, total, err := u.repo.List(ctx, perPage, offset)
	if err != nil {
		if u.metrics != nil {
			u.metrics.IncOperation("list", "error")
		}
		return nil, 0, err
	}

	if u.metrics != nil {
		u.metrics.IncOperation("list", "success")
	}
	return items, total, nil
}

func (u *SampleUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	if u.trace != nil {
		var end func()
		ctx, end = u.trace.Start(ctx, "sample.usecase.Update")
		defer end()
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if u.metrics != nil {
			u.metrics.IncOperation("update", "error")
		}
		return nil, err
	}

	existing.Name = name
	existing.Description = description
	existing.Status = status

	if err := u.repo.Update(ctx, existing); err != nil {
		if u.metrics != nil {
			u.metrics.IncOperation("update", "error")
		}
		if u.log != nil {
			u.log.Error(ctx, "failed to update sample", "id", id, "error", err)
		}
		return nil, err
	}

	if u.metrics != nil {
		u.metrics.IncOperation("update", "success")
	}
	if u.log != nil {
		u.log.Info(ctx, "sample updated", "id", existing.ID)
	}
	return existing, nil
}

func (u *SampleUsecase) Delete(ctx context.Context, id string) error {
	if u.trace != nil {
		var end func()
		ctx, end = u.trace.Start(ctx, "sample.usecase.Delete")
		defer end()
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		if u.metrics != nil {
			u.metrics.IncOperation("delete", "error")
		}
		if u.log != nil {
			u.log.Error(ctx, "failed to delete sample", "id", id, "error", err)
		}
		return err
	}

	if u.metrics != nil {
		u.metrics.IncOperation("delete", "success")
	}
	if u.log != nil {
		u.log.Info(ctx, "sample deleted", "id", id)
	}
	return nil
}
