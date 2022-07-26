package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/metricsAndAlerting/internal/models"
)

const (
	BaseURL = "http://127.0.0.1:8080"
)

type Client struct {
	baseURL    string
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: BaseURL,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) SendMetricByPath(params models.AgentMetrics) error {
	var value string
	if strings.ToLower(params.MType) == "gauge" {
		value = strconv.FormatFloat(params.Value, 'f', 6, 64)
	} else {
		value = strconv.FormatInt(params.Delta, 10)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/%s/%s/%v", c.baseURL, params.MType, params.ID, value), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if err := c.sendRequest(req); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendMetrics(metrics models.AgentMetrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if err := c.sendRequest(req); err != nil {
		return err
	}

	return nil
}

func (c *Client) sendRequest(req *http.Request) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil

}
