package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/VH288/tc-sdk/httpmiddleware"
	"github.com/VH288/tc-template/internal/app/middleware"
	"github.com/VH288/tc-template/internal/infrastructure/metrics"
)

func TestMetricsAndTracingMiddleware(t *testing.T) {
	oldProvider := otel.GetTracerProvider()
	oldPropagator := otel.GetTextMapPropagator()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(oldProvider)
		otel.SetTextMapPropagator(oldPropagator)
		_ = tp.Shutdown(context.Background())
	})

	var capturedSpan trace.Span
	r := chi.NewRouter()
	r.Use(httpmiddleware.Tracing("test-service"))
	r.Use(middleware.Metrics)
	r.Get("/test/{id}", func(w http.ResponseWriter, req *http.Request) {
		capturedSpan = trace.SpanFromContext(req.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test/42", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedSpan)
	assert.True(t, capturedSpan.SpanContext().IsValid())

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "GET /test/42", spans[0].Name())

	var hasMethod, hasRoute, hasStatus bool
	for _, attr := range spans[0].Attributes() {
		switch attr.Key {
		case "http.method":
			assert.Equal(t, "GET", attr.Value.AsString())
			hasMethod = true
		case "http.route":
			assert.Equal(t, "/test/42", attr.Value.AsString())
			hasRoute = true
		case "http.status_code":
			assert.Equal(t, int64(200), attr.Value.AsInt64())
			hasStatus = true
		}
	}
	assert.True(t, hasMethod)
	assert.True(t, hasRoute)
	assert.True(t, hasStatus)

	// Check metrics
	metricsRec := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Contains(t, metricsRec.Body.String(), `http_requests_total{method="GET",path="/test/{id}",status="200"}`)
}

func TestMetricsAuth(t *testing.T) {
	auth := middleware.MetricsAuth("secret-123")
	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Unauthorized without key
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Authorized with key
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("X-API-Key", "secret-123")
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCORSMiddleware(t *testing.T) {
	handler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	// Test regular GET request gets CORS headers
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/samples", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")

	// Test OPTIONS preflight request
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/api/v1/samples", nil)
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "X-API-Key")
}
