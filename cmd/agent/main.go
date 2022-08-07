package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	client "github.com/metricsAndAlerting/internal/agent"
	"github.com/metricsAndAlerting/internal/models"
)

const numJobs = 25

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	respClient := client.NewClient()
	runtimeStats := runtime.MemStats{}
	requestValue := models.Metrics{}
	var counter models.Counter

	metricsChan := make(chan []models.Metrics, numJobs)
	for w := 1; w <= numJobs; w++ {
		go sendMetrics(metricsChan, respClient)
	}
	reportIntervalTicker := time.NewTicker(respClient.Config.ReportInterval)
	pollIntervalTicker := time.NewTicker(respClient.Config.PollInterval)
	for {
		select {
		case <-ctx.Done():
			close(metricsChan)
			log.Println("Stopped by user")
			os.Exit(0)
		default:
			<-pollIntervalTicker.C
			GetRuntimeStat(&runtimeStats)
			metricValue := requestValue.SetMetrics(&runtimeStats)

			go func() {
				<-reportIntervalTicker.C
				GetRuntimeStat(&runtimeStats)
				metricValue = requestValue.SetMetrics(&runtimeStats)
				metricsChan <- metricValue
				metricsChan <- counter.SetPollCountMetricValue()
			}()
		}
	}
}

func sendMetrics(jobs <-chan []models.Metrics, resp *client.Client) {
	for j := range jobs {
		for _, metrics := range j {
			if !metrics.MetricISEmpty() {
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

func GetRuntimeStat(metrics *runtime.MemStats) {
	runtime.ReadMemStats(metrics)
}
