package bootstrap

import (
	"context"
	"net/http"

	"github.com/VH288/tc-template/internal/app/middleware"
	"github.com/VH288/tc-template/internal/app/router"
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

func (c *Container) HTTPRouter() http.Handler {
	auth := middleware.SkipPaths(
		[]string{"/api/v1/health"},
		middleware.APIKeyAuth(c.Config.App.Secret),
	)

	return router.New(router.Dependencies{
		SampleHTTP:  c.Domain.Sample.HTTP,
		Middlewares: c.middlewares(),
		MetricsPath: c.Config.Observability.MetricsPath,
		MetricsAuth: middleware.MetricsAuth(c.Config.App.Secret),
		APIAuth:     auth,
		DB:          c.Infra.DB,
	})
}

func (c *Container) middlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.CORS,
		middleware.Recoverer(c.Logger),
		middleware.RequestID,
		middleware.Tracing(c.Config.Observability.ServiceName),
		middleware.Metrics,
		middleware.Logging(c.Logger),
	}
}
