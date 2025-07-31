package main

import (
	"fmt"
	"log"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
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

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
