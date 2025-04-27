package handlers

import (
	"encoding/json"
	"net/http"
)

type LoginHandler struct {
	authService interface{}
}

func NewLoginHandler(authService interface{}) *LoginHandler {
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

	}
}
