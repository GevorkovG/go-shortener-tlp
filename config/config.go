package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	DBhost     = "localhost"
	DBuser     = "postgres"
	DBpassword = "!"
	DBdbname   = "short"
)

type AppConfig struct {
	Host           string
	ResultURL      string
	FilePATH       string
	DataBaseString string
}

func NewCfg() *AppConfig {

	a := &AppConfig{}

	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		DBhost, DBuser, DBpassword, DBdbname)

	flag.StringVar(&a.Host, "a", "localhost:8080", "It's a Host")
	flag.StringVar(&a.ResultURL, "b", "http://localhost:8080", "It's a Result URL")
	flag.StringVar(&a.FilePATH, "f", "/tmp/shortener-db.json", "It's a FilePATH")
	flag.StringVar(&a.DataBaseString, "d", ps, "it's conn string")

	if host := os.Getenv("SERVER_ADDRESS"); host != "" {
		a.Host = strings.TrimSpace(host)
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		a.ResultURL = strings.TrimSpace(baseURL)
	}
	if baseURL := os.Getenv("FILE_STORAGE_PATH"); baseURL != "" {
		a.FilePATH = strings.TrimSpace(baseURL)
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		a.ResultURL = strings.TrimSpace(baseURL)
	}
	if host := os.Getenv("SERVER_ADDRESS"); host != "" {
		a.Host = strings.TrimSpace(host)
	}
	if file := os.Getenv("FILE_STORAGE_PATH"); file != "" {
		a.FilePATH = file
	}
	if datab := os.Getenv("DATABASE_DSN"); datab != "" {
		a.DataBaseString = datab
	}

	return a
}
