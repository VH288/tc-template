package sample

import (
	"context"

	"github.com/VH288/tc-template/internal/domain/sample/entity"
)

type Repository interface {
	Create(ctx context.Context, sample *entity.Sample) error
	GetByID(ctx context.Context, id string) (*entity.Sample, error)
	List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error)
	Update(ctx context.Context, sample *entity.Sample) error
	Delete(ctx context.Context, id string) error
}

type Usecase interface {
	Create(ctx context.Context, name, description, status string) (*entity.Sample, error)
	GetByID(ctx context.Context, id string) (*entity.Sample, error)
	List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error)
	Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error)
	Delete(ctx context.Context, id string) error
}
