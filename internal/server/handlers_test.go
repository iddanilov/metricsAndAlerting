package server

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/iddanilov/metricsAndAlerting/internal/db"
	client "github.com/iddanilov/metricsAndAlerting/internal/models"
)

var (
	baseFloat     float64 = 5.5
	baseZeroFloat float64 = 0
	baseInt       int64   = 5
	baseZeroInt   int64   = 0
)

func TestSendGauge(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult client.Metrics
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
				metricResult: client.Metrics{
					ID:    "Alloc",
					MType: "Gauge",
					Value: &baseFloat,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge с некорректным паттерном - получаю 404; данные не сохранены",
			url:  "/update/Gauge/Alloc/5.5/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Value: &baseZeroFloat,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge по незвестному пути - получаю 404; данные не сохранены",
			url:  "/unknown/Gauge/Alloc/5.5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Value: &baseZeroFloat,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge без параметра - получаю 404; данные не сохранены",
			url:  "/unknown/Gauge/Alloc/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Value: &baseZeroFloat,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge по неполному пути - получаю 404; данные не сохранены",
			url:  "/update/Gauge",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Value: &baseZeroFloat,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/update/Gauge/Alloc/none",
			want: want{
				code: 400,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Value: &baseZeroFloat,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением gauge с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/updater/Gauge/Alloc/5.5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Value: &baseZeroFloat,
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
			mu := sync.Mutex{}
			storage := Storage{
				Metrics: make(map[string]client.Metrics, 10),
				Mutex:   &mu,
			}
			db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
			if err != nil {
				panic(err)
			}

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			assert.Equal(t, tt.want.metricResult.ID, storage.Metrics[tt.want.metricResult.ID].ID, "Can't save metric")
			assert.Equal(t, tt.want.metricResult.MType, storage.Metrics[tt.want.metricResult.ID].MType, "Can't save metric")
			if *tt.want.metricResult.Value != 0 {
				assert.Equal(t, *tt.want.metricResult.Value, *storage.Metrics[tt.want.metricResult.ID].Value, "Can't save metric")
			} else {
				assert.Equal(t, storage.Metrics, map[string]client.Metrics{})
			}
			// получаем и проверяем тело запроса
			defer res.Body.Close()
		})
	}
}

func TestSendCounter(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult client.Metrics
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
				metricResult: client.Metrics{
					ID:    "PollCount",
					MType: "Counter",
					Delta: &baseInt,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter с некорректным паттерном - получаю 404; данные не сохранены",
			url:  "/update/Counter/PollCount/5/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Delta: &baseZeroInt,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter по незвестному пути - получаю 404; данные не сохранены",
			url:  "/unknown/Counter/PollCount/5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Delta: &baseZeroInt,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter без параметра - получаю 404; данные не сохранены",
			url:  "/unknown/Counter/PollCount/",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Delta: &baseZeroInt,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter по неполному пути - получаю 404; данные не сохранены",
			url:  "/update/Counter",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Delta: &baseZeroInt,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/update/Counter/PollCount/none",
			want: want{
				code: 400,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Delta: &baseZeroInt,
				},
			},
		},
		{
			name: "[Negative] Запрос с обновлением counter с некорректным параметром - получаю 400; данные не сохранены",
			url:  "/updater/Counter/PollCount/5",
			want: want{
				code: http.StatusNotFound,
				metricResult: client.Metrics{
					ID:    "",
					MType: "",
					Delta: &baseZeroInt,
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
			mu := sync.Mutex{}
			storage := Storage{
				Metrics: make(map[string]client.Metrics, 10),
				Mutex:   &mu,
			}
			db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
			if err != nil {
				panic(err)
			}
			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			assert.Equal(t, tt.want.metricResult.ID, storage.Metrics[tt.want.metricResult.ID].ID, "Can't save metric")
			assert.Equal(t, tt.want.metricResult.MType, storage.Metrics[tt.want.metricResult.ID].MType, "Can't save metric")
			if *tt.want.metricResult.Delta != 0 {
				assert.Equal(t, *tt.want.metricResult.Delta, *storage.Metrics[tt.want.metricResult.ID].Delta, "Can't save metric")
			} else {
				assert.Equal(t, storage.Metrics, map[string]client.Metrics{})
			}
			// получаем и проверяем тело запроса
			defer res.Body.Close()
		})
	}
}

func TestGetMetric(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code     int
		response string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                string
		url                 string
		metricResult        client.Metrics
		counterMetricResult client.Metrics
		want                want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос на получение метрики типа gauge с названием Alloc - получаю 200; получаю значение метрики",
			url:  "/value/Gauge/Alloc",
			metricResult: client.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &baseFloat,
			},
			want: want{
				code:     http.StatusOK,
				response: "5.5",
			},
		},
		{
			name: "[Negative] Запрос на получение метрики типа counter с названием PollCount - получаю 200; получаю значение метрики",
			url:  "/value/Counter/PollCount",
			counterMetricResult: client.Metrics{
				ID:    "PollCount",
				MType: "Counter",
				Delta: &baseInt,
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
			request.Header.Set("Content-Type", "application/json")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			mu := sync.Mutex{}
			storage := Storage{
				Metrics: make(map[string]client.Metrics, 10),
				Mutex:   &mu,
			}
			if !tt.metricResult.MetricISEmpty() {
				storage.Metrics[tt.metricResult.ID] = client.Metrics{ID: tt.metricResult.ID, MType: tt.metricResult.MType, Value: tt.metricResult.Value, Delta: tt.metricResult.Delta}
			}
			if !tt.counterMetricResult.MetricISEmpty() {
				storage.Metrics[tt.counterMetricResult.ID] = client.Metrics{ID: tt.counterMetricResult.ID, MType: tt.counterMetricResult.MType, Value: tt.counterMetricResult.Value, Delta: tt.counterMetricResult.Delta}
			}
			r := gin.New()
			r.RedirectTrailingSlash = false
			db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
			if err != nil {
				panic(err)
			}
			rg := NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
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

func TestGetGauge(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code     int
		response string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name                string
		url                 string
		metricResult        client.Metrics
		counterMetricResult client.Metrics
		want                want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос на получение метрики типа gauge с названием Alloc - получаю 200; получаю значение метрики",
			url:  "/value/Gauge/Alloc",
			metricResult: client.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &baseFloat,
			},
			want: want{
				code:     http.StatusOK,
				response: "5.5",
			},
		},
		{
			name: "[Positive] Запрос на получение метрики типа counter с названием PollCount - получаю 200; получаю значение метрики",
			url:  "/value/Counter/PollCount",
			metricResult: client.Metrics{
				ID:    "PollCount",
				MType: "Counter",
				Delta: &baseInt,
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
			mu := sync.Mutex{}
			storage := Storage{
				Metrics: make(map[string]client.Metrics, 10),
				Mutex:   &mu,
			}
			if !tt.metricResult.MetricISEmpty() {
				storage.Metrics[tt.metricResult.ID] = client.Metrics{ID: tt.metricResult.ID, MType: tt.metricResult.MType, Value: tt.metricResult.Value, Delta: tt.metricResult.Delta}
				storage.Metrics[tt.metricResult.ID] = client.Metrics{ID: tt.metricResult.ID, MType: tt.metricResult.MType, Value: tt.metricResult.Value, Delta: tt.metricResult.Delta}
			}
			if !tt.counterMetricResult.MetricISEmpty() {
				storage.Metrics[tt.counterMetricResult.ID] = client.Metrics{ID: tt.counterMetricResult.ID, MType: tt.counterMetricResult.MType, Value: tt.counterMetricResult.Value, Delta: tt.counterMetricResult.Delta}
			}
			db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
			if err != nil {
				panic(err)
			}

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
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
			mu := sync.Mutex{}
			storage := Storage{
				Metrics: make(map[string]client.Metrics, 10),
				Mutex:   &mu,
			}
			db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
			if err != nil {
				panic(err)
			}
			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
			rg.Routes()

			// запускаем сервер
			// запускаем сервер
			if tt.load != nil {
				r.ServeHTTP(w, tt.load)
			}
			r.ServeHTTP(w, request)
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
		metricResult      client.Metrics
		countMetricResult client.Metrics
		result            string
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода GetCreateResponse",
			metricResult: client.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &baseZeroFloat,
			},
			countMetricResult: client.Metrics{
				ID:    "PollCount",
				MType: "Counter",
				Delta: &baseInt,
			},
			result: "<h1><ul><li>Alloc</li><li>PollCount</li></ul></h1>",
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			storage := Storage{
				Metrics: make(map[string]client.Metrics, 10),
			}
			if !tt.metricResult.MetricISEmpty() {
				storage.Metrics[tt.metricResult.ID] = client.Metrics{ID: tt.metricResult.ID, MType: tt.metricResult.MType, Value: tt.metricResult.Value, Delta: tt.metricResult.Delta}

			}
			if !tt.countMetricResult.MetricISEmpty() {
				storage.Metrics[tt.countMetricResult.ID] = client.Metrics{ID: tt.countMetricResult.ID, MType: tt.countMetricResult.MType, Value: tt.countMetricResult.Value, Delta: tt.countMetricResult.Delta}
			}

			assert.True(t, strings.Contains(createResponse(&storage), tt.metricResult.ID))
			assert.True(t, strings.Contains(createResponse(&storage), tt.countMetricResult.ID))

		})
	}
}
