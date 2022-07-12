package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/metricsAndAlerting/internal/models"
)

const (
	BaseURL = "http://127.0.0.1:8080"
)

type Client struct {
	baseURL    string
	apiKey     string
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

func (c *Client) SendMetrics(params models.GaugeMetric) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/%s/%s/%v", c.baseURL, params.MetricType, params.Name, params.Value), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if err := c.sendRequest(req); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendPollCountMetric(params models.CountMetric) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/%s/%s/%v", c.baseURL, params.MetricType, params.Name, params.Value), nil)
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
	_, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}
