package main

import (
	"context"
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
	storage := server.NewStorages(cfg)

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)
	times := make(chan int64, 1)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close(times)
				log.Println("Stop program")
				os.Exit(100)
			default:
				<-reportIntervalTicker.C
				err := storage.SaveMetricInFile(ctx)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

	}(ctx)
	r := gin.New()
	r.RedirectTrailingSlash = false

	rg := server.NewRouterGroup(&r.RouterGroup, storage)
	rg.Routes()

	r.Run(cfg.Address)
}
