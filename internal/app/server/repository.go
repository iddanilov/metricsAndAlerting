package task

import (
	"context"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

// Repository represent the metric and server repository contract
type Repository interface {
	Ping() error
	CreateTable(ctx context.Context) error
	UpdateMetric(ctx context.Context, metrics models.Metrics) error
	UpdateMetrics(metrics []models.Metrics) error
	GetMetric(ctx context.Context, metricID string) (models.Metrics, error)
	GetMetricNames(ctx context.Context) ([]string, error)
	GetCounterMetric(ctx context.Context, metricID string) (*int64, error)
	GetGaugeMetric(ctx context.Context, metricID string) (*float64, error)
}
