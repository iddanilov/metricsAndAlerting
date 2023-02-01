package app

import (
	"context"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	serverDelivery "github.com/iddanilov/metricsAndAlerting/internal/app/server/delivery/http"
	serverRepository "github.com/iddanilov/metricsAndAlerting/internal/app/server/repository/postgres"
	serverusecase "github.com/iddanilov/metricsAndAlerting/internal/app/server/usecase"
	logger "github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/io"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/postgresql"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"time"
)

func Run() {
	var err error
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

	//var useDB bool

	// init db
	pg := &postgresql.DB{}
	if cfg.Postgres.DSN != "" {
		pg, err = postgresql.NewDB(cfg.DSN)
		if err != nil {
			log.Println(err)
		}
		err = pg.CreateTable(ctx)
		if err != nil {
			log.Println(err)
		}

	}

	// init IO storage
	file := io.NewStorages(ctx, cfg.Storage.IO)

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

	//ginSwagger.WrapHandler(swaggerfiles.Handler,
	//	ginSwagger.URL("http://localhost:8080/swagger/doc.json"),
	//	ginSwagger.DefaultModelsExpandDepth(-1))
	//
	//pprof.Register(r)

	serverRep := serverRepository.NewServerRepository(logger)
	serverUsecase := serverusecase.NewServerUsecase(cfg.Storage, serverRep, logger)

	rg := serverDelivery.NewRouterGroup(&r.RouterGroup, serverUsecase)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	rg.Routes()
	pprof.RouteRegister(&r.RouterGroup, "pprof")

	r.Run(cfg.Address)
}
