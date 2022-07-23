package server

import (
	"log"
	"sync"

	client "github.com/metricsAndAlerting/internal/models"
)

type Storage struct {
	Gauge   map[string]client.GaugeMetric
	Counter map[string]client.CountMetric
	Mutex   *sync.Mutex
}

func (s *Storage) SaveGaugeMetric(metric client.GaugeMetric) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Gauge[metric.Name] = metric
}

func (s *Storage) SaveCountMetric(metric client.CountMetric) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	result, ok := s.Counter[metric.Name]
	if ok {
		result.Value = result.Value + metric.Value
		log.Println(result.Value)
		s.Counter[metric.Name] = result
	} else {
		s.Counter[metric.Name] = metric
	}
}
