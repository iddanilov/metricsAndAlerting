// Package server/main running server application
package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/iddanilov/metricsAndAlerting/pkg/certs"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/iddanilov/metricsAndAlerting/internal/db"
	"github.com/iddanilov/metricsAndAlerting/internal/server"
)

var buildVersion string
var buildDate string
var buildCommit string

// @title Metric and Alerting
// @version 0.0.2
// @description API Server for Metric and Alerting Application

// @host localhost:8000
// @BasePath /

func main() {
	StartServer()

	var useDB bool
	var err error
	storage := &db.DB{}
	log.Println("create router")

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	cfg := server.NewConfig()
	file := server.NewStorages(cfg)

	if cfg.DSN != "" {
		storage, err = db.NewDB(cfg.DSN)
		if err != nil {
			log.Println(err)
		}
		err = storage.CreateTable(ctx)
		if err != nil {
			log.Println(err)
		}

		useDB = true
	}
	log.Println(useDB)

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)

	go writeDBScheduler(ctx, reportIntervalTicker, file)

	r := gin.New()

	ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("http://localhost:8080/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1))

	pprof.Register(r)

	r.RedirectTrailingSlash = false

	rg := server.NewRouterGroup(&r.RouterGroup, file, cfg.Key, storage, useDB)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	rg.Routes()

	pprof.RouteRegister(&r.RouterGroup, "pprof")

	var privateKey *rsa.PrivateKey

	cert := certs.NewCrypto(privateKey)

	if cfg.CryptoKey != "" {
		r.RunTLS(cfg.Address, cert.Path, cfg.CryptoKey)
	} else {
		r.Run()
	}
}

func StartServer() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "N/A" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = ""
	}
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}

func writeDBScheduler(ctx context.Context, reportIntervalTicker *time.Ticker, file *server.Storage) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopped by user")
			log.Println("Write data in file")
			file.SaveMetricInFile()
			os.Exit(0)
		case <-reportIntervalTicker.C:
			log.Println("Write data in file")
			err := file.SaveMetricInFile()
			if err != nil {
				log.Println(err)
			}
		}
	}

}
