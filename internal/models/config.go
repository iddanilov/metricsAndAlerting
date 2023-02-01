package models

import "time"

type ServerConfig struct {
	Address     string `env:"ADDRESS"`
	LoggerLevel string `env:"LOGGER_LEVEL"`
}

type IO struct {
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
}

type Postgres struct {
	DSN string `env:"DATABASE_DSN"`
}

type Storage struct {
	Postgres
	IO
}

type Config struct {
	ServerConfig
	Storage
}
