package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	client "github.com/metricsAndAlerting/internal/models"
)

func TestSendGauge(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult client.GaugeMetric
	}
	type metric struct {
		name       string
		metricType string
		value      float64
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		metric metric
		want   want
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			metric: metric{
				name:       "Alloc",
				metricType: "Gauge",
				value:      5.5,
			},
			want: want{
				code: 200,
				metricResult: client.GaugeMetric{
					Name:       "Alloc",
					MetricType: "Gauge",
					Value:      5.5,
				},
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, "/update"+fmt.Sprintf("/%s/%s/%v/", tt.metric.metricType, tt.metric.name, tt.metric.value), nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			router := httprouter.New()
			storage := Storage{}
			h := NewHandler(&storage)
			h.Register(router)
			// запускаем сервер
			router.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			assert.Equal(t, tt.want.metricResult, storage.Alloc, "Can't save metric")

			// получаем и проверяем тело запроса
			defer res.Body.Close()
		})
	}
}

func TestSendCounter(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult client.CountMetric
	}
	type metric struct {
		name       string
		metricType string
		value      int64
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		metric metric
		want   want
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			metric: metric{
				name:       "PollCount",
				metricType: "counter",
				value:      5,
			},
			want: want{
				code: 200,
				metricResult: client.CountMetric{
					Name:       "PollCount",
					MetricType: "Counter",
					Value:      5,
				},
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, "/update"+fmt.Sprintf("/%s/%s/%v/", tt.metric.metricType, tt.metric.name, tt.metric.value), nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			router := httprouter.New()
			storage := Storage{}
			h := NewHandler(&storage)
			h.Register(router)
			// запускаем сервер
			router.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			assert.Equal(t, tt.want.metricResult, storage.PollCount, "Can't save metric")

			// получаем и проверяем тело запроса
			defer res.Body.Close()
		})
	}
}
