package server

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"

	client "github.com/metricsAndAlerting/internal/models"
)

func TestSaveGaugeMetric(t *testing.T) {
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		gaugeMetricResult client.GaugeMetric
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода SaveGaugeMetric",
			gaugeMetricResult: client.GaugeMetric{
				Name:       "Alloc",
				MetricType: "Gauge",
				Value:      5.5,
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			mu := sync.Mutex{}
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
				Mutex:   &mu,
			}
			storage.SaveGaugeMetric(tt.gaugeMetricResult)

			assert.Equal(t, tt.gaugeMetricResult, storage.Gauge[tt.gaugeMetricResult.Name])

		})
	}
}

func TestSaveCounterMetric(t *testing.T) {
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		countMetricResult client.CountMetric
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода SaveCounterMetric",
			countMetricResult: client.CountMetric{
				Name:       "PollCount",
				MetricType: "Counter",
				Value:      5,
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			mu := sync.Mutex{}
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
				Mutex:   &mu,
			}
			storage.SaveCountMetric(tt.countMetricResult)

			assert.Equal(t, tt.countMetricResult, storage.Counter[tt.countMetricResult.Name])

		})
	}
}
