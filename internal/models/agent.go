package models

import (
	"net/http"
	"time"
)

type AgentConfig struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Key            string        `env:"KEY"`
	LoggerLevel    string        `env:"LOGGER_LEVEL"`
}

type Agent struct {
	HTTPClient  *http.Client
	AgentConfig AgentConfig
}
