package application

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/RichCake/calc_api_go/orchestrator/internal/config"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/auth"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
	"github.com/RichCake/calc_api_go/orchestrator/internal/storage"
	"github.com/RichCake/calc_api_go/orchestrator/internal/transport/handlers"
	"github.com/RichCake/calc_api_go/orchestrator/internal/transport/middlewares"
)

func setUpLogger(logFile *os.File) error {
	var logger = slog.New(slog.NewTextHandler(logFile, nil))
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
	storage := storage.NewStorage()

	// А вот и сервис по работе с выражениями. Он используется в хендлерах для обработки запросов
	expressionService := expression.NewExpressionService(storage, a.config.TimeConf)
	a.service = expressionService
	authService := auth.NewAuthService(storage, []byte(a.config.SecretKey))

	slog.Info("Starting server", "port", a.config.Addr)
	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)

	r.Handle("/auth/login", handlers.NewLoginHandler(authService)).Methods(http.MethodPost)
	r.Handle("/auth/register", handlers.NewRegisterHandler(authService)).Methods(http.MethodPost)
	r.Handle("/internal/task", handlers.NewTaskHandler(expressionService))

	authRequired := r.NewRoute().Subrouter()
	authRequired.Use(middlewares.NewAuthMiddleware([]byte(a.config.SecretKey)))

	authRequired.Handle("/api/v1/calculate", handlers.NewCalcHandler(expressionService)).Methods(http.MethodPost)
	authRequired.Handle("/api/v1/expressions", handlers.NewExpressionListHandler(expressionService)).Methods(http.MethodGet)
	authRequired.Handle("/api/v1/expressions/{id:[0-9]+}", handlers.NewExpressionHandler(expressionService)).Methods(http.MethodGet)

	http.Handle("/", r)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}

func (a *Application) Close() {
	slog.Info("Application shutdown")
	a.service.Close()
}