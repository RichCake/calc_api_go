package application

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/RichCake/calc_api_go/internal/config"
	"github.com/RichCake/calc_api_go/internal/storage"
	"github.com/RichCake/calc_api_go/internal/services/expression"
	"github.com/RichCake/calc_api_go/internal/transport/handlers"
	"github.com/RichCake/calc_api_go/internal/transport/middlewares"
)

func setUpLogger(logFile *os.File) error {
	var logger = slog.New(slog.NewTextHandler(logFile, nil))
	slog.SetDefault(logger)
	return nil
}

type Application struct {
	config *config.Config
}

func New() *Application {
	return &Application{
		config: config.ConfigFromEnv(),
	}
}

func (a *Application) RunServer() error {
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Error while opening log file", "error", err)
	}
	defer logFile.Close()
	err = setUpLogger(logFile)
	if err != nil {
		slog.Error("Error while setting up logger", "error", err)
	}

	storage := storage.NewStorage()

	expressionService := expression.NewExpressionService(storage)

	slog.Info("Starting server", "port", a.config.Addr)
	http.Handle("/api/v1/calculate", middlewares.LoggingMiddleware(handlers.NewCalcHandler(expressionService)))
	http.Handle("/api/v1/expressions",middlewares.LoggingMiddleware(handlers.NewExpressionHandler(expressionService)))
	http.Handle("/internal/task", middlewares.LoggingMiddleware(handlers.NewTaskHandler(expressionService)))
	return http.ListenAndServe(":"+a.config.Addr, nil)
}