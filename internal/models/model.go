package models

import (
	"log"
	"math/rand"
	"reflect"
	"runtime"
)

type Counter int64

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type AgentMetrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (am AgentMetrics) MetricISEmpty() bool {
	return am.ID == ""
}
func (m Metrics) MetricISEmpty() bool {
	return m.ID == ""
}

var gaugeMetric = [...]string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
	"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
	"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC",
	"NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
	"StackSys", "Sys", "TotalAlloc", "RandomValue",
}

func (*Metrics) SetMetrics(runtimeStat *runtime.MemStats) []Metrics {
	var result []Metrics
	v := reflect.ValueOf(*runtimeStat)
	for _, s := range gaugeMetric {
		var value float64
		if s == "RandomValue" {
			value = rand.Float64()
		} else {
			valueType := v.FieldByName(s).Type()
			if valueType.Name() == "uint64" || valueType.Name() == "uint32" {
				value = float64(v.FieldByName(s).Uint())
			} else if valueType.Name() == "float64" {
				value = v.FieldByName(s).Float()
			}
		}
		result = append(result, Metrics{
			ID:    s,
			MType: "Gauge",
			Value: &value,
		})
	}
	return result
}

func (c *Counter) SetPollCountMetricValue() []Metrics {
	*c++
	value := int64(*c)
	log.Println(*c)
	return []Metrics{{
		ID:    "PollCount",
		MType: "Counter",
		Delta: &value,
	},
	}
}
