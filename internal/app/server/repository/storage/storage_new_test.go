package storage

import (
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMetricsList(t *testing.T) {
	tests := []struct {
		name         string
		metricResult []models.Metrics
		expected     []string
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода TestGetMetricsList with Gauge",
			metricResult: []models.Metrics{
				{
					ID:    "Alloc",
					MType: "Gauge",
					Value: &floatValue,
				},
			},
			expected: []string{
				"Alloc",
			},
		},
		{
			name: "[Positive] Проверка метода TestGetMetricsList with Counter",
			metricResult: []models.Metrics{
				{
					ID:    "Counter",
					MType: "Counter",
					Delta: &intValue,
				},
			},
			expected: []string{
				"Counter",
			},
		},
		{
			name: "[Positive] Проверка метода TestGetMetricsList with Counter",
			metricResult: []models.Metrics{
				{
					ID:    "Alloc",
					MType: "Gauge",
					Value: &floatValue,
				},
				{
					ID:    "Counter",
					MType: "Counter",
					Delta: &intValue,
				},
			},
			expected: []string{
				"Counter",
				"Alloc",
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			cfg := &models.Server{}

			// Init logger
			logger := logger.GetLogger("debug")
			logger.Logger.Info("Init logger")
			s := NewStorages(cfg, logger)

			for _, metrics := range tt.metricResult {
				if metrics.MType == "Gauge" {
					s.SaveGaugeMetric(&metrics)
				} else {
					s.SaveCountMetric(metrics)
				}
			}

			result, err := s.GetMetricsList()
			assert.NoError(t, err)
			assert.Equal(t, result, tt.expected)
		})
	}
}
