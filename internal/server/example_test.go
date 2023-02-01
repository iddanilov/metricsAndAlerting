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

func ExampleUpdateMetricByPath() {

	request := httptest.NewRequest(http.MethodPost, "/update/Gauge/Alloc/5.5", nil)
	request.Header.Set("Content-Type", "text/plain")
	// создаём новый Recorder
	w := httptest.NewRecorder()
	// определяем хендлер
	mu := sync.Mutex{}
	storage := server.Storage{
		Metrics: make(map[string]client.Metrics, 10),
		Mutex:   &mu,
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
