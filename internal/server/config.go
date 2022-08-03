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
	ADDRESS       *string        = flag.StringP("a", "a", "127.0.0.1:8080", "help message for flagname")
	StoreFile     *string        = flag.StringP("f", "f", "/tmp/devops-metrics-db.json", "help message for flagname")
	StoreInterval *time.Duration = flag.DurationP("i", "i", time.Duration(300*time.Second), "help message for flagname")
	RESTORE       *bool          = flag.BoolP("r", "r", true, "help message for flagname")
)

type Config struct {
	ADDRESS       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	RESTORE       bool          `env:"RESTORE"`
}

func NewConfig() *Config {
	var cfg Config

	err := env.Parse(&cfg)
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if cfg.ADDRESS == "" {
		cfg.ADDRESS = *ADDRESS
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = *StoreInterval
	}
	if cfg.StoreFile == "" {
		cfg.StoreFile = *StoreFile
	}
	if os.Getenv("RESTORE") == "" {
		cfg.RESTORE = *RESTORE
	}

	log.Println(cfg.ADDRESS)
	log.Println(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &cfg

}
