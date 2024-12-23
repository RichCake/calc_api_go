package application

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/RichCake/calc_api_go/pkg/calculation"
)

func setUpLogger(logFile *os.File) error {
	var logger = slog.New(slog.NewTextHandler(logFile, nil))
	slog.SetDefault(logger)
	return nil
}

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

type Request struct {
	Expression string `json:"expression"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		slog.Warn("Method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err == io.EOF {
		slog.Warn("Missing request body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing request body"})
		return
	}
	if err != nil {
		slog.Error("Error decoding request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if request.Expression == "" {
		slog.Warn("Missing 'expression' field")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "'expression' field is required"})
		return
	}
	slog.Info("Get expression", "expression", request.Expression)

	result, err := calculation.Calc(request.Expression)
	if errors.Is(calculation.ErrCalculation, err) {
		slog.Warn("Unprocessable entity error", "error", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	} else if err != nil {
		slog.Error("Unknown calculation error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	} else {
		slog.Info("Calculation result", "result", result)
		json.NewEncoder(w).Encode(map[string]float64{"result": result})
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
	slog.Info("Starting server", "port", a.config.Addr)
	http.Handle("/api/v1/calculate", loggingMiddleware(http.HandlerFunc(CalcHandler)))
	return http.ListenAndServe(":"+a.config.Addr, nil)
}