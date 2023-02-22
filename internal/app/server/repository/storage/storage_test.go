package storage

import (
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

var floatValue float64 = 5.5
var intValue int64 = 5

func TestSaveGaugeMetric(t *testing.T) {
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		gaugeMetricResult models.Metrics
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода SaveGaugeMetric",
			gaugeMetricResult: models.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &floatValue,
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			logger := logger.GetLogger("debug")
			logger.Logger.Info("Init logger")
			cfg := &models.Server{
				Storage: models.Storage{
					IO: models.IO{
						Restore:   true,
						StoreFile: "/tmp/devops-metrics-db.json",
					},
				},
			}

			storage := NewStorages(cfg, logger)
			storage.SaveGaugeMetric(&tt.gaugeMetricResult)
			metric, err := storage.GetMetricValue(tt.gaugeMetricResult.ID)
			assert.NoError(t, err)
			assert.Equal(t, tt.gaugeMetricResult.Value, metric)
		})
	}
}

func TestSaveCounterMetric(t *testing.T) {
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		countMetricResult models.Metrics
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода SaveCounterMetric",
			countMetricResult: models.Metrics{
				ID:    "PollCount",
				MType: "Counter",
				Delta: &intValue,
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		logger := logger.GetLogger("debug")
		logger.Logger.Info("Init logger")
		cfg := &models.Server{
			Storage: models.Storage{
				IO: models.IO{
					Restore:   true,
					StoreFile: "/tmp/devops-metrics-db.json",
				},
			},
		}

		storage := NewStorages(cfg, logger)
		storage.SaveCountMetric(tt.countMetricResult)
		metric, err := storage.GetMetricDelta(tt.countMetricResult.ID)
		assert.NoError(t, err)
		assert.Equal(t, tt.countMetricResult.Delta, metric)
	}
}
