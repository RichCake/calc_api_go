package main

import (
	"encoding/gob"
	"log/slog"

	"github.com/joho/godotenv"

	"github.com/RichCake/calc_api_go/internal/application"
	"github.com/RichCake/calc_api_go/internal/services/calculation"
)

func init() {
	// Инициализируем логер
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
	if err := app.RunServer(); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}