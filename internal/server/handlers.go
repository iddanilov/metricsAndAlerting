// Package server is a server handler and storage.
package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/iddanilov/metricsAndAlerting/internal/db"
	"github.com/iddanilov/metricsAndAlerting/internal/middleware"
	client "github.com/iddanilov/metricsAndAlerting/internal/models"
)

type RouterGroup struct {
	rg            *gin.RouterGroup
	s             *Storage
	key           string
	db            *db.DB
	useDB         bool
	trustedSubnet string
}

// NewRouterGroup - create new gin route group
func NewRouterGroup(rg *gin.RouterGroup, s *Storage, key string, db *db.DB, useDB bool, trustedSubnet string) *RouterGroup {
	return &RouterGroup{
		rg:            rg,
		s:             s,
		key:           key,
		db:            db,
		useDB:         useDB,
		trustedSubnet: trustedSubnet,
	}
}

func (h *RouterGroup) Routes() {
	group := h.rg.Group("/")
	group.Use()
	{
		group.GET("/", middleware.Middleware(h.MetricList, h.trustedSubnet))
		group.POST("/update/:type/:name/:value", middleware.Middleware(h.UpdateMetricByPath, h.trustedSubnet))
		group.POST("/update/", middleware.Middleware(h.UpdateMetric, h.trustedSubnet))
		group.POST("/updates/", middleware.Middleware(h.UpdateMetrics, h.trustedSubnet))
		group.POST("/value/", middleware.Middleware(h.GetMetric, h.trustedSubnet))
		group.GET("/value/:type/:name", middleware.Middleware(h.GetMetricByPath, h.trustedSubnet))
		group.GET("/ping", middleware.Middleware(h.Ping, h.trustedSubnet))
	}
}

// Ping - GET request for checking db working.
func (h *RouterGroup) Ping(c *gin.Context) ([]byte, error) {
	log.Println("Ping")
	if h.db.DB == nil {
		err := errors.New("can't connect to db")
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	if err := h.db.DBPing(c); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return nil, nil
}

// GetMetric - GET request for get metric by body value
func (h *RouterGroup) GetMetric(c *gin.Context) ([]byte, error) {
	var hashValue string
	var err error
	r := c.Request
	w := c.Writer

	requestBody := client.Metrics{}
	responseBody := client.Metrics{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil, err
	}

	if requestBody.ID == "" {
		return nil, middleware.ErrNotFound
	}
	if h.useDB {
		responseBody, err = h.db.GetMetric(c, requestBody.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil, err
		}
	} else {
		response, ok := h.s.Metrics[requestBody.ID]
		if !ok {
			return nil, middleware.ErrNotFound
		}
		response.MType = strings.ToLower(response.MType)
		responseBody = response
	}
	if h.key != "" {
		if strings.ToLower(requestBody.MType) == "gauge" {
			hashValue, err = hashCreate(fmt.Sprintf("%s:gauge:%f", responseBody.ID, *responseBody.Value), []byte(h.key))
		} else {
			hashValue, err = hashCreate(fmt.Sprintf("%s:counter:%d", responseBody.ID, *responseBody.Delta), []byte(h.key))
		}
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil, err
		}
		responseBody.Hash = hashValue
	}

	if !strings.EqualFold(responseBody.MType, requestBody.MType) {
		http.Error(w, "type is not correct", http.StatusNotFound)
		return nil, err
	}

	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(responseBody)
	if err != nil {
		return nil, err
	}
	log.Println("Get Metrics", string(body))

	return body, err

}

// GetMetricByPath - GET request for get metric by url value
func (h *RouterGroup) GetMetricByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	var err error
	log.Println("Get Metrics", r.URL)
	log.Println("Metrics Body: ", r.Body)
	mType := c.Params.ByName("type")
	name := c.Params.ByName("name")
	if mType == "" || name == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	var response []byte
	if strings.ToLower(mType) == "gauge" {
		var result *float64
		if h.useDB {
			result, err = h.db.GetGaugeMetric(c, name)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
		} else {
			metric, ok := h.s.Metrics[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
			if metric.Value == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
			result = metric.Value
		}

		response = []byte(fmt.Sprintf("%v", *result))
		w.WriteHeader(http.StatusOK)
	} else if strings.ToLower(mType) == "counter" {
		var result *int64
		if h.useDB {
			result, err = h.db.GetCounterMetric(c, name)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
		} else {
			metric, ok := h.s.Metrics[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
			if metric.Delta == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
			result = metric.Delta
		}

		response = []byte(fmt.Sprintf("%v", *result))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	return response, nil
}

// MetricList - GET request for get all metrics
func (h *RouterGroup) MetricList(c *gin.Context) ([]byte, error) {
	c.Writer.Header().Set("Content-Type", "text/html")
	var values []string
	var err error
	if h.useDB {
		values, err = h.db.GetMetricNames(c)
		if err != nil {
			return nil, err
		}
	} else {
		for _, m := range h.s.Metrics {
			values = append(values, m.ID)
		}

	}
	return []byte(createResponse(values)), nil
}

// UpdateMetricByPath - GET request for update metric by url value
func (h *RouterGroup) UpdateMetricByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetricByPath Metrics", r.URL)
	mType := c.Params.ByName("type")
	name := c.Params.ByName("name")
	mValue := c.Params.ByName("value")
	if mType == "" || name == "" || mValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	if strings.ToLower(mType) == "gauge" {
		v, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", mType))
		}
		if h.useDB {
			err = h.db.UpdateMetric(c, client.Metrics{
				ID:    name,
				MType: mType,
				Delta: nil,
				Value: &v,
			})
		} else {
			h.s.SaveGaugeMetric(&client.Metrics{
				ID:    name,
				MType: mType,
				Value: &v,
			})
		}
		if err != nil {
			log.Println(err)
		}
	} else if strings.ToLower(mType) == "counter" {
		v, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", mValue))
		}

		if h.useDB {
			err = h.db.UpdateMetric(c, client.Metrics{
				ID:    name,
				MType: mType,
				Delta: &v,
				Value: nil,
			})
		} else {
			h.s.SaveCountMetric(client.Metrics{
				ID:    name,
				MType: mType,
				Delta: &v,
			})
		}
		if err != nil {
			log.Println(err)
		}
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

// UpdateMetric - POST request for update metric by body value
func (h *RouterGroup) UpdateMetric(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	requestBody := client.Metrics{}
	log.Println("UpdateMetric Metrics", r.URL)

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return nil, err
	}

	if strings.ToLower(requestBody.MType) == "gauge" {
		if requestBody.Value == nil {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}
		if h.key != "" && requestBody.Hash != "" {
			ok, err := hash(requestBody.Hash, fmt.Sprintf("%s:gauge:%f", requestBody.ID, *requestBody.Value), []byte(h.key))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
		}
		if h.useDB {
			err := h.db.UpdateMetric(c, client.Metrics{
				ID:    requestBody.ID,
				MType: strings.ToLower(requestBody.MType),
				Delta: nil,
				Value: requestBody.Value,
			})
			if err != nil {
				log.Println(err)
			}
		} else {
			h.s.SaveGaugeMetric(&client.Metrics{
				ID:    requestBody.ID,
				MType: strings.ToLower(requestBody.MType),
				Value: requestBody.Value,
			})
		}

	} else if strings.ToLower(requestBody.MType) == "counter" {
		if requestBody.Delta == nil {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}
		if h.key != "" && requestBody.Hash != "" {
			ok, err := hash(requestBody.Hash, fmt.Sprintf("%s:counter:%d", requestBody.ID, *requestBody.Delta), []byte(h.key))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
		}

		if h.useDB {
			err := h.db.UpdateMetric(c, client.Metrics{
				ID:    requestBody.ID,
				MType: strings.ToLower(requestBody.MType),
				Delta: requestBody.Delta,
				Value: nil,
			})
			if err != nil {
				log.Println(err)
			}
		} else {
			h.s.SaveCountMetric(client.Metrics{
				ID:    requestBody.ID,
				MType: strings.ToLower(requestBody.MType),
				Delta: requestBody.Delta,
			})
		}
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return nil, middleware.ErrNotFound
	}
	w.WriteHeader(http.StatusOK)

	return nil, nil
}

// UpdateMetrics - POST request for update all metrics in body by body value
func (h *RouterGroup) UpdateMetrics(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetric Metrics", r.URL)
	log.Println("Metrics Body: ", r.Body)

	var requestBody []client.Metrics

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	if h.useDB {
		err := h.db.UpdateMetrics(requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	w.WriteHeader(http.StatusOK)

	return nil, nil
}

func hashCreate(m string, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(m))
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func hash(bodyHash string, m string, key []byte) (bool, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(m))
	if err != nil {
		return false, err
	}
	if bytes.Equal([]byte(fmt.Sprintf("%x", h.Sum(nil))), []byte(bodyHash)) {
		fmt.Println("Всё правильно! Хеши равны")
		return true, nil
	} else {
		fmt.Println("Что-то пошло не так")
		return false, nil
	}
}

func createResponse(metrics []string) string {
	baseHTML := `<h1><ul>`
	finish := "</ul></h1>"
	for _, gmetric := range metrics {
		baseHTML = baseHTML + fmt.Sprintf("<li>%s</li>", gmetric)
	}
	baseHTML = baseHTML + finish

	return baseHTML
}
