package main

import (
	"log/slog"
	"github.com/RichCake/calc_api_go/internal/application"
)

func main() {
	slog.Info("Starting application")
	app := application.New()
	if err := app.RunServer(); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}