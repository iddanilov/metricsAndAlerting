package app

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	agentUseCase "github.com/iddanilov/metricsAndAlerting/internal/app/agent/usecase"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
)

const (
	//
	addr = ":2222" // адрес сервера
)

const numJobs = 25

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func Run() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	cfg := NewConfig()

	StartServer()

	// Init logger
	logger := logger.GetLogger(cfg.AgentConfig.LoggerLevel)
	logger.Logger.Info("Init logger")

	useCase := agentUseCase.NewAgentUseCase(logger, cfg.AgentConfig.Key, cfg.AgentConfig.Address)

	runtimeStats := runtime.MemStats{}
	requestValue := models.Metrics{}
	var metricValues []models.Metrics
	var memMetricValues []models.Metrics
	var counter models.Counter

	metricsChan := make(chan []models.Metrics, numJobs)
	for w := 1; w <= numJobs; w++ {
		go useCase.SendMetrics(metricsChan, cfg)
	}
	reportIntervalTicker := time.NewTicker(cfg.AgentConfig.ReportInterval)
	pollIntervalTicker := time.NewTicker(cfg.AgentConfig.PollInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(metricsChan)
				logger.Info("Stopped by user")
				os.Exit(0)
			case <-pollIntervalTicker.C:
				go func() {
					useCase.GetRuntimeStat(&runtimeStats)
					metricValues = requestValue.SetMetrics(&runtimeStats)
				}()
				go func() {
					memMetricValues = requestValue.SetVirtualMemoryMetrics(useCase.GetVirtualMemoryStat(ctx))
				}()
			case <-reportIntervalTicker.C:
				go func() {
					useCase.GetRuntimeStat(&runtimeStats)
					metricValues = requestValue.SetMetrics(&runtimeStats)
					metricsChan <- metricValues
					metricsChan <- memMetricValues
					metricsChan <- counter.SetPollCountMetricValue()
				}()
			}
		}
	}()

	http.ListenAndServe(addr, nil)
}

func StartServer() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}
