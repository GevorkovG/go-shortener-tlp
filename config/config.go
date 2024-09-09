package config

import (
	"flag"
	"os"
)

var AppConfig struct {
	Host      string
	ResultURL string
}

func init() {

	flag.StringVar(&AppConfig.Host, "a", "localhost:8080", "It's a Host")
	flag.StringVar(&AppConfig.ResultURL, "b", "http://localhost:8080", "It's a Result URL")

	if host := os.Getenv("SERVER_ADDRESS"); host != "" {
		AppConfig.Host = host
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		AppConfig.ResultURL = baseURL
	}
}
