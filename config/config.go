package config

import (
	"flag"

	"github.com/caarlos0/env"
)

/*
const (
	DBhost     = "localhost"
	DBuser     = "postgres"
	DBpassword = "6u8t3d804!"
	DBdbname   = "videos"
)
*/

type AppConfig struct {
	Host           string `env:"SERVER_ADDRESS"`
	ResultURL      string `env:"BASE_URL"`
	FilePATH       string `env:"FILE_STORAGE_PATH"`
	DataBaseString string `env:"DATABASE_DSN"`
}

func NewCfg() *AppConfig {

	a := AppConfig{}

	flag.StringVar(&a.Host, "a", "localhost:8080", "It's a Host")
	flag.StringVar(&a.ResultURL, "b", "http://localhost:8080", "It's a Result URL")
	flag.StringVar(&a.FilePATH, "f", "", "It's a FilePATH")
	flag.StringVar(&a.DataBaseString, "d", "", "it's conn string")

	flag.Parse()

	err := env.Parse(&a)
	if err != nil {
		panic(err)
	}

	return &a
}
