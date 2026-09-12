package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maxvast/contact-form-app/backend/internal/auth"
	"github.com/maxvast/contact-form-app/backend/internal/middleware"
	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/service"
)

type AuthService interface {
	Authenticate(ctx context.Context, email string, password string) (*model.AdminUser, error)
}

type AuthHandler struct {
	authService AuthService
	jwtService  *auth.JWTService
}

func NewAuthHandler(authService AuthService, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Authenticate(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := loginResponse{
		Message:     "authentication successful",
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.jwtService.Expiration().Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}{
		ID:    claims.Subject,
		Email: claims.Email,
		Role:  claims.Role,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
