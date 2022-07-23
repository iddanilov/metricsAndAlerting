package server

import (
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
	urlValues := strings.Split(r.URL.Path, "/")
	if len(urlValues) < 4 {
		w.WriteHeader(http.StatusNotFound)
		return middleware.ErrNotFound
	}

	if strings.ToLower(urlValues[2]) == "gauge" {
		result, ok := h.storage.Gauge[urlValues[3]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", result.Value))
		w.WriteHeader(http.StatusOK)
	} else if strings.ToLower(urlValues[2]) == "counter" {
		result, ok := h.storage.Counter[urlValues[3]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", result.Value))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return middleware.ErrNotFound
	}

	_, err := w.Write(response)

	return err
}

func (h *handler) GetMetricsName(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(createResponse(h.storage)))
	return err
}

func (h *handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) error {
	log.Println("UpdateMetrics Metrics", r.URL)
	urlValue := strings.Split(r.URL.Path, "/")
	if len(urlValue) < 5 {
		w.WriteHeader(http.StatusNotFound)
		return middleware.ErrNotFound
	}

	if strings.ToLower(urlValue[2]) == "gauge" {
		v, err := strconv.ParseFloat(urlValue[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", urlValue[3]))
		}
		h.storage.SaveGaugeMetric(client.GaugeMetric{
			Name:       urlValue[3],
			MetricType: "Gauge",
			Value:      v,
		})
	} else if strings.ToLower(urlValue[2]) == "counter" {
		v, err := strconv.ParseInt(urlValue[4], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", urlValue[3]))
		}
		h.storage.SaveCountMetric(client.CountMetric{
			Name:       urlValue[3],
			MetricType: "Counter",
			Value:      v,
		})
	} else {
		w.WriteHeader(501)
		return middleware.ErrNotFound
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte{})
	if err != nil {
		log.Println("Write err: ", err.Error())
		return err
	}

	return nil
}

func createResponse(s *Storage) string {
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
