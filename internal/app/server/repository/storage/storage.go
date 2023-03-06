package storage

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"

	serverapp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
)

type serverStorage struct {
	Metrics map[string]models.Metrics
	mutex   sync.Mutex
	File    string
	logger  logger.Logger
}

func NewStorages(cfg *models.Server, logger logger.Logger) serverapp.Storage {
	events := make(map[string]models.Metrics, 10)
	if cfg.Restore {
		result, err := readEvents(cfg.StoreFile)
		if err != nil {
			logger.Fatal(err)
		}
		if result != nil {
			events = result
		}
	}
	return &serverStorage{
		Metrics: events,
		mutex:   sync.Mutex{},
		File:    cfg.StoreFile,
		logger:  logger,
	}
}

func (s *serverStorage) GetMetricsList() (values []string, err error) {
	for _, m := range s.Metrics {
		values = append(values, m.ID)
	}
	return
}

func (s *serverStorage) GetMetricValue(name string) (*float64, error) {
	metric, ok := s.Metrics[name]
	if !ok {
		return nil, errors.New("metrics not found")
	}
	if metric.Value == nil {
		return nil, errors.New("metrics not found")
	}
	return metric.Value, nil
}
func (s *serverStorage) GetMetricDelta(name string) (*int64, error) {
	metric, ok := s.Metrics[name]
	if !ok {
		return nil, errors.New("metrics not found")
	}
	if metric.Delta == nil {
		return nil, errors.New("metrics not found")
	}
	return metric.Delta, nil
}

func (s *serverStorage) SaveMetricInFile() error {
	if len(s.Metrics) == 0 {
		return nil
	}
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

func (s *serverStorage) GetMetric(requestBody models.Metrics) (models.Metrics, error) {
	metric, ok := s.Metrics[requestBody.ID]
	if !ok {
		return models.Metrics{}, errors.New("metrics not found")
	}
	metric.MType = strings.ToLower(metric.MType)
	return metric, nil
}

func (s *serverStorage) SaveGaugeMetric(metric *models.Metrics) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Metrics[metric.ID] = *metric
}

func (s *serverStorage) SaveCountMetric(metric models.Metrics) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	result := s.Metrics[metric.ID]
	if result.Delta != nil {
		*metric.Delta = *metric.Delta + *result.Delta
	}
	s.Metrics[metric.ID] = metric

}

func readEvents(fileName string) (metrics map[string]models.Metrics, err error) {
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
