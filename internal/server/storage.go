package server

import (
	"log"
	"sync"

	client "github.com/metricsAndAlerting/internal/models"
)

type Storage struct {
	Metrics map[string]client.Metrics
	Mutex   *sync.Mutex
}

func (s *Storage) SaveMetric(metric client.Metrics) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Metrics[metric.ID] = metric
}

func (s *Storage) SaveCountMetric(metric client.Metrics) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	result, ok := s.Metrics[metric.ID]
	if ok {
		*result.Delta = *result.Delta + *metric.Delta
		log.Println(result.Delta)
		s.Metrics[metric.ID] = result
	} else {
		s.Metrics[metric.ID] = metric
	}
}
