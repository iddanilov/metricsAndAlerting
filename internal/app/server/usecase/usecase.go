package usecase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	serverApp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	client "github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/middleware"
)

type Config struct {
}

type serverUseCase struct {
	repository serverApp.Repository
	storage    serverApp.Storage
	logger     logger.Logger
	useDB      bool
	key        string
}

func NewServerUseCase(
	serverRepo serverApp.Repository,
	storage serverApp.Storage,
	logger logger.Logger,
	useDB bool,
	key string) serverApp.Usecase {
	return &serverUseCase{
		repository: serverRepo,
		logger:     logger,
		useDB:      useDB,
		storage:    storage,
		key:        key,
	}
}

// Ping - check db working.
func (u *serverUseCase) Ping(c *gin.Context) ([]byte, error) {
	u.logger.Info("Ping")
	if err := u.repository.Ping(); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return nil, nil
}

func (u *serverUseCase) GetMetric(c *gin.Context) ([]byte, error) {
	var hashValue string
	var err error
	r := c.Request
	w := c.Writer

	requestBody := client.Metrics{}
	responseBody := client.Metrics{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.logger.Error(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil, err
	}

	if requestBody.ID == "" {
		return nil, middleware.ErrNotFound
	}
	if u.useDB {
		responseBody, err = u.repository.GetMetric(c, requestBody.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil, err
		}
	} else {
		responseBody, err = u.storage.GetMetric(requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil, err
		}

	}
	if u.key != "" {
		if strings.ToLower(requestBody.MType) == "gauge" {
			hashValue, err = hashCreate(fmt.Sprintf("%s:gauge:%f", responseBody.ID, *responseBody.Value), []byte(u.key))
		} else {
			hashValue, err = hashCreate(fmt.Sprintf("%s:counter:%d", responseBody.ID, *responseBody.Delta), []byte(u.key))
		}
		if err != nil {
			u.logger.Error(err)
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
	u.logger.Info("Get Metrics", string(body))

	return body, err

}

func (u *serverUseCase) GetMetricByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	var err error
	u.logger.Info("Get Metrics", r.URL)
	u.logger.Info("Metrics Body: ", r.Body)

	mType := c.Params.ByName("type")
	name := c.Params.ByName("name")
	if mType == "" || name == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	var response []byte
	if strings.ToLower(mType) == "gauge" {
		var result *float64
		if u.useDB {
			result, err = u.repository.GetGaugeMetric(c, name)
			if err != nil {
				u.logger.Error(err)
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
		} else {
			result, err = u.storage.GetMetricValue(name)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
		}

		response = []byte(fmt.Sprintf("%v", *result))
		w.WriteHeader(http.StatusOK)
	} else if strings.ToLower(mType) == "counter" {
		var result *int64
		if u.useDB {
			result, err = u.repository.GetCounterMetric(c, name)
			if err != nil {
				u.logger.Error(err)
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
		} else {
			result, err = u.storage.GetMetricDelta(name)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return nil, middleware.ErrNotFound
			}
		}

		response = []byte(fmt.Sprintf("%v", *result))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	return response, nil
}

func (u *serverUseCase) MetricList(c *gin.Context) ([]byte, error) {
	c.Writer.Header().Set("Content-Type", "text/html")
	var values []string
	var err error
	if u.useDB {
		values, err = u.repository.GetMetricNames(c)
		if err != nil {
			return nil, err
		}
	} else {
		values, err = u.storage.GetMetricsList()
		if err != nil {
			return nil, err
		}
	}
	return []byte(createResponse(values)), nil
}

func (u *serverUseCase) UpdateMetricByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	u.logger.Info("UpdateMetricByPath Metrics", r.URL)
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
		if u.useDB {
			err = u.repository.UpdateMetric(c, client.Metrics{
				ID:    name,
				MType: mType,
				Delta: nil,
				Value: &v,
			})
		} else {
			u.storage.SaveGaugeMetric(&client.Metrics{
				ID:    name,
				MType: mType,
				Value: &v,
			})
		}
		if err != nil {
			u.logger.Error(err)
		}
	} else if strings.ToLower(mType) == "counter" {
		v, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", mValue))
		}

		if u.useDB {
			err = u.repository.UpdateMetric(c, client.Metrics{
				ID:    name,
				MType: mType,
				Delta: &v,
				Value: nil,
			})
		} else {
			u.storage.SaveCountMetric(client.Metrics{
				ID:    name,
				MType: mType,
				Delta: &v,
			})
		}
		if err != nil {
			u.logger.Error(err)
		}
	} else {
		return nil, middleware.UnknownMetricName
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte{})
	if err != nil {
		u.logger.Error("Write err: ", err.Error())
		return nil, err
	}

	return nil, nil
}

func (u *serverUseCase) UpdateMetric(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	requestBody := client.Metrics{}
	u.logger.Info("UpdateMetric Metrics", r.URL)

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return nil, err
	}

	if strings.ToLower(requestBody.MType) == "gauge" {
		if requestBody.Value == nil {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}
		if u.key != "" && requestBody.Hash != "" {
			ok, err := hash(requestBody.Hash, fmt.Sprintf("%s:gauge:%f", requestBody.ID, *requestBody.Value), []byte(u.key))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
		}
		if u.useDB {
			err := u.repository.UpdateMetric(c, client.Metrics{
				ID:    requestBody.ID,
				MType: strings.ToLower(requestBody.MType),
				Delta: nil,
				Value: requestBody.Value,
			})
			if err != nil {
				u.logger.Error(err)
			}
		} else {
			u.storage.SaveGaugeMetric(&client.Metrics{
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
		if u.key != "" && requestBody.Hash != "" {
			ok, err := hash(requestBody.Hash, fmt.Sprintf("%s:counter:%d", requestBody.ID, *requestBody.Delta), []byte(u.key))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
		}

		if u.useDB {
			err := u.repository.UpdateMetric(c, client.Metrics{
				ID:    requestBody.ID,
				MType: strings.ToLower(requestBody.MType),
				Delta: requestBody.Delta,
				Value: nil,
			})
			if err != nil {
				u.logger.Error(err)
			}
		} else {
			u.storage.SaveCountMetric(client.Metrics{
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

func (u *serverUseCase) UpdateMetrics(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	u.logger.Info("UpdateMetric Metrics", r.URL)
	u.logger.Info("Metrics Body: ", r.Body)

	var requestBody []client.Metrics

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	if u.useDB {
		err := u.repository.UpdateMetrics(requestBody)
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
	for _, metric := range metrics {
		baseHTML = baseHTML + fmt.Sprintf("<li>%s</li>", metric)
	}
	baseHTML = baseHTML + finish

	return baseHTML
}
