package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/RichCake/calc_api_go/internal/services/expression"
)

// ExpressionHandler - обработчик для работы с выражениями
type ExpressionHandler struct {
	expressionService *expression.ExpressionService
}

// Конструктор хендлера
func NewExpressionHandler(expressionService *expression.ExpressionService) *ExpressionHandler {
	return &ExpressionHandler{
		expressionService: expressionService,
	}
}

// Метод для получения списка выражений
func (h *ExpressionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	// Получаем список выражений из сервиса
	expressions := h.expressionService.GetExpressions()

	// Возвращаем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expressions)
}