package server

import client "github.com/metricsAndAlerting/internal/agent"

type Storage struct {
	client.RuntimeMetric
}

func (s *Storage) SetGaugeMetric(metric client.GaugeMetric) {
	switch metric.Name {
	case "Alloc":
		s.Alloc = metric
	default:
	}
}

func (s *Storage) SetCountMetric(metric client.CountMetric) {
	s.PollCount = metric
}
