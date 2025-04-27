package main

import (
	"encoding/gob"
	"log/slog"

	"github.com/joho/godotenv"

	"github.com/RichCake/calc_api_go/orchestrator/internal/application"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/calculation"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}
	// Регистрируем структуры которые будем сериализовать для хранения в базе
	gob.Register(calculation.Tree{})
	gob.Register(calculation.TreeNode{})
}

func main() {
	slog.Info("Starting application")
	app := application.New()
	defer app.Close()
	err := app.RunServer()
	if err != nil {
		panic(err)
	}
}
