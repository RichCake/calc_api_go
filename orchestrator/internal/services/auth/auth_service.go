package auth

import (
	"errors"
	"log/slog"
	"time"

	"github.com/RichCake/calc_api_go/orchestrator/internal/models"
	"github.com/RichCake/calc_api_go/orchestrator/internal/storage"
	"github.com/golang-jwt/jwt/v5"
)
type ContextKey string
const ContextKeyUserID ContextKey = "user_id"

type AuthService struct {
	storage *storage.Storage
	secret_key []byte
}

func NewAuthService(storage *storage.Storage, secret_key []byte) *AuthService {
	return &AuthService{
		storage: storage,
		secret_key: secret_key,
	}
}

func (s *AuthService) Register(login string, password string) (int, error) {
	var newUser models.User
	newUser.Login = login
	newUser.Password = password
	passwordHash, err := generate(password)
	if err != nil {
		slog.Error("Encryption failed", "error", err.Error())
		return 0, ErrEncryption
	}
	newUser.PasswordHash = passwordHash
	_, err = s.storage.SaveUser(&newUser)
	if errors.Is(err, storage.ErrUsernameTaken) {
		return 0, ErrUserExists
	} else if err != nil {
		slog.Error("Registration failed", "error", err.Error())
		return newUser.ID, ErrStorage
	}
	return newUser.ID, nil
}

func (s *AuthService) Login(login string, password string) (string, error) {
	user, err := s.storage.GetUser(login)
	if errors.Is(err, storage.ErrItemNotFound) {
		return "", ErrBadCredentials
	} else if err != nil {
		slog.Error("Login failed", "error", err.Error())
		return "", ErrStorage
	}
	if compare(user.PasswordHash, password) != nil {
		return "", ErrBadCredentials
	}
	payload := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour*24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	t, err := token.SignedString(s.secret_key)
	if err != nil {
		slog.Error("Signing failed", "error", err.Error())
		return "", ErrEncryption
	}
	return t, nil
}
