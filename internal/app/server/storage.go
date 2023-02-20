package task

import (
	client "github.com/iddanilov/metricsAndAlerting/internal/models"
)

// Storage represent the metric and server's storage
type Storage interface {
	SaveMetricInFile() error
	GetMetric(requestBody client.Metrics) (client.Metrics, error)
	GetMetricValue(name string) (*float64, error)
	GetMetricDelta(name string) (*int64, error)

	SaveGaugeMetric(metric *client.Metrics)
	SaveCountMetric(metric client.Metrics)
	GetMetricsByPath() ([]string, error)
}
