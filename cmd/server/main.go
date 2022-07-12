package main

import (
	client "github.com/metricsAndAlerting/internal/models"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	log.Println("create router")
	router := httprouter.New()
	router.RedirectTrailingSlash = false

	storage := server.Storage{
		Gauge:   make(map[string]client.GaugeMetric, 10),
		Counter: make(map[string]client.CountMetric, 10),
	}

	log.Println("register service handler")
	mutex := sync.Mutex{}
	handler := server.NewHandler(&storage, &mutex)
	handler.Register(router)

	s := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      router,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}
