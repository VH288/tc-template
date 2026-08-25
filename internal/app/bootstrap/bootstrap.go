package bootstrap

import (
	"context"

	"github.com/VH288/tc-template/internal/config"
)

func Build(ctx context.Context, cfg *config.Config) (*Container, error) {
	infra, err := wireInfra(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:    cfg,
		Logger:    infra.Logger,
		Infra:     infra,
		Domain:    getDomain(infra),
		Lifecycle: wireLifecycle(infra),
	}, nil
}
