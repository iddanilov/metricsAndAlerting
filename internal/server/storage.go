package server

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	client "github.com/iddanilov/metricsAndAlerting/internal/models"
)

type Storage struct {
	Metrics map[string]client.Metrics
	mutex   sync.Mutex
	File    string
}

func NewStorages(cfg *Config) *Storage {
	events := make(map[string]client.Metrics, 10)
	if cfg.Restore {
		result, err := ReadEvents(cfg.StoreFile)
		if err != nil {
			log.Fatal(err)
		}
		if result != nil {
			events = result
		}
	}
	return &Storage{
		Metrics: events,
		mutex:   sync.Mutex{},
		File:    cfg.StoreFile,
	}
}

func ReadEvents(fileName string) (metrics map[string]client.Metrics, err error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	defer func(file *os.File) {
		err = file.Close()
	}(file)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(file).Decode(&metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *Storage) SaveMetricInFile(ctx context.Context) error {
	if len(s.Metrics) == 0 {
		return nil
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	file, err := os.OpenFile(s.File, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	err = json.NewEncoder(file).Encode(s.Metrics)
	if err != nil {
		return err
	}
	return err
}

func (s *Storage) SaveGaugeMetric(metric *client.Metrics) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Metrics[metric.ID] = *metric
}

func (s *Storage) SaveCountMetric(metric client.Metrics) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	result := s.Metrics[metric.ID]
	if result.Delta != nil {
		*metric.Delta = *metric.Delta + *result.Delta
	}
	s.Metrics[metric.ID] = metric

}
