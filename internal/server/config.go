package server

import "github.com/caarlos0/env/v6"

type Config struct {
	ADDRESS string `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
}

func NewConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil
	}
	return &cfg

}
