// Package tasks provides polling task implementations.
package tasks

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task/poller"
)

// MetricsCollectHandler simulates collecting system metrics.
type MetricsCollectHandler struct {
	logger *zap.Logger
}

// NewMetricsCollectHandler creates a new metrics collection handler.
func NewMetricsCollectHandler(logger *zap.Logger) *MetricsCollectHandler {
	return &MetricsCollectHandler{
		logger: logger.Named("metrics-collect-handler"),
	}
}

// Register registers this handler with the poller.
func (t *MetricsCollectHandler) Register(p *poller.Poller) error {
	// Runs every 30 seconds
	if err := p.Register("metrics-collect", 30*time.Second, t.Execute); err != nil {
		return fmt.Errorf("failed to register metrics-collect handler: %w", err)
	}
	return nil
}

// Execute runs the metrics collection handler.
func (t *MetricsCollectHandler) Execute(ctx context.Context) error {
	t.logger.Info("Collecting system metrics")

	// Simulate collecting various metrics
	metrics := map[string]float64{
		"requests_per_second": float64(rand.Intn(1000)),
		"response_time_ms":    float64(rand.Intn(500)),
		"error_rate":          rand.Float64() * 0.1,
	}

	t.logger.Info("Metrics collected",
		zap.Float64("requests_per_second", metrics["requests_per_second"]),
		zap.Float64("response_time_ms", metrics["response_time_ms"]),
		zap.Float64("error_rate", metrics["error_rate"]))

	return nil
}
