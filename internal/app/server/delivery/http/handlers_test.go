package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	serverApp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	serverRepository "github.com/iddanilov/metricsAndAlerting/internal/app/server/repository/postgres"
	"github.com/iddanilov/metricsAndAlerting/internal/app/server/repository/storage"
	serverUseCase "github.com/iddanilov/metricsAndAlerting/internal/app/server/usecase"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/postgresql"
)

var (
	baseFloat     float64 = 5.5
	baseZeroFloat float64 = 0
	baseInt       int64   = 5
	baseZeroInt   int64   = 0
)

func InitTestHandlers(repo bool) *testHandlers {
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
	storage := storage.NewStorages(cfg, logger)

	db, err := postgresql.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
	if err != nil {
		panic(err)
	}
	repository, err := serverRepository.NewServerRepository(context.Background(), *db, logger, repo)
	if err != nil {
		logger.Fatal("db didn't create")
	}

	us := serverUseCase.NewServerUseCase(repository, storage, logger, repo, cfg.Key)

	return &testHandlers{
		us:         us,
		logger:     logger,
		db:         db,
		storage:    storage,
		repository: repository,
	}

}

type testHandlers struct {
	logger     logger.Logger
	us         serverApp.Usecase
	db         *postgresql.DB
	storage    serverApp.Storage
	repository serverApp.Repository
}

func TestSaveGaugeInRepository(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult models.Metrics
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
				code: http.StatusOK,
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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

			// создаём новый Recorder
			w := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			r := gin.New()
			r.RedirectTrailingSlash = false
			initHandlers := InitTestHandlers(true)
			rg := NewRouterGroup(&r.RouterGroup, initHandlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t,
				res.StatusCode,
				tt.want.code,
				fmt.Sprintf("Expected status code %d, got %d", tt.want.code, w.Code),
			)
			time.Sleep(time.Second * 1)

			metricValue, err := initHandlers.repository.GetGaugeMetric(context.Background(), tt.want.metricResult.ID)

			if tt.want.metricResult.ID == "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "metrics not found")
			} else {
				assert.NoError(t, err)
				err = initHandlers.repository.DeleteMetrics(context.Background(), []string{tt.want.metricResult.ID})
				assert.NoError(t, err)
				assert.Equal(t, *tt.want.metricResult.Value, *metricValue, "Can't save metric")
			}
		})
	}
}

func TestSaveGaugeInStorage(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult models.Metrics
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
				code: http.StatusOK,
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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

			// создаём новый Recorder
			w := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			r := gin.New()
			r.RedirectTrailingSlash = false
			initHandlers := InitTestHandlers(false)
			rg := NewRouterGroup(&r.RouterGroup, initHandlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			// проверяем код ответа
			assert.Equal(t,
				res.StatusCode,
				tt.want.code,
				fmt.Sprintf("Expected status code %d, got %d", tt.want.code, w.Code),
			)
			metricValue, err := initHandlers.storage.GetMetricValue(tt.want.metricResult.ID)

			if tt.want.metricResult.ID == "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "metrics not found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tt.want.metricResult.Value, *metricValue, "Can't save metric")
			}
		})
	}
}

func TestSaveCounterInRepository(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult models.Metrics
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
			url:  "/update/Counter/TestPollCount/5",
			want: want{
				code: 200,
				metricResult: models.Metrics{
					ID:    "TestPollCount",
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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

			handlers := InitTestHandlers(true)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, handlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			time.Sleep(time.Second * 1)

			metricDelta, err := handlers.repository.GetCounterMetric(context.Background(), tt.want.metricResult.ID)

			if tt.want.metricResult.ID != "" {
				assert.NoError(t, err)
				err = handlers.repository.DeleteMetrics(context.Background(), []string{tt.want.metricResult.ID})
				assert.NoError(t, err)
				assert.Equal(t, *tt.want.metricResult.Delta, *metricDelta, "Can't save metric")
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSaveCounterInStorage(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code         int
		metricResult models.Metrics
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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
				metricResult: models.Metrics{
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

			handlers := InitTestHandlers(false)

			// создаём новый Recorder
			w := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, handlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			time.Sleep(time.Second * 1)

			metricDelta, err := handlers.storage.GetMetricDelta(tt.want.metricResult.ID)

			if tt.want.metricResult.ID != "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.metricResult.Delta, metricDelta, "Can't save metric")
			} else {
				assert.Error(t, err)
			}
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
		name         string
		url          string
		metricResult models.Metrics
		want         want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос на получение метрики типа gauge с названием Alloc - получаю 200; получаю значение метрики",
			url:  "/value/Gauge/Alloc",
			metricResult: models.Metrics{
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
			metricResult: models.Metrics{
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
			// создаём новый Recorder
			w := httptest.NewRecorder()

			handlers := InitTestHandlers(true)

			err := handlers.repository.UpdateMetric(context.Background(), tt.metricResult)
			assert.NoError(t, err)

			defer func() {
				err = handlers.repository.DeleteMetrics(context.Background(), []string{tt.metricResult.ID})
				assert.NoError(t, err)
			}()

			time.Sleep(time.Second * 1)

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, handlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, http.StatusOK, res.StatusCode)

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want.response, string(body), "Не корректный ответ")
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
		metricResult        models.Metrics
		counterMetricResult models.Metrics
		want                want
	}{
		// определяем все тесты
		{
			name: "[Positive] Запрос на получение метрики типа gauge с названием Alloc - получаю 200; получаю значение метрики",
			url:  "/value/Gauge/Alloc",
			metricResult: models.Metrics{
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
			metricResult: models.Metrics{
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
			handlers := InitTestHandlers(true)

			err := handlers.repository.UpdateMetric(context.Background(), tt.metricResult)
			assert.NoError(t, err)

			defer func() {
				err = handlers.repository.DeleteMetrics(context.Background(), []string{tt.metricResult.ID})
				assert.NoError(t, err)
			}()

			time.Sleep(time.Second * 1)

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			logger := logger.GetLogger("debug")
			logger.Logger.Info("Init logger")

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, handlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, res.StatusCode, tt.want.code,
				fmt.Sprintf("Expected status code %d, got %d", tt.want.code, w.Code))

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
		name         string
		metricResult models.Metrics
		url          string
		want         want
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
			metricResult: models.Metrics{
				ID:    "Alloc",
				MType: "Gauge",
				Value: &baseFloat,
			},
			want: want{
				code:     http.StatusOK,
				response: "<h1><ul><li>Alloc</li></ul></h1>",
			},
		},
		{
			name: "[Positive] Запрос на получение метрик; метрика Counter загружена - получаю 200; данные загружены",
			url:  "/",
			metricResult: models.Metrics{
				ID:    "Alloc",
				MType: "Counter",
				Delta: &baseInt,
			},
			want: want{
				code:     http.StatusOK,
				response: "<h1><ul><li>Alloc</li></ul></h1>",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			handlers := InitTestHandlers(true)

			fmt.Println("tt.metricResult.MetricISEmpty()", tt.metricResult.MetricISEmpty())
			fmt.Println("tt.metricResult", tt.metricResult)

			if !tt.metricResult.MetricISEmpty() {
				err := handlers.repository.UpdateMetric(context.Background(), tt.metricResult)
				assert.NoError(t, err)

				defer func() {
					err = handlers.repository.DeleteMetrics(context.Background(), []string{tt.metricResult.ID})
					assert.NoError(t, err)
				}()
			}

			time.Sleep(time.Second * 1)

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			log.Println(request)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			logger := logger.GetLogger("debug")
			logger.Logger.Info("Init logger")

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, handlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want.response, string(body), "Не корректный ответ")
		})
	}
}

func TestGetCreateResponse(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code     int
		response string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name         string
		metricResult []models.Metrics
		url          string
		want         want
	}{
		// определяем все тесты
		{
			name: "[Positive] Проверка метода GetCreateResponse",
			url:  "/",
			metricResult: []models.Metrics{
				{
					ID:    "Alloc",
					MType: "Gauge",
					Value: &baseZeroFloat,
				},
				{
					ID:    "PollCount",
					MType: "Counter",
					Delta: &baseInt,
				},
			},
			want: want{
				http.StatusOK,
				"<h1><ul><li>Alloc</li><li>PollCount</li></ul></h1>",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			handlers := InitTestHandlers(true)

			for _, metrics := range tt.metricResult {
				err := handlers.repository.UpdateMetric(context.Background(), metrics)
				assert.NoError(t, err)
			}
			defer func() {
				for _, metrics := range tt.metricResult {
					err := handlers.repository.DeleteMetrics(context.Background(), []string{metrics.ID})
					assert.NoError(t, err)
				}
			}()

			time.Sleep(time.Second * 1)

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			// создаём новый Recorder
			w := httptest.NewRecorder()
			logger := logger.GetLogger("debug")
			logger.Logger.Info("Init logger")

			r := gin.New()
			r.RedirectTrailingSlash = false
			rg := NewRouterGroup(&r.RouterGroup, handlers.us)
			rg.Routes()

			// запускаем сервер
			r.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, res.StatusCode, tt.want.code,
				fmt.Sprintf("Expected status code %d, got %d", tt.want.code, w.Code))

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equalf(t, string(body), tt.want.response, "Не корректный ответ")
		})
	}
}
