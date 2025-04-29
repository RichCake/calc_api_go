package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RichCake/calc_api_go/orchestrator/internal/services/auth"
)

type LoginHandler struct {
	authService *auth.AuthService
}

func NewLoginHandler(authService *auth.AuthService) *LoginHandler {
	return &LoginHandler{
		authService: authService,
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	token, err := h.authService.Login(request.Login, request.Password)
	if errors.Is(err, auth.ErrBadCredentials) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"access_token": token})
}
