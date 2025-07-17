// Package config содержит конфигурацию приложения
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

// AppConfig содержит конфигурационные параметры приложения.
// Поля структуры:
//   - Host: адрес сервера (env:"SERVER_ADDRESS")
//   - ResultURL: базовый URL для сокращенных ссылок (env:"BASE_URL")
//   - FilePATH: путь к файлу хранилища (env:"FILE_STORAGE_PATH")
//   - DataBaseString: строка подключения к БД (env:"DATABASE_DSN")
//   - EnableHTTPS:    включить HTTPS (env:"ENABLE_HTTPS")
type AppConfig struct {
	Host           string `env:"SERVER_ADDRESS"`
	ResultURL      string `env:"BASE_URL"`
	FilePATH       string `env:"FILE_STORAGE_PATH"`
	DataBaseString string `env:"DATABASE_DSN"`
	EnableHTTPS    string `env:"ENABLE_HTTPS"`
}

// NewCfg создает и инициализирует конфигурацию приложения, используя:
// 1. Флаги командной строки (со значениями по умолчанию)
// 2. Переменные окружения (переопределяют значения флагов)
//
// Поддерживаемые параметры конфигурации:
//   - Host (флаг -a) - адрес сервера (по умолчанию "localhost:8080")
//   - ResultURL (флаг -b) - базовый URL для результатов (по умолчанию "http://localhost:8080")
//   - FilePATH (флаг -f) - путь к файлу хранилища (по умолчанию "")
//   - DataBaseString (флаг -d) - строка подключения к БД (по умолчанию "")
//   - EnableHTTPS (флаг -s) - включить HTTPS (по умолчанию "")
//
// Возвращает:
//   - *AppConfig: указатель на инициализированную конфигурацию
//   - panic: в случае ошибки парсинга переменных окружения
//
// Пример использования:
//
//	cfg := NewCfg()
//	// Или с параметрами командной строки:
//	// ./app -a :9090 -b http://example.com -f ./data.json
//
// Примечания:
// 1. Переменные окружения имеют приоритет над флагами командной строки
// 2. Для парсинга переменных окружения используется пакет env
// 3. Имена переменных окружения должны соответствовать полям структуры AppConfig
func NewCfg() *AppConfig {

	a := AppConfig{}

	flag.StringVar(&a.Host, "a", "localhost:8080", "It's a Host")
	flag.StringVar(&a.ResultURL, "b", "http://localhost:8080", "It's a Result URL")
	flag.StringVar(&a.FilePATH, "f", "", "It's a FilePATH")
	flag.StringVar(&a.DataBaseString, "d", "", "it's conn string")
	flag.StringVar(&a.EnableHTTPS, "s", "", "using HTTPS")

	flag.Parse()

	err := env.Parse(&a)
	if err != nil {
		panic(err)
	}

	return &a
}
