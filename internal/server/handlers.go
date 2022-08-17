package server

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/metricsAndAlerting/internal/db"
	"github.com/metricsAndAlerting/internal/middleware"
	client "github.com/metricsAndAlerting/internal/models"
)

type routerGroup struct {
	rg  *gin.RouterGroup
	s   *Storage
	key string
	db  *db.DB
}

func NewRouterGroup(rg *gin.RouterGroup, s *Storage, key string, db *db.DB) *routerGroup {
	return &routerGroup{
		rg:  rg,
		s:   s,
		key: key,
		db:  db,
	}
}

func (h *routerGroup) Routes() {
	group := h.rg.Group("/")
	group.Use()
	{
		group.GET("/", middleware.Middleware(h.MetricList))
		group.POST("/update/:type/:name/:value", middleware.Middleware(h.UpdateMetricByPath))
		group.POST("/update/", middleware.Middleware(h.UpdateMetric))
		group.POST("/updates/", middleware.Middleware(h.UpdateMetrics))
		group.POST("/value/", middleware.Middleware(h.GetMetric))
		group.GET("/value/:type/:name", middleware.Middleware(h.GetMetricByPath))
		group.GET("/ping/", middleware.Middleware(h.Ping))
	}
}

func (h *routerGroup) Ping(c *gin.Context) ([]byte, error) {
	if err := h.db.DBPing(c); err != nil {
		return nil, middleware.DisconnectDB
	}
	return nil, nil
}

func (h *routerGroup) GetMetric(c *gin.Context) ([]byte, error) {
	var hashValue string
	var err error
	r := c.Request
	w := c.Writer
	log.Println("Get Metrics", r.Body)
	requestBody := client.Metrics{}
	responseBody := client.Metrics{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	log.Println("Get Metrics", requestBody)

	if requestBody.ID == "" {
		return nil, middleware.ErrNotFound
	}
	if requestBody.Hash != "" {
		hashValue, err = hashCreate(fmt.Sprintf("%s:gauge:%f", requestBody.ID, *requestBody.Value), []byte(h.key))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, err
		}
		responseBody.Hash = hashValue
	}
	responseBody, err = h.db.GetMetric(c, requestBody.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

func (h *routerGroup) GetMetricByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("Get Metrics", r.URL)
	mType := c.Params.ByName("type")
	name := c.Params.ByName("name")
	if mType == "" || name == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	var response []byte
	if strings.ToLower(mType) == "gauge" {
		result, err := h.db.GetGaugeMetric(c, name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}

		response = []byte(fmt.Sprintf("%v", result))
		w.WriteHeader(http.StatusOK)
	} else if strings.ToLower(mType) == "counter" {
		result, err := h.db.GetCounterMetric(c, name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return nil, middleware.ErrNotFound
		}
		response = []byte(fmt.Sprintf("%v", result))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}

	return response, nil
}

func (h *routerGroup) MetricList(c *gin.Context) ([]byte, error) {
	c.Writer.Header().Set("Content-Type", "text/html")
	return []byte(createResponse(h.s)), nil
}

func (h *routerGroup) UpdateMetricByPath(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	mType := c.Params.ByName("type")
	name := c.Params.ByName("name")
	mValue := c.Params.ByName("value")
	if mType == "" || name == "" || mValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, middleware.ErrNotFound
	}
	log.Println("UpdateMetricByPath Metrics", r.URL)

	if strings.ToLower(mType) == "gauge" {
		v, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type float64: value%s", mType))
		}
		h.s.SaveGaugeMetric(&client.Metrics{
			ID:    name,
			MType: mType,
			Value: &v,
		})
		err = h.db.UpdateMetric(c, client.Metrics{
			ID:    name,
			MType: mType,
			Delta: nil,
			Value: &v,
		})
		if err != nil {
			log.Println(err)
		}
	} else if strings.ToLower(mType) == "counter" {
		v, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, middleware.NewAppError(nil, fmt.Sprintf("Value should be type int64: value%s", mValue))
		}
		h.s.SaveCountMetric(client.Metrics{
			ID:    name,
			MType: mType,
			Delta: &v,
		})
		err = h.db.UpdateMetric(c, client.Metrics{
			ID:    name,
			MType: mType,
			Delta: &v,
			Value: nil,
		})
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

func (h *routerGroup) UpdateMetric(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetric Metrics", r.URL)

	requestBody := client.Metrics{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	if strings.ToLower(requestBody.MType) == "gauge" {
		if requestBody.Hash != "" {
			ok, err := hash(requestBody.Hash, fmt.Sprintf("%s:gauge:%f", requestBody.ID, *requestBody.Value), []byte(h.key))
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if ok {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
		}
		h.s.SaveGaugeMetric(&client.Metrics{
			ID:    requestBody.ID,
			MType: strings.ToLower(requestBody.MType),
			Value: requestBody.Value,
		})
		err := h.db.UpdateMetric(c, client.Metrics{
			ID:    requestBody.ID,
			MType: strings.ToLower(requestBody.MType),
			Delta: nil,
			Value: requestBody.Value,
		})
		if err != nil {
			log.Println(err)
		}
	} else if strings.ToLower(requestBody.MType) == "counter" {
		if requestBody.Hash != "" {
			ok, err := hash(requestBody.Hash, fmt.Sprintf("%s:counter:%v", requestBody.ID, *requestBody.Delta), []byte(h.key))
			if err != nil {
				return nil, err
			}
			if ok {
				w.WriteHeader(http.StatusBadRequest)
				return nil, err
			}
		}
		h.s.SaveCountMetric(client.Metrics{
			ID:    requestBody.ID,
			MType: strings.ToLower(requestBody.MType),
			Delta: requestBody.Delta,
		})
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
		w.WriteHeader(http.StatusNotImplemented)
		return nil, middleware.ErrNotFound
	}
	w.WriteHeader(http.StatusOK)

	return nil, nil
}

func (h *routerGroup) UpdateMetrics(c *gin.Context) ([]byte, error) {
	r := c.Request
	w := c.Writer
	log.Println("UpdateMetric Metrics", r.URL)

	var requestBody []client.Metrics

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	err := h.db.UpdateMetrics(requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)

	return nil, nil
}

func hashCreate(m string, key []byte) (string, error) {
	src := []byte(m)

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]
	dst := aesgcm.Seal(nil, nonce, src, nil)
	log.Println("dst: ", dst)
	return hex.EncodeToString(dst), err
}

func hash(bodyHash string, m string, key []byte) (bool, error) {
	encrypted, err := hex.DecodeString(bodyHash)
	if err != nil {
		log.Printf("error: %v\n", err)
		return false, err
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("error: %v\n", err)
		return false, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Printf("error cipher:: %v\n", err)
		return false, err
	}
	nonce := key[len(key)-aesgcm.NonceSize():]

	src2, err := aesgcm.Open(nil, nonce, encrypted, nil) // расшифровываем
	if err != nil {
		log.Printf("error aesgcm: %v\n", err)
		return false, err
	}

	if bytes.Equal(src2, []byte(m)) {
		log.Println("-------- Хеши равны -------------")
		return true, nil
	}

	return false, err
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
