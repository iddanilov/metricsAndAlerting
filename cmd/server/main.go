package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	log.Println("create router")

	cfg := server.NewConfig()
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

	r := gin.New()

	rg := server.NewRouterGroup(&r.RouterGroup, storage)
	rg.Routes()

	r.Run(cfg.ADDRESS)
}
