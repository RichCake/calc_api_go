package application

import (
	"log/slog"
	"os"

	grpcserver "github.com/RichCake/calc_api_go/orchestrator/internal/application/grpc"
	httpserver "github.com/RichCake/calc_api_go/orchestrator/internal/application/http"
	"github.com/RichCake/calc_api_go/orchestrator/internal/config"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/auth"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
	"github.com/RichCake/calc_api_go/orchestrator/internal/storage"
)

func setUpLogger(logFile *os.File) error {
	opts := slog.HandlerOptions{}
	var logger = slog.New(slog.NewTextHandler(logFile, &opts))
	slog.SetDefault(logger)
	return nil
}

type Application struct {
	config  *config.Config
	service *expression.ExpressionService
}

func New() *Application {
	config, err := config.ConfigFromEnv()
	if err != nil {
		panic(err)
	}

	return &Application{
		config: config,
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

	// Создаем хранилище, которое будет передаваться вглубь приложения по ссылке,
	// то есть все сервисы будут работать с одним и тем же хранилищем
	storage := storage.NewStorage(false)

	// А вот и сервис по работе с выражениями. Он используется в хендлерах для обработки запросов
	expressionService := expression.NewExpressionService(storage, a.config.TimeConf)
	a.service = expressionService
	authService := auth.NewAuthService(storage, []byte(a.config.SecretKey))

	go httpserver.RunHTTPServer(authService, expressionService, *a.config)
	grpcserver.RunGRPCServer(expressionService, *a.config)
	return nil
}

func (a *Application) Close() {
	slog.Info("Application shutdown")
	a.service.Close()
}
