package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	client "github.com/metricsAndAlerting/internal/agent"
	"github.com/metricsAndAlerting/internal/handlers"
	"github.com/metricsAndAlerting/internal/middleware"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	update = "/update/:type/:name/:value/"
	value  = "/value/:type/:name/"
)

type RequestBody struct {
	URL string `json:"url"`
}

type handler struct {
	storage Storage
}

func NewHandler(storage Storage) handlers.Handler {
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
	if len(urlValue) < 3 {
		w.WriteHeader(404)
		return middleware.ErrNotFound
	}

	if urlValue[1] == "Gauge" {
		v, err := strconv.ParseFloat(urlValue[3], 8)
		if err != nil {
			w.WriteHeader(500)
			return middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", urlValue[3]), err.Error())
		}
		h.storage.SetGaugeMetric(client.GaugeMetric{
			Name:       urlValue[2],
			MetricType: urlValue[1],
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
