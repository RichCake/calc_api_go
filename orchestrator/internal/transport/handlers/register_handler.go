package handlers

import (
	"encoding/json"
	"net/http"
)

type RegisterHandler struct {
	authService interface{}
}

func NewRegisterHandler(authService interface{}) *RegisterHandler {
	return &RegisterHandler{
		authService: authService,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {

	}
}
