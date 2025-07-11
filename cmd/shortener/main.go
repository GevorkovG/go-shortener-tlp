package main

import (
	"fmt"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/log"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	log.InitLogger()     // Инициализация логгера
	defer zap.L().Sync() // Очистка буферов логгера при завершении программы

	zap.L().Info("Application started",
		zap.String("version", buildVersion),
		zap.String("build_date", buildDate),
		zap.String("commit", buildCommit),
	)

	app.Run()
}
