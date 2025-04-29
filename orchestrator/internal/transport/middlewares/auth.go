package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/RichCake/calc_api_go/orchestrator/internal/services/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

func NewAuthMiddleware(secret_key []byte) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}
	
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
	
			tokenStr := parts[1]
	
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return secret_key, nil
			})
			
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
	
			user_id, ok := claims["sub"].(float64)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
	
			ctx := context.WithValue(r.Context(), auth.ContextKeyUserID, int(user_id))
	
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
