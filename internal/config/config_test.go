package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VH288/tc-template/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := config.Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "info", cfg.Observability.LogLevel)
	assert.Equal(t, "http://127.0.0.1:3100", cfg.Observability.LokiURL)
	assert.Equal(t, "127.0.0.1:4317", cfg.Observability.TempoEndpoint)
	assert.Equal(t, "/metrics", cfg.Observability.MetricsPath)
	assert.Equal(t, "tc-template", cfg.Observability.ServiceName)
	assert.Equal(t, 1.0, cfg.Observability.TraceSampleRatio)
}
