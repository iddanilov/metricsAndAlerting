package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/metricsAndAlerting/internal/db"
	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	var useDB bool
	storage := &db.DB{}
	log.Println("create router")

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg := server.NewConfig()
	file := server.NewStorages(cfg)

	if cfg.DSN != "" {
		useDB = true
	}
	log.Println(cfg.DSN)
	log.Println(useDB)
	if useDB {
		storage, err := db.NewDB(cfg.DSN)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		err = storage.CreateTable(ctx)
		if err != nil {
			log.Println(err)
			panic(err)
		}
	}

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)
	times := make(chan int64, 1)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close(times)
				log.Println("Stop program")
				os.Exit(0)
			default:
				<-reportIntervalTicker.C
				log.Println("Write data in file")
				err := file.SaveMetricInFile(ctx)
				if err != nil {
					log.Println(err)
				}
			}
		}

	}(ctx)
	r := gin.New()
	r.RedirectTrailingSlash = false

	rg := server.NewRouterGroup(&r.RouterGroup, file, cfg.Key, storage, useDB)
	rg.Routes()

	r.Run(cfg.Address)
}
