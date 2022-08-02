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

	cfg := server.NewConfig()
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	storage := server.NewStorages(cfg)

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)

	go func() {
		for {
			<-reportIntervalTicker.C
			err := storage.SaveMetricInFile()
			if err != nil {
				log.Fatal(err)
			}
		}

	}()

	log.Println("register service handler")
	handler := server.NewHandler(storage)
	handler.Register(router)

	s := &http.Server{
		Addr:         cfg.ADDRESS,
		Handler:      router,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}
