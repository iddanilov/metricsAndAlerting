package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/metricsAndAlerting/internal/handlers"
	"github.com/metricsAndAlerting/internal/middleware"
	client "github.com/metricsAndAlerting/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	update      = "/update/:type/:name/:value/"
	value       = "/value/:type/:name/"
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
}

func (h *handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) error {
	log.Println("UpdateMetrics Get Metrics", r.URL)
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
	}

	w.WriteHeader(200)
	_, err := w.Write([]byte{})
	if err != nil {
		log.Println("Write err: ", err.Error())
		return err
	}

	return nil
}
