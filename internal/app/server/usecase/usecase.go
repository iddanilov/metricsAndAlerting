package usecase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	serverapp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	client "github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/middleware"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Config struct {
}

type serverUsecase struct {
	config    models.Storage
	serverapp serverapp.Repository
	logger    logger.Logger
}

func NewServerUsecase(config models.Storage, noteRepo serverapp.Repository, logger logger.Logger) serverapp.Usecase {
	return &serverUsecase{
		config:    config,
		serverapp: noteRepo,
		logger:    logger,
	}
}

// Ping - check db working.
func (h *serverUsecase) Ping(c *gin.Context) ([]byte, error) {
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

func (h *serverUsecase) GetMetric(c *gin.Context) ([]byte, error) {
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

func (h *serverUsecase) GetMetricByPath(c *gin.Context) ([]byte, error) {
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

func (h *serverUsecase) MetricList(c *gin.Context) ([]byte, error) {
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

func (h *serverUsecase) UpdateMetricByPath(c *gin.Context) ([]byte, error) {
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

func (h *serverUsecase) UpdateMetric(c *gin.Context) ([]byte, error) {
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

func (h *serverUsecase) UpdateMetrics(c *gin.Context) ([]byte, error) {
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
