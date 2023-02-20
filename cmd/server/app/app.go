package app

import (
	"context"
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
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/io"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/postgresql"
)

func Run() {
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

	// init db
	pg := &postgresql.DB{}
	if cfg.Postgres.DSN != "" {
		pg, err = postgresql.NewDB(cfg.DSN)
		if err != nil {
			logger.Error(err)
		}
		err = pg.CreateTable(ctx)
		if err != nil {
			logger.Error(err)
		}
		useDB = true
	}

	// init IO storage
	file := io.NewStorages(ctx, cfg.Storage.IO)

	reportIntervalTicker := time.NewTicker(cfg.StoreInterval)

	go func(ctx context.Context) {
		for {
			<-reportIntervalTicker.C
			log.Println("Write data in storage")
			err := file.SaveMetricInFile()
			if err != nil {
				log.Println(err)
			}
		}

	}(ctx)

	//ginSwagger.WrapHandler(swaggerFiles.Handler,
	//	ginSwagger.URL("http://localhost:8080/swagger/doc.json"),
	//	ginSwagger.DefaultModelsExpandDepth(-1))
	//
	//pprof.Register(r)

	repository := serverRepository.NewServerRepository(*pg, logger)
	storage := serverStorage.NewStorages(cfg, logger)

	useCase := serverUseCase.NewServerUseCase(repository, storage, logger, useDB, cfg.Key)

	rg := serverDelivery.NewRouterGroup(&r.RouterGroup, useCase)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rg.Routes()
	pprof.RouteRegister(&r.RouterGroup, "pprof")

	r.Run(cfg.Address)
}
