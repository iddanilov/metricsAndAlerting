// Package server/main running server application
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"

	"github.com/iddanilov/metricsAndAlerting/internal/db"
	"github.com/iddanilov/metricsAndAlerting/internal/server"
	grpcServer "github.com/iddanilov/metricsAndAlerting/internal/server/grpc"
	pb "github.com/iddanilov/metricsAndAlerting/proto"
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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

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

	go func(ctx context.Context) {
		for {
			<-reportIntervalTicker.C
			log.Println("Write data in file")
			err := file.SaveMetricInFile()
			if err != nil {
				log.Println(err)
			}
		}

	}(ctx)
	r := gin.New()

	ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("http://localhost:8080/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1))

	pprof.Register(r)

	r.RedirectTrailingSlash = false

	rg := server.NewRouterGroup(&r.RouterGroup, file, cfg.Key, storage, useDB, cfg.TrustedSubnet)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	rg.Routes()
	pprof.RouteRegister(&r.RouterGroup, "pprof")

	go ListingGrpc(file, storage)

	r.Run(cfg.Address)
}

func ListingGrpc(storage *server.Storage, db *db.DB) {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer()

	grpcServer := grpcServer.NewMetricsAndAlertingServer(storage, db)

	// регистрируем сервис
	pb.RegisterMetricsAndAlertingServer(s, grpcServer)

	fmt.Println("Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
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
