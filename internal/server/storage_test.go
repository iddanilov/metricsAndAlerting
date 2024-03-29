package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	client "github.com/iddanilov/metricsAndAlerting/internal/models"
)

var floatValue = 5.5

func TestSaveGaugeMetric(t *testing.T) {
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		gaugeMetricResult client.Metrics
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода SaveGaugeMetric",
			gaugeMetricResult: client.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &floatValue,
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			storage := NewStorages(cfg)
			storage.SaveGaugeMetric(&tt.gaugeMetricResult)

			assert.Equal(t, tt.gaugeMetricResult, storage.Metrics[tt.gaugeMetricResult.ID])

		})
	}
}

func TestSaveCounterMetric(t *testing.T) {
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		countMetricResult client.Metrics
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода SaveCounterMetric",
			countMetricResult: client.Metrics{
				ID:    "PollCount",
				MType: "Counter",
				Value: &floatValue,
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			storage := NewStorages(cfg)
			storage.SaveCountMetric(tt.countMetricResult)

			assert.Equal(t, tt.countMetricResult, storage.Metrics[tt.countMetricResult.ID])

		})
	}
}
