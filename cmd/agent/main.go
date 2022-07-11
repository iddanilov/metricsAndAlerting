package main

import (
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
	requestValue := client.RuntimeMetric{}

	metricsChan := make(chan client.GaugeMetric, numJobs)
	pollCountMetricsChan := make(chan client.CountMetric, 1)
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

func sendMetric(jobs <-chan client.GaugeMetric, resp *client.Client) {
	for j := range jobs {
		log.Println(j)
		err := resp.SendMetrics(j)
		if err != nil {
			log.Println("Err: ", err.Error())
		}
	}
}

func sendPollCountMetric(jobs <-chan client.CountMetric, resp *client.Client) {
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
