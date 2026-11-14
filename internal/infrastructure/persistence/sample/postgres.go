package sample

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	apperrors "github.com/VH288/tc-sdk/errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresModel struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Status      string    `db:"status"`
	Version     int       `db:"version"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toModel(s *entity.Sample) postgresModel {
	return postgresModel{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		Version:     s.Version,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func fromModel(m postgresModel) *entity.Sample {
	return &entity.Sample{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		Version:     m.Version,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type PostgresRepository struct {
	db *sqlx.DB
}

var _ sample.Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, sample *entity.Sample) error {
	if sample.ID == "" {
		sample.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	sample.CreatedAt = now
	sample.UpdatedAt = now
	if sample.Version == 0 {
		sample.Version = 1
	}

	const q = `
		INSERT INTO samples (id, name, description, status, version, created_at, updated_at)
		VALUES (:id, :name, :description, :status, :version, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, q, toModel(sample)); err != nil {
		return fmt.Errorf("insert sample: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	var model postgresModel
	const q = `
		SELECT id, name, description, status, version, created_at, updated_at
		FROM samples WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &model, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get sample: %w", err)
	}
	return fromModel(model), nil
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM samples`); err != nil {
		return nil, 0, fmt.Errorf("count samples: %w", err)
	}

	var models []postgresModel
	const q = `
		SELECT id, name, description, status, version, created_at, updated_at
		FROM samples
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	if err := r.db.SelectContext(ctx, &models, q, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list samples: %w", err)
	}

	items := make([]entity.Sample, 0, len(models))
	for _, m := range models {
		items = append(items, *fromModel(m))
	}
	return items, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, sample *entity.Sample) error {
	expectedVersion := sample.Version
	sample.UpdatedAt = time.Now().UTC()
	sample.Version = expectedVersion + 1

	const q = `
		UPDATE samples
		SET name = :name, description = :description, status = :status,
		    updated_at = :updated_at, version = :version
		WHERE id = :id AND version = :expected_version
	`
	args := map[string]any{
		"id":               sample.ID,
		"name":             sample.Name,
		"description":      sample.Description,
		"status":           sample.Status,
		"updated_at":       sample.UpdatedAt,
		"version":          sample.Version,
		"expected_version": expectedVersion,
	}
	res, err := r.db.NamedExecContext(ctx, q, args)
	if err != nil {
		return fmt.Errorf("update sample: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		exists, err := r.existsByID(ctx, sample.ID)
		if err != nil {
			return err
		}
		if !exists {
			return apperrors.ErrNotFound
		}
		return apperrors.ErrConflict
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM samples WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete sample: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) existsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM samples WHERE id = $1)`, id); err != nil {
		return false, err
	}
	return exists, nil
}
