package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

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
	mu      *sync.Mutex
}

func NewHandler(storage *Storage, mu *sync.Mutex) handlers.Handler {
	return &handler{
		storage: storage,
		mu:      mu,
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
		result, ok := h.storage.Gauge[urlValue[3]]
		if !ok {
			w.WriteHeader(404)
			return middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%f", result.Value))
		w.WriteHeader(200)
	} else if strings.ToLower(urlValue[2]) == "counter" {
		result, ok := h.storage.Counter[urlValue[3]]
		if !ok {
			w.WriteHeader(404)
			return middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", result.Value))
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
	w.Write([]byte(CreateResponse(h.storage)))
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
			w.WriteHeader(400)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", urlValue[3]), err.Error())
		}
		h.storage.SaveGaugeMetric(client.GaugeMetric{
			Name:       urlValue[3],
			MetricType: "Gauge",
			Value:      v,
		}, h.mu)
	} else if strings.ToLower(urlValue[2]) == "counter" {
		v, err := strconv.ParseInt(urlValue[4], 10, 64)
		if err != nil {
			w.WriteHeader(400)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", urlValue[3]), err.Error())
		}
		h.storage.SaveCountMetric(client.CountMetric{
			Name:       urlValue[3],
			MetricType: "Counter",
			Value:      v,
		}, h.mu)
	} else {
		w.WriteHeader(501)
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

func CreateResponse(s *Storage) string {
	baseHTML := `<h1><ul>`
	finish := "</ul></h1>"
	for _, gmetric := range s.Gauge {
		baseHTML = baseHTML + fmt.Sprintf("<li>%s</li>", gmetric.Name)
	}
	for _, cmetric := range s.Counter {
		baseHTML = baseHTML + fmt.Sprintf("<li>%s</li>", cmetric.Name)
	}
	baseHTML = baseHTML + finish

	return baseHTML
}
