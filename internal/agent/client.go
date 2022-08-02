package client

import (
	"bytes"
	"encoding/json"
	goflag "flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"

	"github.com/metricsAndAlerting/internal/models"
)

var (
	ADDRESS        *string        = flag.String("a", "http://127.0.0.1:8080", "help message for flagname")
	PollInterval   *time.Duration = flag.Duration("p", time.Duration(2*time.Second), "help message for flagname")
	ReportInterval *time.Duration = flag.Duration("r", time.Duration(10*time.Second), "help message for flagname")
)

type Config struct {
	ADDRESS        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

type Client struct {
	HTTPClient *http.Client
	Config     Config
}

func NewClient() *Client {
	var cfg Config
	err := env.Parse(&cfg)
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	if cfg.ADDRESS == "" {
		cfg.ADDRESS = *ADDRESS
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *ReportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *PollInterval
	}

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

func (c *Client) SendMetricByPath(params models.Metrics) error {
	var value string
	if strings.ToLower(params.MType) == "gauge" {
		value = strconv.FormatFloat(*params.Value, 'f', 6, 64)
	} else {
		value = strconv.FormatInt(*params.Delta, 10)
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

func (c *Client) SendMetrics(metrics models.Metrics) error {
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
