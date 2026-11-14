package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/VH288/tc-sdk/apm"
	"github.com/VH288/tc-sdk/database"
	sdklogger "github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-template/internal/config"
)

func wireInfra(ctx context.Context, cfg *config.Config) (*Infra, error) {
	// temp
	logger := sdklogger.New(sdklogger.Config{
		ServiceName: cfg.App.Name,
		LogLevel:    cfg.Observability.LogLevel,
	})
	slog.SetDefault(logger)

	apmProvider, err := apm.Init(ctx, apm.Config{
		TempoEndpoint:    cfg.Observability.TempoEndpoint,
		ServiceName:      cfg.Observability.ServiceName,
		TraceSampleRatio: cfg.Observability.TraceSampleRatio,
	}, logger)
	if err != nil {
		logger.Warn("apm init failed, continuing without remote traces", "error", err)
	}

	db, err := database.NewPostgres(database.PostgresConfig{
		DSN:             cfg.Postgres.DSN(),
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	return &Infra{
		DB:     db,
		APM:    apmProvider,
		Logger: logger,
		Config: cfg,
	}, nil
}
