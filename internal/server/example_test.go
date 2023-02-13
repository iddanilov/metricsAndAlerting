package server_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/iddanilov/metricsAndAlerting/internal/db"
	client "github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/server"
	"net/http"
	"net/http/httptest"
	"sync"
)

var baseFloat float64 = 5.5

func ExampleRouterGroup_UpdateMetricByPath() {

	request := httptest.NewRequest(http.MethodPost, "/update/Gauge/Alloc/5.5", nil)
	request.Header.Set("Content-Type", "text/plain")
	// создаём новый Recorder
	w := httptest.NewRecorder()
	// определяем хендлер
	mu := sync.Mutex{}
	storage := server.Storage{
		Metrics: make(map[string]client.Metrics, 10),
		Mutex:   mu,
	}
	db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
	if err != nil {
		panic(err)
	}

	r := gin.New()
	r.RedirectTrailingSlash = false
	rg := server.NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
	rg.Routes()

	// запускаем сервер
	r.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 200

}

func ExampleRouterGroup_GetMetric() {

	request := httptest.NewRequest(http.MethodGet, "/value/Gauge/Alloc/", nil)
	request.Header.Set("Content-Type", "text/plain")
	// создаём новый Recorder
	w := httptest.NewRecorder()
	// определяем хендлер
	mu := sync.Mutex{}
	storage := server.Storage{
		Metrics: make(map[string]client.Metrics, 10),
		Mutex:   mu,
	}
	metricResult := client.Metrics{
		ID:    "Alloc",
		MType: "Gauge",
		Value: &baseFloat,
	}
	storage.Metrics[metricResult.ID] = client.Metrics{ID: metricResult.ID, MType: metricResult.MType, Value: metricResult.Value, Delta: metricResult.Delta}

	db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
	if err != nil {
		panic(err)
	}

	r := gin.New()
	r.RedirectTrailingSlash = false
	rg := server.NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, true)
	rg.Routes()

	// запускаем сервер
	r.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 404

}

//func ExampleRouterGroup_Ping() {
//
//	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
//	request.Header.Set("Content-Type", "text/plain")
//	// создаём новый Recorder
//	w := httptest.NewRecorder()
//	// определяем хендлер
//	mu := sync.Mutex{}
//	storage := server.Storage{
//		Metrics: make(map[string]client.Metrics, 10),
//		Mutex:   mu,
//	}
//	db, err := db.NewDB("host=localhost user=admin password=password dbname=postgres port=6432 sslmode=disable")
//	if err != nil {
//		panic(err)
//	}
//
//	r := gin.New()
//	r.RedirectTrailingSlash = false
//	rg := server.NewRouterGroup(&r.RouterGroup, &storage, "LOOOOOOOOOOOOOOL", db, false)
//	rg.Routes()
//
//	// запускаем сервер
//	r.ServeHTTP(w, request)
//	res := w.Result()
//	defer res.Body.Close()
//
//	fmt.Println(res.StatusCode)
//
//	// Output:
//	// 200
//
//}
