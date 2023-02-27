package task

import (
	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

//go:generate mockgen -package mock -destination usecase/storage/mock/server_mock.go -source=usecase.go

// Storage represent the metric and server's storage
type Storage interface {
	SaveMetricInFile() error
	GetMetric(requestBody models.Metrics) (models.Metrics, error)
	GetMetricValue(name string) (*float64, error)
	GetMetricDelta(name string) (*int64, error)
	SaveGaugeMetric(metric *models.Metrics)
	SaveCountMetric(metric models.Metrics)
	GetMetricsList() ([]string, error)
}
