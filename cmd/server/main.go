package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/metricsAndAlerting/internal/db"
	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	var useDB bool
	var err error
	storage := &db.DB{}
	log.Println("create router")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	cfg := server.NewConfig()
	file := server.NewStorages(cfg)

	if cfg.DSN != "" {
		storage, err = db.NewDB(cfg.DSN)
		if err != nil {
			log.Println(err)
			//panic(err)
		}
		err = storage.CreateTable(ctx)
		if err != nil {
			log.Println(err)
			//panic(err)
		}

		useDB = true
	}
	log.Println(useDB)

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)

	go func(ctx context.Context) {
		for {
			<-reportIntervalTicker.C
			log.Println("Write data in file")
			err := file.SaveMetricInFile(ctx)
			if err != nil {
				log.Println(err)
			}
		}

	}(ctx)
	r := gin.New()
	r.RedirectTrailingSlash = false

	rg := server.NewRouterGroup(&r.RouterGroup, file, cfg.Key, storage, useDB)
	rg.Routes()

	r.Run(cfg.Address)
}
