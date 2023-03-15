// Package agent/main running agent application
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	client "github.com/iddanilov/metricsAndAlerting/internal/agent"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

const (
	//
	addr = ":2222" // адрес сервера
)

const numJobs = 25

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	StartServer()

	respClient := client.NewClient()
	runtimeStats := runtime.MemStats{}
	requestValue := models.Metrics{}
	var metricValues []models.Metrics
	var memMetricValues []models.Metrics
	var counter models.Counter

	metricsChan := make(chan []models.Metrics, numJobs)
	for w := 1; w <= numJobs; w++ {
		go sendMetrics(metricsChan, respClient)
	}
	reportIntervalTicker := time.NewTicker(respClient.Config.ReportInterval)
	pollIntervalTicker := time.NewTicker(respClient.Config.PollInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(metricsChan)
				log.Println("Stopped by user")
				os.Exit(0)
			case <-pollIntervalTicker.C:
				go func() {
					GetRuntimeStat(&runtimeStats)
					metricValues = requestValue.SetMetrics(&runtimeStats)
				}()
				go func() {
					memMetricValues = requestValue.SetVirtualMemoryMetrics(GetVirtualMemoryStat(ctx))
				}()
			case <-reportIntervalTicker.C:
				go func() {
					GetRuntimeStat(&runtimeStats)
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

func hash(m string, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(m))
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func sendMetrics(jobs <-chan []models.Metrics, resp *client.Client) {
	var hashValue string
	var err error
	for j := range jobs {
		for _, metrics := range j {
			if !metrics.MetricISEmpty() {
				if resp.Config.Key != "" {
					if metrics.Value != nil {
						hashValue, err = hash(fmt.Sprintf("%s:gauge:%f", metrics.ID, *metrics.Value), []byte(resp.Config.Key))
						if err != nil {
							log.Fatal(err)
						}
						metrics.Hash = hashValue
					} else if metrics.Delta != nil {
						if resp.Config.Key != "" {
							hashValue, err = hash(fmt.Sprintf("%s:counter:%d", metrics.ID, *metrics.Delta), []byte(resp.Config.Key))
							if err != nil {
								log.Fatal(err)
							}
							metrics.Hash = hashValue
						}
					}
				}

				log.Println("body: ", metrics)

				err := resp.SendMetricByPath(metrics)
				if err != nil {
					log.Println("Err: ", err.Error())
				}
				err = resp.SendMetrics(metrics)
				if err != nil {
					log.Println("Err: ", err.Error())
				}
			}
		}
	}
}

// GetRuntimeStat - получение значение MemStats из runtime
func GetRuntimeStat(metrics *runtime.MemStats) {
	runtime.ReadMemStats(metrics)
}

func GetVirtualMemoryStat(ctx context.Context) *mem.VirtualMemoryStat {
	metrics, err := mem.VirtualMemory()
	if err != nil {
		log.Println("Error: ", err)
		<-ctx.Done()
	}
	return metrics
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
