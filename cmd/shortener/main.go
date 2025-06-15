package main

import (
	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/log"
	"go.uber.org/zap"
)

func main() {
	log.InitLogger()     // Инициализация логгера
	defer zap.L().Sync() // Очистка буферов логгера при завершении программы

	app.Run()
}
