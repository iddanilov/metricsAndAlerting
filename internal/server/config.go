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
	Address       = flag.StringP("a", "a", "127.0.0.1:8080", "help message for Address")
	StoreFile     = flag.StringP("f", "f", "tmp/devops-metrics-db.json", "help message for StoreFile")
	StoreInterval = flag.DurationP("i", "i", 300*time.Second, "help message for StoreInterval")
	Restore       = flag.BoolP("r", "r", true, "help message for Restore")
	Key           = flag.StringP("k", "k", "", "help message for KEY")
	DSN           = flag.StringP("d", "d", "", "help message for DSN")
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DSN           string        `env:"DATABASE_DSN"`
}

func NewConfig() *Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
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
	if *Key != "" {
		cfg.Key = *Key
	}

	cfg.DSN = "***postgres:5432/praktikum?sslmode=disable"

	//if cfg.DSN == "" {
	//	cfg.DSN =
	//}
	if os.Getenv("RESTORE") == "" {
		cfg.Restore = *Restore
	}

	log.Println(cfg.Address)
	log.Println(cfg)
	return &cfg

}
