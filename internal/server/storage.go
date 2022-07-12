package server

import (
	client "github.com/metricsAndAlerting/internal/models"
)

type Storage struct {
	client.RuntimeMetric
}

func (s *Storage) SaveGaugeMetric(metric client.GaugeMetric) {
	switch metric.Name {
	case "Alloc":
		s.Alloc = metric
	case "BuckHashSys":
		s.BuckHashSys = metric
	case "Frees":
		s.Frees = metric
	case "GCCPUFraction":
		s.GCCPUFraction = metric
	case "GCSys":
		s.GCSys = metric
	case "HeapAlloc":
		s.HeapAlloc = metric
	case "HeapIdle":
		s.HeapIdle = metric
	case "HeapInuse":
		s.HeapInuse = metric
	case "HeapObjects":
		s.HeapObjects = metric
	case "HeapReleased":
		s.HeapReleased = metric
	case "HeapSys":
		s.HeapSys = metric
	case "LastGC":
		s.LastGC = metric
	case "Lookups":
		s.Lookups = metric
	case "MCacheInuse":
		s.MCacheInuse = metric
	case "MCacheSys":
		s.MCacheSys = metric
	case "MSpanInuse":
		s.MSpanInuse = metric
	case "MSpanSys":
		s.MSpanSys = metric
	case "Mallocs":
		s.Mallocs = metric
	case "NextGC":
		s.NextGC = metric
	case "NumForcedGC":
		s.NumForcedGC = metric
	case "NumGC":
		s.NumGC = metric
	case "OtherSys":
		s.OtherSys = metric
	case "PauseTotalNs":
		s.PauseTotalNs = metric
	case "StackInuse":
		s.StackInuse = metric
	case "StackSys":
		s.StackSys = metric
	case "Sys":
		s.Sys = metric
	case "TotalAlloc":
		s.TotalAlloc = metric
	case "RandomValue":
		s.RandomValue = metric

	default:
	}
}

func (s *Storage) SaveCountMetric(metric client.CountMetric) {
	s.PollCount = metric
}
