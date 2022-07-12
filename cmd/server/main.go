package main

import (
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	log.Println("create router")
	router := httprouter.New()

	storage := server.Storage{}

	log.Println("register service handler")
	handler := server.NewHandler(&storage)
	handler.Register(router)

	s := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}
