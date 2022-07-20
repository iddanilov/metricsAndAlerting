package server

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
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
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		url  string
		want want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос с обновлением gauge - получаю 200; данные сохранены",
			url:  "/update/Gauge/Alloc/5.5",
			want: want{
				code: 200,
				metricResult: client.GaugeMetric{
					Name:       "Alloc",
					MetricType: "Gauge",
					Value:      5.5,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge с некорректным паттерном - получаю 404; данные не сохранены",
			url:  "/update/Gauge/Alloc/5.5/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.GaugeMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge по незвестному пути - получаю 404; данные не сохранены",
			url:  "/unknown/Gauge/Alloc/5.5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.GaugeMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge без параметра - получаю 404; данные не сохранены",
			url:  "/unknown/Gauge/Alloc/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.GaugeMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge по неполному пути - получаю 404; данные не сохранены",
			url:  "/update/Gauge",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.GaugeMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/update/Gauge/Alloc/none",
			want: want{
				code: 400,
				metricResult: client.GaugeMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/updater/Gauge/Alloc/5.5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.GaugeMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			router := httprouter.New()
			router.RedirectTrailingSlash = false
			mu := sync.Mutex{}
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
				Mutex:   &mu,
			}
			h := NewHandler(&storage)
			h.Register(router)
			// запускаем сервер
			router.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			assert.Equal(t, tt.want.metricResult, storage.Gauge[tt.want.metricResult.Name], "Can't save metric")

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
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		url  string
		want want
	}{
		// определяем все тесты
		{
			name: "Запрос с обновлением counter - получаю 200; данные сохранены",
			url:  "/update/Counter/PollCount/5",
			want: want{
				code: 200,
				metricResult: client.CountMetric{
					Name:       "PollCount",
					MetricType: "Counter",
					Value:      5,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter с некорректным паттерном - получаю 404; данные не сохранены",
			url:  "/update/Counter/PollCount/5/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.CountMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter по незвестному пути - получаю 404; данные не сохранены",
			url:  "/unknown/Counter/PollCount/5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.CountMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter без параметра - получаю 404; данные не сохранены",
			url:  "/unknown/Counter/PollCount/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.CountMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter по неполному пути - получаю 404; данные не сохранены",
			url:  "/update/Counter",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.CountMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/update/Counter/PollCount/none",
			want: want{
				code: 400,
				metricResult: client.CountMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/updater/Counter/PollCount/5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.CountMetric{
					Name:       "",
					MetricType: "",
					Value:      0,
				},
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			router := httprouter.New()
			router.RedirectTrailingSlash = false
			mu := sync.Mutex{}
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
				Mutex:   &mu,
			}
			h := NewHandler(&storage)
			h.Register(router)
			// запускаем сервер
			router.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			assert.Equal(t, tt.want.metricResult, storage.Counter[tt.want.metricResult.Name], "Can't save metric")

			// получаем и проверяем тело запроса
			defer res.Body.Close()
		})
	}
}

func TestGetGauge(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code     int
		response string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		url               string
		gaugeMetricResult client.GaugeMetric
		countMetricResult client.CountMetric
		want              want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос на получение метрики типа gauge с названием Alloc - получаю 200; получаю значение метрики",
			url:  "/value/Gauge/Alloc",
			gaugeMetricResult: client.GaugeMetric{
				Name:       "Alloc",
				MetricType: "Gauge",
				Value:      5.5,
			},
			want: want{
				code:     http.StatusOK,
				response: "5.5",
			},
		},
		{
			name: "[Positive] Запрос на получение метрики типа counter с названием PollCount - получаю 200; получаю значение метрики",
			url:  "/value/Counter/PollCount",
			countMetricResult: client.CountMetric{
				Name:       "PollCount",
				MetricType: "Counter",
				Value:      5,
			},
			want: want{
				code:     http.StatusOK,
				response: "5",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			router := httprouter.New()
			router.RedirectTrailingSlash = false
			mu := sync.Mutex{}
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
				Mutex:   &mu,
			}
			if !tt.gaugeMetricResult.GaugeMetricISEmpty() {
				storage.Gauge[tt.gaugeMetricResult.Name] = tt.gaugeMetricResult
			}
			if !tt.countMetricResult.CountMetricISEmpty() {
				storage.Counter[tt.countMetricResult.Name] = tt.countMetricResult
			}
			h := NewHandler(&storage)
			h.Register(router)
			// запускаем сервер
			router.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equalf(t, string(body), tt.want.response, "Не корректный ответ")
		})
	}
}

func TestGetMetrics(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code     int
		response string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		load *http.Request
		url  string
		want want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос на получение метрик; метрики не загружены - получаю 200; данных нет",
			url:  "/",
			want: want{
				code:     http.StatusOK,
				response: "<h1><ul></ul></h1>",
			},
		},
		{
			name: "[Positive] Запрос на получение метрик; метрика загружена - получаю 200; данные загружены",
			url:  "/",
			load: httptest.NewRequest(http.MethodPost, "/update/Gauge/Alloc/5.5", nil),
			want: want{
				code:     http.StatusOK,
				response: "<h1><ul><li>Alloc</li></ul></h1>",
			},
		},
		{
			name: "[Positive] Запрос на получение метрик; метрика Counter загружена - получаю 200; данные загружены",
			url:  "/",
			load: httptest.NewRequest(http.MethodPost, "/update/Counter/Alloc/5", nil),
			want: want{
				code:     http.StatusOK,
				response: "<h1><ul><li>Alloc</li></ul></h1>",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			log.Println(request)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			router := httprouter.New()
			router.RedirectTrailingSlash = false
			mu := sync.Mutex{}
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
				Mutex:   &mu,
			}
			h := NewHandler(&storage)
			h.Register(router)
			// запускаем сервер
			if tt.load != nil {
				router.ServeHTTP(w, tt.load)
			}
			router.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equalf(t, string(body), tt.want.response, "Не корректный ответ")
		})
	}
}

func TestGetCreateResponse(t *testing.T) {
	// определяем структуру теста

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name              string
		gaugeMetricResult client.GaugeMetric
		countMetricResult client.CountMetric
		result            string
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода GetCreateResponse",
			gaugeMetricResult: client.GaugeMetric{
				Name:       "Alloc",
				MetricType: "Gauge",
				Value:      5.5,
			},
			countMetricResult: client.CountMetric{
				Name:       "PollCount",
				MetricType: "Counter",
				Value:      5,
			},
			result: "<h1><ul><li>Alloc</li><li>PollCount</li></ul></h1>",
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			storage := Storage{
				Gauge:   make(map[string]client.GaugeMetric, 10),
				Counter: make(map[string]client.CountMetric, 10),
			}
			if !tt.gaugeMetricResult.GaugeMetricISEmpty() {
				storage.Gauge[tt.gaugeMetricResult.Name] = tt.gaugeMetricResult
			}
			if !tt.countMetricResult.CountMetricISEmpty() {
				storage.Counter[tt.countMetricResult.Name] = tt.countMetricResult
			}

			assert.Equal(t, createResponse(&storage), tt.result)

		})
	}
}
