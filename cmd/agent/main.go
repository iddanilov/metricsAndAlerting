package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
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

func hash(m string, key []byte) (string, error) {
	src := []byte(m) // данные, которые хотим зашифровать

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]
	dst := aesgcm.Seal(nil, nonce, src, nil)
	log.Println("dst: ", dst)
	return hex.EncodeToString(dst), err
	// создаём вектор инициализации
}

func sendMetrics(jobs <-chan []models.Metrics, resp *client.Client) {
	for j := range jobs {
		for _, metrics := range j {
			if !metrics.MetricISEmpty() {
				if metrics.Value != nil {
					hashValue, err := hash(fmt.Sprintf("%s:gauge:%f", metrics.ID, *metrics.Value), []byte(resp.Config.Key))
					if err != nil {
						log.Fatal(err)
					}
					metrics.Hash = hashValue
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

func GetRuntimeStat(metrics *runtime.MemStats) {
	runtime.ReadMemStats(metrics)
}
