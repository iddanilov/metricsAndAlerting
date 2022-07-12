package models

import (
	"math/rand"
	"runtime"
)

type GaugeMetric struct {
	Name       string
	MetricType string
	Value      float64
}

func (gm GaugeMetric) GaugeMetricISEmpty() bool {
	return gm.Name == ""
}

type CountMetric struct {
	Name       string
	MetricType string
	Value      int64
}

func (cm CountMetric) CountMetricISEmpty() bool {
	return cm.Name == ""
}

type RuntimeMetric struct {
	Alloc         GaugeMetric
	BuckHashSys   GaugeMetric
	Frees         GaugeMetric
	GCCPUFraction GaugeMetric
	GCSys         GaugeMetric
	HeapAlloc     GaugeMetric
	HeapIdle      GaugeMetric
	HeapInuse     GaugeMetric
	HeapObjects   GaugeMetric
	HeapReleased  GaugeMetric
	HeapSys       GaugeMetric
	LastGC        GaugeMetric
	Lookups       GaugeMetric
	MCacheInuse   GaugeMetric
	MCacheSys     GaugeMetric
	MSpanInuse    GaugeMetric
	MSpanSys      GaugeMetric
	Mallocs       GaugeMetric
	NextGC        GaugeMetric
	NumForcedGC   GaugeMetric
	NumGC         GaugeMetric
	OtherSys      GaugeMetric
	PauseTotalNs  GaugeMetric
	StackInuse    GaugeMetric
	StackSys      GaugeMetric
	Sys           GaugeMetric
	TotalAlloc    GaugeMetric
	RandomValue   GaugeMetric
	PollCount     CountMetric
}

func (d *RuntimeMetric) SetPollCountMetricValue() CountMetric {
	d.PollCount.Value++
	return CountMetric{
		Name:       "PollCount",
		MetricType: "counter",
		Value:      d.PollCount.Value,
	}
}

func (d *RuntimeMetric) SetMetricValue(runtimeStat runtime.MemStats) []GaugeMetric {
	metric := make([]GaugeMetric, 25)
	metric = append(metric, GaugeMetric{
		Name:       "Alloc",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.Alloc),
	})
	metric = append(metric, GaugeMetric{
		Name:       "BuckHashSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.BuckHashSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "Frees",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.Frees),
	})
	metric = append(metric, GaugeMetric{
		Name:       "GCCPUFraction",
		MetricType: "Gauge",
		Value:      runtimeStat.GCCPUFraction,
	})
	metric = append(metric, GaugeMetric{
		Name:       "GCSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.GCSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "HeapAlloc",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.HeapAlloc),
	})
	metric = append(metric, GaugeMetric{
		Name:       "HeapIdle",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.HeapIdle),
	})
	metric = append(metric, GaugeMetric{
		Name:       "HeapInuse",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.HeapInuse),
	})
	metric = append(metric, GaugeMetric{
		Name:       "HeapObjects",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.HeapObjects),
	})
	metric = append(metric, GaugeMetric{
		Name:       "HeapReleased",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.HeapReleased),
	})
	metric = append(metric, GaugeMetric{
		Name:       "HeapSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.HeapSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "LastGC",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.LastGC),
	})
	metric = append(metric, GaugeMetric{
		Name:       "Lookups",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.Lookups),
	})
	metric = append(metric, GaugeMetric{
		Name:       "MCacheInuse",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.MCacheInuse),
	})
	metric = append(metric, GaugeMetric{
		Name:       "MCacheSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.MCacheSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "MSpanInuse",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.MSpanInuse),
	})
	metric = append(metric, GaugeMetric{
		Name:       "MSpanSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.MSpanSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "Mallocs",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.Mallocs),
	})
	metric = append(metric, GaugeMetric{
		Name:       "NextGC",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.NextGC),
	})
	metric = append(metric, GaugeMetric{
		Name:       "NumForcedGC",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.NumForcedGC),
	})
	metric = append(metric, GaugeMetric{
		Name:       "NumGC",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.NumGC),
	})
	metric = append(metric, GaugeMetric{
		Name:       "OtherSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.OtherSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "PauseTotalNs",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.PauseTotalNs),
	})
	metric = append(metric, GaugeMetric{
		Name:       "StackInuse",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.StackInuse),
	})
	metric = append(metric, GaugeMetric{
		Name:       "StackSys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.StackSys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "Sys",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.Sys),
	})
	metric = append(metric, GaugeMetric{
		Name:       "TotalAlloc",
		MetricType: "Gauge",
		Value:      float64(runtimeStat.TotalAlloc),
	})
	metric = append(metric, GaugeMetric{
		Name:       "RandomValue",
		MetricType: "Gauge",
		Value:      rand.Float64(),
	})

	return metric
}
