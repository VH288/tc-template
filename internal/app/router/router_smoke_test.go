package router_test

import (
	"context"
	"net/http"
	"testing"

	apperrors "github.com/VH288/tc-sdk/errors"
	"github.com/VH288/tc-sdk/formatter"
	"github.com/VH288/tc-sdk/httpmiddleware"
	"github.com/VH288/tc-template/internal/app/middleware"
	"github.com/VH288/tc-template/internal/app/router"
	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/dto"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	samplehandler "github.com/VH288/tc-template/internal/domain/sample/handler"
	"github.com/VH288/tc-template/internal/domain/sample/usecase"
	"github.com/VH288/tc-template/internal/infrastructure/telemetry"
	"github.com/VH288/tc-template/internal/pkg/testutil"

	"github.com/stretchr/testify/require"
)

type smokeMemRepo struct {
	items map[string]*entity.Sample
}

func newSmokeMemRepo() *smokeMemRepo {
	return &smokeMemRepo{items: make(map[string]*entity.Sample)}
}

var _ sample.Repository = (*smokeMemRepo)(nil)

func (m *smokeMemRepo) Create(ctx context.Context, sample *entity.Sample) error {
	sample.ID = "smoke-sample-1"
	sample.Version = 1
	cp := *sample
	m.items[sample.ID] = &cp
	return nil
}

func (m *smokeMemRepo) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *smokeMemRepo) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	out := make([]entity.Sample, 0, len(m.items))
	for _, s := range m.items {
		out = append(out, *s)
	}
	return out, int64(len(out)), nil
}

func (m *smokeMemRepo) Update(ctx context.Context, sample *entity.Sample) error {
	if _, ok := m.items[sample.ID]; !ok {
		return apperrors.ErrNotFound
	}
	sample.Version++
	cp := *sample
	m.items[sample.ID] = &cp
	return nil
}

func (m *smokeMemRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func buildTestHandler(repo sample.Repository, secret string) http.Handler {
	uc := usecase.New(repo, telemetry.NoopLogger{}, telemetry.NoopTracer{}, telemetry.NoopSampleMetrics{})
	h := samplehandler.NewHTTPHandler(uc)

	auth := middleware.SkipPaths(
		[]string{"/api/v1/health"},
		middleware.APIKeyAuth(secret),
	)

	return router.New(router.Dependencies{
		SampleHTTP: h,
		Middlewares: []func(http.Handler) http.Handler{
			middleware.Recoverer(telemetry.DiscardLogger()),
			httpmiddleware.RequestID,
			middleware.Metrics,
			httpmiddleware.Logging(telemetry.DiscardLogger()),
		},
		APIAuth: auth,
		DB:      nil,
	})
}

func TestSmokeRoutes(t *testing.T) {
	repo := newSmokeMemRepo()
	appRouter := buildTestHandler(repo, "SECRET_KEY") // dev mode: auth bypassed

	// 1. GET /
	rr := testutil.PerformRequest(t, appRouter, http.MethodGet, "/", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)
	require.JSONEq(t, `{"service":"tc-template"}`, rr.Body.String())

	// 2. GET /api/v1/health (DB nil -> 503)
	rr = testutil.PerformRequest(t, appRouter, http.MethodGet, "/api/v1/health", nil)
	testutil.AssertStatus(t, rr, http.StatusServiceUnavailable)

	// 3. POST /api/v1/samples (Valid)
	createReq := dto.CreateSampleRequest{
		Name:        "Concert VIP",
		Description: "Early entrance",
		Status:      "active",
	}
	rr = testutil.PerformRequest(t, appRouter, http.MethodPost, "/api/v1/samples", createReq)
	testutil.AssertStatus(t, rr, http.StatusCreated)

	var createResp formatter.Response
	testutil.DecodeJSON(t, rr, &createResp)
	require.True(t, createResp.Success)
	require.Equal(t, "sample created", createResp.Message)

	// 4. GET /api/v1/samples
	rr = testutil.PerformRequest(t, appRouter, http.MethodGet, "/api/v1/samples", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var listResp formatter.Response
	testutil.DecodeJSON(t, rr, &listResp)
	require.True(t, listResp.Success)
	require.NotNil(t, listResp.Meta)
	require.Equal(t, int64(1), listResp.Meta.Total)

	// 5. GET /api/v1/samples/{id}
	rr = testutil.PerformRequest(t, appRouter, http.MethodGet, "/api/v1/samples/smoke-sample-1", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var getResp formatter.Response
	testutil.DecodeJSON(t, rr, &getResp)
	require.True(t, getResp.Success)

	// 6. PUT /api/v1/samples/{id}
	updateReq := dto.UpdateSampleRequest{
		Name:        "Concert VIP Updated",
		Description: "Updated desc",
		Status:      "inactive",
	}
	rr = testutil.PerformRequest(t, appRouter, http.MethodPut, "/api/v1/samples/smoke-sample-1", updateReq)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var updateResp formatter.Response
	testutil.DecodeJSON(t, rr, &updateResp)
	require.True(t, updateResp.Success)
	require.Equal(t, "sample updated", updateResp.Message)

	// 7. DELETE /api/v1/samples/{id}
	rr = testutil.PerformRequest(t, appRouter, http.MethodDelete, "/api/v1/samples/smoke-sample-1", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var deleteResp formatter.Response
	testutil.DecodeJSON(t, rr, &deleteResp)
	require.True(t, deleteResp.Success)
	require.Equal(t, "sample deleted", deleteResp.Message)

	// 8. POST /api/v1/samples (Invalid -> 422)
	invalidReq := dto.CreateSampleRequest{
		Name: "", // missing/invalid
	}
	rr = testutil.PerformRequest(t, appRouter, http.MethodPost, "/api/v1/samples", invalidReq)
	testutil.AssertStatus(t, rr, http.StatusUnprocessableEntity)

	var errResp formatter.Response
	testutil.DecodeJSON(t, rr, &errResp)
	require.False(t, errResp.Success)
	require.NotNil(t, errResp.Error)
	require.Equal(t, string(apperrors.CodeValidation), errResp.Error.Code)

	// 9. GET /metrics
	rr = testutil.PerformRequest(t, appRouter, http.MethodGet, "/metrics", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)
	require.Contains(t, rr.Body.String(), "http_requests_total")
	require.Contains(t, rr.Body.String(), "http_request_duration_seconds")
}
func TestSmokeAuth(t *testing.T) {
	repo := newSmokeMemRepo()
	appRouter := buildTestHandler(repo, "custom-secret-key-123")

	// Protected route without key -> 401
	rr := testutil.PerformRequest(t, appRouter, http.MethodGet, "/api/v1/samples", nil)
	testutil.AssertStatus(t, rr, http.StatusUnauthorized)

	// Protected route with wrong key -> 401
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/samples", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	rr2 := testutil.PerformRequest(t, appRouter, http.MethodGet, "/api/v1/samples", nil)
	testutil.AssertStatus(t, rr2, http.StatusUnauthorized)

	// Public route bypasses auth -> 200
	rr3 := testutil.PerformRequest(t, appRouter, http.MethodGet, "/", nil)
	testutil.AssertStatus(t, rr3, http.StatusOK)
}
