package http

import (
	"log/slog"
	"net/http"

	"github.com/RichCake/calc_api_go/orchestrator/internal/config"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/auth"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
	"github.com/RichCake/calc_api_go/orchestrator/internal/transport/handlers"
	"github.com/RichCake/calc_api_go/orchestrator/internal/transport/middlewares"
	"github.com/gorilla/mux"
)

func RunHTTPServer(authService *auth.AuthService, expressionService *expression.ExpressionService, config config.Config) {
	slog.Info("Starting server", "port", config.Addr)
	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)

	r.Handle("/api/v1/login", handlers.NewLoginHandler(authService)).Methods(http.MethodPost)
	r.Handle("/api/v1/register", handlers.NewRegisterHandler(authService)).Methods(http.MethodPost)

	authRequired := r.NewRoute().Subrouter()
	authRequired.Use(middlewares.NewAuthMiddleware([]byte(config.SecretKey)))

	authRequired.Handle("/api/v1/calculate", handlers.NewCalcHandler(expressionService)).Methods(http.MethodPost)
	authRequired.Handle("/api/v1/expressions", handlers.NewExpressionListHandler(expressionService)).Methods(http.MethodGet)
	authRequired.Handle("/api/v1/expressions/{id:[0-9]+}", handlers.NewExpressionHandler(expressionService)).Methods(http.MethodGet)

	http.Handle("/", r)
	if err := http.ListenAndServe(":"+config.Addr, nil); err != nil {
		panic(err)
	}
}
