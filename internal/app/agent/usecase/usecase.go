package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	agentApp "github.com/iddanilov/metricsAndAlerting/internal/app/agent"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/crypto"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type agentUseCase struct {
	logger     logger.Logger
	address    string
	key        string
	httpClient *http.Client
}

func NewAgentUseCase(logger logger.Logger, key string) agentApp.AgentUseCase {
	return &agentUseCase{
		logger: logger,
		httpClient: &http.Client{
			Timeout: time.Minute,
		},
		key: key,
	}
}

func (u *agentUseCase) SendMetricByPath(params models.Metrics) error {
	var value string
	if strings.ToLower(params.MType) == "gauge" {
		value = strconv.FormatFloat(*params.Value, 'f', 6, 64)
	} else {
		value = strconv.FormatInt(*params.Delta, 10)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/%s/%s/%v", u.address, params.MType, params.ID, value), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if err := u.sendRequest(req); err != nil {
		return err
	}

	return nil
}

func (u *agentUseCase) SendMetric(metrics models.Metrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", u.address), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := u.sendRequest(req); err != nil {
		return err
	}

	return nil
}

func (u *agentUseCase) sendRequest(req *http.Request) error {
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil

}

func (u *agentUseCase) SendMetrics(jobs <-chan []models.Metrics, cfg *models.Agent) {
	var hashValue string
	var err error
	for j := range jobs {
		for _, metrics := range j {
			if !metrics.MetricISEmpty() {
				if u.key != "" {
					if metrics.Value != nil {
						hashValue, err = crypto.CreateHash(
							fmt.Sprintf(
								"%s:gauge:%f", metrics.ID, *metrics.Value), []byte(cfg.AgentConfig.Key))
						if err != nil {
							log.Fatal(err)
						}
						metrics.Hash = hashValue
					} else if metrics.Delta != nil {
						if u.key != "" {
							hashValue, err = crypto.CreateHash(
								fmt.Sprintf(
									"%s:counter:%d", metrics.ID, *metrics.Delta), []byte(cfg.AgentConfig.Key))
							if err != nil {
								log.Fatal(err)
							}
							metrics.Hash = hashValue
						}
					}
				}

				log.Println("body: ", metrics)

				err := u.SendMetricByPath(metrics)
				if err != nil {
					log.Println("Err: ", err.Error())
				}
				err = u.SendMetric(metrics)
				if err != nil {
					log.Println("Err: ", err.Error())
				}
			}
		}
	}
}

// GetRuntimeStat - получение значение MemStats из runtime
func (u *agentUseCase) GetRuntimeStat(metrics *runtime.MemStats) {
	runtime.ReadMemStats(metrics)
}

func (u *agentUseCase) GetVirtualMemoryStat(ctx context.Context) *mem.VirtualMemoryStat {
	metrics, err := mem.VirtualMemory()
	if err != nil {
		log.Println("Error: ", err)
		<-ctx.Done()
	}
	return metrics
}
