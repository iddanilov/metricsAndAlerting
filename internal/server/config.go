package server

import (
	"encoding/json"
	goflag "flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

var (
	Address       = flag.StringP("a", "a", "127.0.0.1:8080", "help message for Address")
	StoreFile     = flag.StringP("f", "f", "/tmp/devops-metrics-db.json", "help message for StoreFile")
	StoreInterval = flag.DurationP("i", "i", 300*time.Second, "help message for StoreInterval")
	Restore       = flag.BoolP("r", "r", true, "help message for Restore")
	Key           = flag.StringP("k", "k", "", "help message for KEY")
	DSN           = flag.StringP("d", "d", "", "help message for DSN")
	CryptoKey     = flag.StringP("certs-key", "", "", "help message for DSN")
	JsonConfig    = flag.StringP("config", "c", "", "help message for DSN")
)

type Config struct {
	Address       string        `env:"ADDRESS" json:"address"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" json:"store_interval"`
	StoreFile     string        `env:"STORE_FILE" json:"store_file"`
	Restore       bool          `env:"RESTORE" json:"restore"`
	Key           string        `env:"KEY"`
	DSN           string        `env:"DATABASE_DSN" json:"database_dsn"`
	CryptoKey     string        `env:"CRYPTO_KEY" json:"crypto_key"`
	JsonConfig    string        `env:"CONFIG"`
}

func NewConfig() *Config {
	var jsonConfig Config

	var cfg Config

	err := env.Parse(&cfg)

	if *JsonConfig != "" || cfg.JsonConfig != "" {
		log.Println("use JsonConfig")
		if jsonConfig.JsonConfig != "" {
			err := readFromJson(jsonConfig.JsonConfig, &jsonConfig)
			if err != nil {
				return nil
			}
		} else {
			err := readFromJson(*JsonConfig, &jsonConfig)
			if err != nil {
				return nil
			}
		}
	}

	if err != nil {
		panic(err)
	}
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if cfg.Address == "" {
		log.Println("cfg.Address: ", cfg.Address)
		log.Println("jsonConfig.Address: ", jsonConfig.Address)
		if Address != nil {
			cfg.Address = *Address
		} else {
			cfg.Address = jsonConfig.Address
		}
	}
	if cfg.StoreInterval == 0 {
		if StoreInterval != nil {
			cfg.StoreInterval = *StoreInterval
		} else {
			cfg.StoreInterval = jsonConfig.StoreInterval
		}

	}
	if cfg.StoreFile == "" {
		cfg.StoreFile = *StoreFile
	}
	if cfg.CryptoKey == "" {
		cfg.CryptoKey = *CryptoKey
	}
	if *Key != "" {
		cfg.Key = *Key
	}
	if cfg.DSN == "" {
		cfg.DSN = *DSN
	}
	if os.Getenv("RESTORE") == "" {
		cfg.Restore = *Restore
	}

	return &cfg

}

func readFromJson(path string, cfg *Config) error {

	var temp []byte

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Read(temp) // filename is the JSON file to read
	if err != nil {
		return err
	}
	err = json.Unmarshal(temp, cfg)
	if err != nil {
		log.Println("Cannot unmarshal the json ", err)
		return err
	}

	return nil
}
