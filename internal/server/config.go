package server

import (
	goflag "flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

var (
	Address       = flag.StringP("a", "a", "127.0.0.1:8080", "help message for flagname")
	StoreFile     = flag.StringP("f", "f", "/tmp/devops-metrics-db.json", "help message for flagname")
	StoreInterval = flag.DurationP("i", "i", 1*time.Second, "help message for flagname")
	Restore       = flag.BoolP("r", "r", true, "help message for flagname")
)

type Config struct {
	Address       string        `env:"Address"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"Restore"`
}

func NewConfig() *Config {
	var cfg Config

	err := env.Parse(&cfg)
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if cfg.Address == "" {
		cfg.Address = *Address
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = *StoreInterval
	}
	if cfg.StoreFile == "" {
		cfg.StoreFile = *StoreFile
	}
	if os.Getenv("Restore") == "" {
		cfg.Restore = *Restore
	}

	log.Println(cfg.Address)
	log.Println(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &cfg

}
