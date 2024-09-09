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

	if host := os.Getenv("SERVER_ADDRESS"); host != "" {
		AppConfig.Host = host
	} else {
		flag.StringVar(&AppConfig.Host, "a", "127.0.0.1:8080", "It's a Host")
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		AppConfig.ResultURL = baseURL
	} else {
		flag.StringVar(&AppConfig.ResultURL, "b", "127.0.0.1:8080/qsd54gFg", "It's a Result URL")
	}

}
