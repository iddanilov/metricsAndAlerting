package storage

import (
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMetricsList(t *testing.T) {
	tests := []struct {
		name              string
		gaugeMetricResult models.Metrics
		expected          []string
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода TestGetMetricsList with Gauge",
			gaugeMetricResult: models.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &floatValue,
			},
			expected: []string{
				"Alloc",
			},
		},
		{
			name: "[Positive] Проверка метода TestGetMetricsList with Counter",
			gaugeMetricResult: models.Metrics{
				ID:    "Counter",
				MType: "Counter",
				Delta: &intValue,
			},
			expected: []string{
				"Alloc",
			},
		},
		{
			name: "[Positive] Проверка метода TestGetMetricsList with Counter",
			gaugeMetricResult: models.Metrics{
				ID:    "Counter",
				MType: "Counter",
				Delta: &intValue,
			},
			expected: []string{
				"Alloc",
				"Counter",
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
			s.SaveCountMetric(tt.gaugeMetricResult)

			result, err := s.GetMetricsList()

			assert.NoError(t, err)
			assert.Equal(t, result, tt.expected)
		})
	}
}
