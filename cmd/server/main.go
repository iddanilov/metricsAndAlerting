package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"

	client "github.com/metricsAndAlerting/internal/models"
	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	log.Println("create router")

	cfg := server.NewConfig()

	router := httprouter.New()
	router.RedirectTrailingSlash = false

	storage := server.Storage{
		Metrics: make(map[string]client.Metrics, 10),
		Mutex:   &sync.Mutex{},
	}

	log.Println("register service handler")
	handler := server.NewHandler(&storage)
	handler.Register(router)

	s := &http.Server{
		Addr:         cfg.ADDRESS,
		Handler:      router,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}
