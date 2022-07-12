package main

import (
	"github.com/metricsAndAlerting/internal/models"
	"log"
	"runtime"
	"time"

	client "github.com/metricsAndAlerting/internal/agent"
)

const (
	reportInterval = 10 * time.Second
	numJobs        = 25
)

func main() {
	respClient := client.NewClient()
	runtimeStats := runtime.MemStats{}
	requestValue := models.RuntimeMetric{}

	metricsChan := make(chan models.GaugeMetric, numJobs)
	pollCountMetricsChan := make(chan models.CountMetric, 1)
	for w := 1; w <= numJobs; w++ {
		go sendMetric(metricsChan, respClient)
		go sendPollCountMetric(pollCountMetricsChan, respClient)
	}
	ticker := time.NewTicker(reportInterval)
	for {
		<-ticker.C
		GetRuntimeStat(&runtimeStats)
		metricValue := requestValue.SetMetricValue(runtimeStats)

		for _, metric := range metricValue {
			if !metric.GaugeMetricISEmpty() {
				metricsChan <- metric
			}
		}

		pollCountMetricsChan <- requestValue.SetPollCountMetricValue()
	}
}

func sendMetric(jobs <-chan models.GaugeMetric, resp *client.Client) {
	for j := range jobs {
		log.Println(j)
		err := resp.SendMetrics(j)
		if err != nil {
			log.Println("Err: ", err.Error())
		}
	}
}

func sendPollCountMetric(jobs <-chan models.CountMetric, resp *client.Client) {
	for j := range jobs {
		log.Println(j)
		err := resp.SendPollCountMetric(j)
		if err != nil {
			log.Println("Err: ", err.Error())
		}
	}
}

func GetRuntimeStat(metrics *runtime.MemStats) {
	runtime.ReadMemStats(metrics)
}
