package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	sdklogger "github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-template/internal/config"
	"github.com/VH288/tc-template/internal/infrastructure/database"
)

func wireInfra(ctx context.Context, cfg *config.Config) (*Infra, error) {
	// temp
	logger := sdklogger.New(sdklogger.Config{
		ServiceName: cfg.App.Name,
		LogLevel:    cfg.Observability.LogLevel,
	})
	slog.SetDefault(logger)

	db, err := database.NewPostgres(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	return &Infra{
		DB:     db,
		Logger: logger,
		Config: cfg,
	}, nil
}
