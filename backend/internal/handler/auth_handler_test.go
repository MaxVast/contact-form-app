package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/maxvast/contact-form-app/backend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maxvast/contact-form-app/backend/internal/auth"
	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/service"
)

type fakeAuthService struct {
	user  *model.AdminUser
	err   error
	email string
}

func (f *fakeAuthService) Authenticate(ctx context.Context, email string, password string) (*model.AdminUser, error) {
	f.email = email

	if f.err != nil {
		return nil, f.err
	}

	return f.user, nil
}

func TestAuthHandler_Login_Success(t *testing.T) {
	authService := &fakeAuthService{
		user: &model.AdminUser{
			ID:    "994c3ed3-bd88-4b54-bb0e-2b560e34a3b1",
			Email: "max@exemple.com",
			Role:  model.AdminRole,
		},
	}

	jwtService, err := auth.NewJWTService("6HnfWwvmDy", time.Hour)
	if err != nil {
		t.Fatalf("failed to create jwt service: %v", err)
	}

	handler := NewAuthHandler(authService, jwtService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/login",
		strings.NewReader(`{
			"email": "max@exemple.com",
			"password": "password"
		}`),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got loginResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.Message != "authentication successful" {
		t.Errorf("expected message %q, got %q", "authentication successful", got.Message)
	}

	if got.TokenType != "Bearer" {
		t.Errorf("expected token_type %q, got %q", "Bearer", got.TokenType)
	}

	if got.ExpiresIn != int64(time.Hour.Seconds()) {
		t.Errorf("expected expires_in %d, got %d", int64(time.Hour.Seconds()), got.ExpiresIn)
	}

	if got.AccessToken == "" {
		t.Fatal("expected a non-empty access_token")
	}

	if strings.Count(got.AccessToken, ".") != 2 {
		t.Fatalf("expected access_token to be a JWT with 3 segments, got %q", got.AccessToken)
	}

	claims, err := jwtService.ValidateToken(got.AccessToken)
	if err != nil {
		t.Fatalf("expected access_token to be valid, got error: %v", err)
	}

	if claims.Subject != authService.user.ID {
		t.Errorf("expected subject %q, got %q", authService.user.ID, claims.Subject)
	}

	if claims.Email != authService.user.Email {
		t.Errorf("expected email %q, got %q", authService.user.Email, claims.Email)
	}

	if claims.Role != authService.user.Role {
		t.Errorf("expected role %q, got %q", authService.user.Role, claims.Role)
	}
}

func TestAuthHandler_Login_Errors(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		serviceErr     error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid json",
			body:           `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "invalid credentials",
			body: `{
				"email": "admin@example.com",
				"password": "wrong"
			}`,
			serviceErr:     service.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid credentials",
		},
		{
			name: "repository error",
			body: `{
				"email": "admin@example.com",
				"password": "password"
			}`,
			serviceErr:     errors.New("database unavailable"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &fakeAuthService{
				err: tt.serviceErr,
			}

			jwtService, err := auth.NewJWTService("6HnfWwvmDy", time.Hour)
			if err != nil {
				t.Fatalf("failed to create jwt service: %v", err)
			}

			handler := NewAuthHandler(authService, jwtService)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/admin/login",
				strings.NewReader(tt.body),
			)

			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}

			if !strings.Contains(rec.Body.String(), tt.expectedBody) {
				t.Fatalf(
					"expected body to contain %q, got %q",
					tt.expectedBody,
					rec.Body.String(),
				)
			}
		})
	}
}

func TestAuthHandler_Me(t *testing.T) {
	jwtService, err := auth.NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	handler := &AuthHandler{}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	user := &model.AdminUser{
		ID:    "994c3ed3-bd88-4b54-bb0e-2b560e34a3b1",
		Email: "admin@example.com",
		Role:  "admin",
	}

	token, err := jwtService.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	next := http.HandlerFunc(handler.Me)
	protectedHandler := jwtMiddleware.Authenticate(next)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != user.ID {
		t.Errorf("expected ID %q, got %q", user.ID, response.ID)
	}

	if response.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, response.Email)
	}

	if response.Role != user.Role {
		t.Errorf("expected role %q, got %q", user.Role, response.Role)
	}
}

func TestAuthHandler_Me_UnauthorizedWithoutToken(t *testing.T) {
	jwtService, err := auth.NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	handler := &AuthHandler{}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	protectedHandler := jwtMiddleware.Authenticate(
		http.HandlerFunc(handler.Me),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestAuthHandler_Me_UnauthorizedWithInvalidToken(t *testing.T) {
	jwtService, err := auth.NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	handler := &AuthHandler{}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	protectedHandler := jwtMiddleware.Authenticate(
		http.HandlerFunc(handler.Me),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestAuthHandler_Me_UnauthorizedWithWrongSecret(t *testing.T) {
	jwtService, err := auth.NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	wrongJWTService, err := auth.NewJWTService("wrong-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create wrong JWT service: %v", err)
	}

	user := &model.AdminUser{
		ID:    "994c3ed3-bd88-4b54-bb0e-2b560e34a3b1",
		Email: "admin@example.com",
		Role:  "admin",
	}

	token, err := wrongJWTService.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := &AuthHandler{}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	protectedHandler := jwtMiddleware.Authenticate(
		http.HandlerFunc(handler.Me),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}
