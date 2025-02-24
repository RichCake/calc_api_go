package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/RichCake/calc_api_go/internal/services/expression"
)

type TaskHandler struct {
	expressionService *expression.ExpressionService
}

func NewTaskHandler(expressionService *expression.ExpressionService) *TaskHandler {
	return &TaskHandler{
		expressionService: expressionService,
	}
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.giveTask(w)
	} else if r.Method == http.MethodPost {
		h.receiveTask(w, r)
	}
}

func (h *TaskHandler) giveTask(w http.ResponseWriter) {
	task := h.expressionService.GetPendingTask()
	if task == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no tasks available"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            task.ID,
		"arg1":          task.Arg1,
		"arg2":          task.Arg2,
		"operation":     task.Operation,
		"operation_time": task.OperationTime.Milliseconds(),
	})
}

func (h *TaskHandler) receiveTask(w http.ResponseWriter, r *http.Request) {
	var agent_request struct {
		TaskID int `json:"id"`
		Result float64 `json:"result"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&agent_request); err != nil {
		slog.Error("Agent sent an invalid body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
    	return
	}
	h.expressionService.ProcessIncomingTask(agent_request.TaskID, agent_request.Result)
}