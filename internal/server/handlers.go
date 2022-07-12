package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/metricsAndAlerting/internal/handlers"
	"github.com/metricsAndAlerting/internal/middleware"
	client "github.com/metricsAndAlerting/internal/models"
)

const (
	update      = "/update/:type/:name/:value"
	value       = "/value/:type/:name"
	metricsName = "/"
)

type RequestBody struct {
	URL string `json:"url"`
}

type handler struct {
	storage *Storage
}

func NewHandler(storage *Storage) handlers.Handler {
	return &handler{
		storage: storage,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, update, middleware.Middleware(h.UpdateMetrics))
	router.HandlerFunc(http.MethodGet, value, middleware.Middleware(h.GetMetricByName))
	router.HandlerFunc(http.MethodGet, metricsName, middleware.Middleware(h.GetMetricsName))
}

func (h *handler) GetMetricByName(w http.ResponseWriter, r *http.Request) error {
	log.Println("Get Metrics", r.URL)
	var response []byte
	urlValue := strings.Split(r.URL.Path, "/")
	if len(urlValue) < 5 {
		w.WriteHeader(404)
		return middleware.ErrNotFound
	}

	if strings.ToLower(urlValue[2]) == "gauge" {
		var mapStorage map[string]client.GaugeMetric
		jsonMetrics, _ := json.Marshal(h.storage)
		json.Unmarshal(jsonMetrics, &mapStorage)
		_, ok := mapStorage[urlValue[3]]
		if !ok {
			w.WriteHeader(404)
			return middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%f", mapStorage[urlValue[3]].Value))
		w.WriteHeader(200)
	} else if strings.ToLower(urlValue[2]) == "counter" {
		var mapStorage map[string]client.CountMetric
		jsonMetrics, _ := json.Marshal(h.storage)
		json.Unmarshal(jsonMetrics, &mapStorage)
		_, ok := mapStorage[urlValue[3]]
		if !ok {
			w.WriteHeader(404)
			return middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", mapStorage[urlValue[3]].Value))
		w.WriteHeader(200)
	} else {
		w.WriteHeader(404)
		return middleware.ErrNotFound
	}

	w.Write(response)

	return nil
}

func (h *handler) GetMetricsName(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(200)
	w.Write([]byte(CreateResponse()))
	return nil
}

func (h *handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) error {
	log.Println("UpdateMetrics Metrics", r.URL)
	urlValue := strings.Split(r.URL.Path, "/")
	if len(urlValue) < 5 {
		w.WriteHeader(404)
		return middleware.ErrNotFound
	}

	if strings.ToLower(urlValue[2]) == "gauge" {
		v, err := strconv.ParseFloat(urlValue[4], 64)
		if err != nil {
			w.WriteHeader(500)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", urlValue[3]), err.Error())
		}
		h.storage.SaveGaugeMetric(client.GaugeMetric{
			Name:       urlValue[3],
			MetricType: "Gauge",
			Value:      v,
		})
	} else if strings.ToLower(urlValue[2]) == "counter" {
		v, err := strconv.ParseInt(urlValue[4], 10, 64)
		if err != nil {
			w.WriteHeader(500)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", urlValue[3]), err.Error())
		}
		h.storage.SaveCountMetric(client.CountMetric{
			Name:       urlValue[3],
			MetricType: "Counter",
			Value:      v,
		})
	} else {
		w.WriteHeader(404)
		return middleware.ErrNotFound
	}

	w.WriteHeader(200)
	_, err := w.Write([]byte{})
	if err != nil {
		log.Println("Write err: ", err.Error())
		return err
	}

	return nil
}

func CreateResponse() string {
	baseHTML := `
	<h1>
		<ul>
			<li>Alloc</li>
			<li>BuckHashSys</li>
			<li>Frees</li>
			<li>GCCPUFraction</li>
			<li>GCSys</li>
			<li>HeapAlloc</li>
			<li>HeapIdle</li>
			<li>HeapInuse</li>
			<li>HeapObjects</li>
			<li>HeapReleased</li>
			<li>HeapSys</li>
			<li>LastGC</li>
			<li>Lookups</li>
			<li>MCacheInuse</li>
			<li>MCacheSys</li>
			<li>MSpanInuse</li>
			<li>MSpanSys</li>
			<li>Mallocs</li>
			<li>NextGC</li>
			<li>NumForcedGC</li>
			<li>NumGC</li>
			<li>OtherSys</li>
			<li>PauseTotalNs</li>
			<li>StackInuse</li>
			<li>StackSys</li>
			<li>Sys</li>
			<li>TotalAlloc</li>
			<li>PollCount</li>
			<li>RandomValue</li>
		</ul>
	</h1>`
	return baseHTML
}
