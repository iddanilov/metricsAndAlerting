// Package app - package client application
// Prepares config and client methods
package app

import (
	goflag "flag"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

var (
	Address        = flag.StringP("a", "a", "127.0.0.1:8080", "help message for flagname")
	PollInterval   = flag.DurationP("p", "p", 2*time.Second, "help message for flagname")
	ReportInterval = flag.DurationP("r", "r", 10*time.Second, "help message for flagname")
	Key            = flag.StringP("k", "k", "", "help message for KEY")
	LoggerLevel    = flag.StringP("l", "l", "debug", "LoggerLevel")
)

func NewConfig() *models.Agent {
	var cfg models.AgentConfig
	err := env.Parse(&cfg)
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	if cfg.Address == "" {
		cfg.Address = *Address
	}
	if cfg.LoggerLevel == "" {
		cfg.LoggerLevel = *LoggerLevel
	}
	if !strings.Contains(cfg.Address, "http") {
		cfg.Address = "http://" + cfg.Address
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *ReportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *PollInterval
	}
	if *Key != "" {
		cfg.Key = *Key
	}

	if err != nil {
		return nil
	}
	return &models.Agent{
		AgentConfig: cfg,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}
