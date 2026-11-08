package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/VH288/tc-template/internal/infrastructure/metrics"
)

func TestMetrics_HandlerAndRecording(t *testing.T) {
	metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
	metrics.HTTPRequestDuration.WithLabelValues("GET", "/test").Observe(0.05)
	metrics.GRPCRequestsTotal.WithLabelValues("/sample.v1.SampleService/Create", "ok").Inc()
	metrics.SampleOperationsTotal.WithLabelValues("create", "success").Inc()

	handler := metrics.Handler()
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, "http_request_duration_seconds")
	assert.Contains(t, body, "grpc_requests_total")
	assert.Contains(t, body, "sample_operations_total")
}
