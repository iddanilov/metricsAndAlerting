package app

import (
	goflag "flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"

	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

var (
	Address       = flag.StringP("a", "a", "127.0.0.1:8080", "help message for Address")
	StoreFile     = flag.StringP("f", "f", "/tmp/devops-metrics-db.json", "help message for StoreFile")
	StoreInterval = flag.DurationP("i", "i", 300*time.Second, "help message for StoreInterval")
	Restore       = flag.BoolP("r", "r", true, "help message for Restore")
	Key           = flag.StringP("k", "k", "", "help message for KEY")
	DSN           = flag.StringP("d", "d", "", "help message for DSN")
)

func NewConfig() *models.Config {
	var cfg models.Config

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if cfg.Address == "" {
		cfg.Address = *Address
	}
	if cfg.Storage.StoreInterval == 0 {
		cfg.StoreInterval = *StoreInterval
	}
	if cfg.Storage.StoreFile == "" {
		cfg.StoreFile = *StoreFile
	}
	if *Key != "" {
		cfg.Storage.Key = *Key
	}
	if cfg.DSN == "" {
		cfg.Postgres.DSN = *DSN
	}
	if os.Getenv("RESTORE") == "" {
		cfg.Restore = *Restore
	}

	log.Println(cfg.Address)
	log.Println(cfg)
	return &cfg

}
