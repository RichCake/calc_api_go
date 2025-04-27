package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
)

type ExpressionListHandler struct {
	expressionService *expression.ExpressionService
}

func NewExpressionListHandler(expressionService *expression.ExpressionService) *ExpressionListHandler {
	return &ExpressionListHandler{
		expressionService: expressionService,
	}
}

func (h *ExpressionListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// Логика спрятана сюда
	expressions, err := h.expressionService.GetExpressions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expressions)
}
