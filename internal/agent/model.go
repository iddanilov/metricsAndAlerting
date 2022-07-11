package client

import "runtime"

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
	metric := make([]GaugeMetric, 25, 25)
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

	return metric
}
