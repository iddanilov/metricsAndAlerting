package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	serverDelivery "github.com/iddanilov/metricsAndAlerting/internal/app/server/delivery/http"
	serverRepository "github.com/iddanilov/metricsAndAlerting/internal/app/server/repository/postgres"
	serverStorage "github.com/iddanilov/metricsAndAlerting/internal/app/server/repository/storage"
	serverUseCase "github.com/iddanilov/metricsAndAlerting/internal/app/server/usecase"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/postgresql"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func Run() {

	StartServer()

	var err error
	var useDB bool
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	log.Println("Start Run function")

	// Init config
	cfg := NewConfig()
	log.Println("Init config")

	// Init logger
	logger := logger.GetLogger(cfg.LoggerLevel)
	logger.Logger.Info("Init logger")

	// Create server engine
	r := gin.New()
	r.RedirectTrailingSlash = false

	//ginSwagger.WrapHandler(swaggerFiles.Handler,
	//	ginSwagger.URL("http://localhost:8080/swagger/doc.json"),
	//	ginSwagger.DefaultModelsExpandDepth(-1))
	//
	//pprof.Register(r)

	// init db
	pg := &postgresql.DB{}
	if cfg.Postgres.DSN != "" {
		pg, err = postgresql.NewDB(cfg.DSN)
		if err != nil {
			logger.Error(err)
		}
		useDB = true
	}
	repository, err := serverRepository.NewServerRepository(ctx, *pg, logger, useDB)
	if err != nil {
		logger.Fatal("db didn't create")
	}
	storage := serverStorage.NewStorages(cfg, logger)

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)

	go func(ctx context.Context) {
		for {
			<-reportIntervalTicker.C
			log.Println("Write data in storage")
			err := storage.SaveMetricInFile()
			if err != nil {
				log.Println(err)
			}
		}

	}(ctx)

	useCase := serverUseCase.NewServerUseCase(repository, storage, logger, useDB, cfg.Key)

	rg := serverDelivery.NewRouterGroup(&r.RouterGroup, useCase)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rg.Routes()
	pprof.RouteRegister(&r.RouterGroup, "pprof")

	r.Run(cfg.Address)
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
