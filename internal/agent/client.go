package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/metricsAndAlerting/internal/models"
)

const (
	BaseURL = "http://127.0.0.1:8080"
)

type Config struct {
	ADDRESS        string `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
}

type Client struct {
	HTTPClient *http.Client
	Config     Config
}

func NewClient() *Client {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil
	}
	return &Client{
		Config: cfg,
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/%s/%s/%v", c.Config.ADDRESS, params.MType, params.ID, value), nil)
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", c.Config.ADDRESS), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
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
