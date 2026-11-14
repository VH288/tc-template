package router

import (
	"net/http"

	samplehandler "github.com/VH288/tc-template/internal/domain/sample/handler"
	sampletransport "github.com/VH288/tc-template/internal/domain/sample/transport"
	"github.com/VH288/tc-template/internal/infrastructure/database"
	"github.com/VH288/tc-template/internal/infrastructure/metrics"
	"github.com/VH288/tc-template/internal/pkg/constant"
	"github.com/VH288/tc-sdk/formatter"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

type Dependencies struct {
	SampleHTTP  *samplehandler.HTTPHandler
	Middlewares []func(http.Handler) http.Handler
	MetricsPath string
	MetricsAuth func(http.Handler) http.Handler
	APIAuth     func(http.Handler) http.Handler
	DB          *sqlx.DB
}

func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	for _, mw := range deps.Middlewares {
		r.Use(mw)
	}

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service":"tc-template"}`))
	})

	if deps.MetricsPath == "" {
		deps.MetricsPath = "/metrics"
	}
	metricsHandler := metrics.Handler()
	if deps.MetricsAuth != nil {
		metricsHandler = deps.MetricsAuth(metricsHandler)
	}
	r.Handle(deps.MetricsPath, metricsHandler)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			if deps.DB == nil {
				formatter.Success(w, http.StatusServiceUnavailable, "healthcheck", map[string]string{
					"status":   constant.HealthStatusDOWN,
					"postgres": "db not configured",
				})
				return
			}
			if err := database.Ping(r.Context(), deps.DB); err != nil {
				formatter.Success(w, http.StatusServiceUnavailable, "healthcheck", map[string]string{
					"status":   constant.HealthStatusDOWN,
					"postgres": err.Error(),
				})
				return
			}
			formatter.Success(w, http.StatusOK, "healthcheck", map[string]string{
				"status":   constant.HealthStatusUP,
				"postgres": constant.HealthStatusUP,
			})
		})

		api.Group(func(protected chi.Router) {
			if deps.APIAuth != nil {
				protected.Use(deps.APIAuth)
			}
			sampletransport.RegisterHTTP(protected, deps.SampleHTTP)
		})
	})

	return r
}
