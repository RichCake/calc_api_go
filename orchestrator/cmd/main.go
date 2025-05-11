package main

import (
	"log/slog"

	"github.com/joho/godotenv"

	"github.com/RichCake/calc_api_go/orchestrator/internal/application"
)

// Выполняется перед main
func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}
}

func main() {
	slog.Info("Starting application")
	app := application.New(false)
	defer app.Close()
	err := app.RunServer()
	if err != nil {
		panic(err)
	}
}
