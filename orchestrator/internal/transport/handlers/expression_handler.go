package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/RichCake/calc_api_go/orchestrator/internal/services/auth"
	"github.com/RichCake/calc_api_go/orchestrator/internal/services/expression"
	"github.com/gorilla/mux"
)

type ExpressionHandler struct {
	expressionService *expression.ExpressionService
}

func NewExpressionHandler(expressionService *expression.ExpressionService) *ExpressionHandler {
	return &ExpressionHandler{
		expressionService: expressionService,
	}
}

func (h *ExpressionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	expression_id_str := vars["id"]

	if expression_id_str == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is required"})
		return
	}

	expression_id, err := strconv.Atoi(expression_id_str)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id must be a number"})
		return
	}
	
	user_id := r.Context().Value(auth.ContextKeyUserID).(int)
	// Логика спрятана сюда
	e, err := h.expressionService.GetExpressionByID(expression_id, user_id)
	if errors.Is(err, expression.ErrExpressionNotFound) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "expression not found"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}
