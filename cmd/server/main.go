package main

import (
	"context"
	"github.com/metricsAndAlerting/internal/db"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/metricsAndAlerting/internal/server"
)

func main() {
	log.Println("create router")

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg := server.NewConfig()
	file := server.NewStorages(cfg)

	storage, err := db.NewDB(cfg.DSN)
	if err != nil {
		panic(err)
	}
	err = storage.CreateTable(ctx)
	if err != nil {
		panic(err)
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
				err := file.SaveMetricInFile(ctx)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

	}(ctx)
	r := gin.New()
	r.RedirectTrailingSlash = false

	rg := server.NewRouterGroup(&r.RouterGroup, file, cfg.Key, storage)
	rg.Routes()

	r.Run(cfg.Address)
}
