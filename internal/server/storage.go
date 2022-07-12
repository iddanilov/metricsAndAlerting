package server

import (
	client "github.com/metricsAndAlerting/internal/models"
	"sync"
)

type Storage struct {
	Gauge   map[string]client.GaugeMetric
	Counter map[string]client.CountMetric
}

func (s *Storage) SaveGaugeMetric(metric client.GaugeMetric, m *sync.Mutex) {
	s.Gauge[metric.Name] = metric
}

func (s *Storage) SaveCountMetric(metric client.CountMetric, m *sync.Mutex) {
	s.Counter[metric.Name] = metric
}
