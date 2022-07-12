package server

import (
	client "github.com/metricsAndAlerting/internal/models"
	"sync"
)

type Storage struct {
	Gauge   map[string]client.GaugeMetric
	Counter map[string]client.CountMetric
}

func (s *Storage) SaveGaugeMetric(metric client.GaugeMetric, mu *sync.Mutex) {
	mu.Lock()
	s.Gauge[metric.Name] = metric
	mu.Unlock()
}

func (s *Storage) SaveCountMetric(metric client.CountMetric, mu *sync.Mutex) {
	mu.Lock()
	s.Counter[metric.Name] = metric
	mu.Unlock()
}
