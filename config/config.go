// Package config содержит конфигурацию приложения
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

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

// AppConfig содержит конфигурационные параметры приложения.
// Поля структуры:
//   - Host: адрес сервера (env:"SERVER_ADDRESS")
//   - ResultURL: базовый URL для сокращенных ссылок (env:"BASE_URL")
//   - FilePATH: путь к файлу хранилища (env:"FILE_STORAGE_PATH")
//   - DataBaseString: строка подключения к БД (env:"DATABASE_DSN")
//   - EnableHTTPS:    включить HTTPS (env:"ENABLE_HTTPS")
type AppConfig struct {
	Host           string `env:"SERVER_ADDRESS" json:"server_address"`
	ResultURL      string `env:"BASE_URL" json:"base_url"`
	FilePATH       string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DataBaseString string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS    bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigJSON     string `env:"CONFIG" json:"-"`
}

// loadConfigFromFile загружает конфигурацию приложения из файла.
func (a *AppConfig) loadConfigFromFile() error {
	if a.ConfigJSON == "" {
		return nil
	}

	data, err := os.ReadFile(a.ConfigJSON)
	if err != nil {
		return err
	}

	var fileConfig AppConfig
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return err
	}

	if a.Host == defaultServerAddress && fileConfig.Host != "" {
		a.Host = fileConfig.Host
	}
	if a.ResultURL == defaultBaseURL && fileConfig.ResultURL != "" {
		a.ResultURL = fileConfig.ResultURL
	}
	if a.FilePATH == "" && fileConfig.FilePATH != "" {
		a.FilePATH = fileConfig.FilePATH
	}
	if a.DataBaseString == "" && fileConfig.DataBaseString != "" {
		a.DataBaseString = fileConfig.DataBaseString
	}
	if !a.EnableHTTPS && fileConfig.EnableHTTPS {
		a.EnableHTTPS = fileConfig.EnableHTTPS
	}

	return nil
}

const (
	defaultServerAddress = "localhost:8080"

	defaultBaseURL = "http://localhost:8080"
)

// NewCfg создает и инициализирует конфигурацию приложения.
// Приоритеты источников конфигурации (от высшего к низшему):
// 1. Флаги командной строки
// 2. Переменные окружения
// 3. JSON-файл конфигурации
// 4. Значения по умолчанию
//
// Поддерживаемые параметры конфигурации:
//   - Host (флаг -a) - адрес сервера (по умолчанию "localhost:8080")
//   - ResultURL (флаг -b) - базовый URL для результатов (по умолчанию "http://localhost:8080")
//   - FilePATH (флаг -f) - путь к файлу хранилища (по умолчанию "")
//   - DataBaseString (флаг -d) - строка подключения к БД (по умолчанию "")
//   - EnableHTTPS (флаг -s) - включить HTTPS (по умолчанию "")
func NewCfg() *AppConfig {

	a := AppConfig{}

	a.Host = defaultServerAddress
	a.ResultURL = defaultBaseURL

	flag.StringVar(&a.Host, "a", defaultServerAddress, "It's a Host")
	flag.StringVar(&a.ResultURL, "b", defaultBaseURL, "It's a Result URL")
	flag.StringVar(&a.FilePATH, "f", "", "It's a FilePATH")
	flag.StringVar(&a.DataBaseString, "d", "", "it's conn string")
	flag.BoolVar(&a.EnableHTTPS, "s", false, "using HTTPS")
	flag.StringVar(&a.ConfigJSON, "c", "", "It's a ConfigJSON file")

	flag.Parse()

	err := env.Parse(&a)
	if err != nil {
		panic(err)
	}

	if a.ConfigJSON != "" {
		if err := a.loadConfigFromFile(); err != nil {
			log.Printf("Warning: failed to load config file: %v", err)
		}
	}

	return &a
}
