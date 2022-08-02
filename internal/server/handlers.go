package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/metricsAndAlerting/internal/middleware"
	client "github.com/metricsAndAlerting/internal/models"
)

const (
	updateByPath = "/update/:type/:name/:value"
	update       = "/update/"
	value        = "/value/"
	valueByPath  = "/value/:type/:name"
	metricsName  = "/"
)

type routerGroup struct {
	rg *gin.RouterGroup
	s  *Storage
}

func NewRouterGroup(rg *gin.RouterGroup, s *Storage) *routerGroup {
	return &routerGroup{
		rg: rg,
		s:  s,
	}
}

func (h *routerGroup) Routes() {
	h.rg.POST(updateByPath, middleware.Middleware(h.UpdateMetricsByPath))
	h.rg.POST(update, middleware.Middleware(h.UpdateMetrics))
	h.rg.GET(value, middleware.Middleware(h.GetMetricByName))
	h.rg.POST(value, middleware.Middleware(h.GetMetric))
	h.rg.GET(valueByPath, middleware.Middleware(h.GetMetricByName))
	h.rg.GET(metricsName, middleware.Middleware(h.GetMetricsName))
}

func (h *routerGroup) GetMetric(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("Get Metrics", r.URL)
	log.Println("Get Metrics", r.Body)
	requestBody := client.Metrics{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	log.Println("Get Metrics", requestBody)

	if requestBody.ID == "" {
		return nil, middleware.ErrNotFound
	}
	response, ok := h.s.Metrics[requestBody.ID]
	if !ok {
		return nil, middleware.ErrNotFound
	}
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return body, err
}

func (h *routerGroup) GetMetricByName(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("Get Metrics", r.URL)
	var response []byte
	urlValues := strings.Split(r.URL.Path, "/")
	if len(urlValues) < 4 {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	if strings.ToLower(urlValues[2]) == "gauge" {
		result, ok := h.s.Metrics[urlValues[3]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", *result.Value))
		w.WriteHeader(http.StatusOK)
	} else if strings.ToLower(urlValues[2]) == "counter" {
		result, ok := h.s.Metrics[urlValues[3]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", *result.Delta))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	return response, nil
}

func (h *routerGroup) GetMetricsName(c *gin.Context) ([]byte, error) {
	return []byte(createResponse(h.s)), nil
}

func (h *routerGroup) UpdateMetricsByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetricsByPath Metrics", r.URL)
	urlValue := strings.Split(r.URL.Path, "/")
	if len(urlValue) < 5 {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	if strings.ToLower(urlValue[2]) == "gauge" {
		v, err := strconv.ParseFloat(urlValue[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", urlValue[3]))
		}
		h.s.SaveGaugeMetric(&client.Metrics{
			ID:    urlValue[3],
			MType: "Gauge",
			Value: &v,
		})
	} else if strings.ToLower(urlValue[2]) == "counter" {
		v, err := strconv.ParseInt(urlValue[4], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", urlValue[3]))
		}
		h.s.SaveCountMetric(client.Metrics{
			ID:    urlValue[3],
			MType: "Counter",
			Delta: &v,
		})
	} else {
		return nil, middleware.UnknownMetricName
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte{})
	if err != nil {
		log.Println("Write err: ", err.Error())
		return nil, err
	}

	return nil, nil
}

func (h *routerGroup) UpdateMetrics(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetrics Metrics", r.URL)

	requestBody := client.Metrics{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	if strings.ToLower(requestBody.MType) == "gauge" {
		h.s.SaveGaugeMetric(&client.Metrics{
			ID:    requestBody.ID,
			MType: requestBody.MType,
			Value: requestBody.Value,
		})
	} else if strings.ToLower(requestBody.MType) == "counter" {
		h.s.SaveCountMetric(client.Metrics{
			ID:    requestBody.ID,
			MType: "Counter",
			Delta: requestBody.Delta,
		})
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return nil, middleware.ErrNotFound
	}
	w.WriteHeader(http.StatusOK)

	return nil, nil
}

func createResponse(s *Storage) string {
	baseHTML := `<h1><ul>`
	finish := "</ul></h1>"
	for _, gmetric := range s.Metrics {
		baseHTML = baseHTML + fmt.Sprintf("<li>%s</li>", gmetric.ID)
	}
	baseHTML = baseHTML + finish

	return baseHTML
}
