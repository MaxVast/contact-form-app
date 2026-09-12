package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/maxvast/contact-form-app/backend/internal/auth"
	"github.com/maxvast/contact-form-app/backend/internal/model"
)

const testJWTSecret = "test-secret"

func newTestJWTService(t *testing.T) *auth.JWTService {
	t.Helper()

	service, err := auth.NewJWTService(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	return service
}

func TestJWTMiddleware_Authenticate(t *testing.T) {
	jwtService := newTestJWTService(t)
	middleware := NewJWTMiddleware(jwtService)

	user := &model.AdminUser{
		ID:    "994c3ed3-bd88-4b54-bb0e-2b560e34a3b1",
		Email: "admin@example.com",
		Role:  "admin",
	}

	validToken, err := jwtService.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			authorization:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization format",
			authorization:  "Basic abc123",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer format",
			authorization:  "Bearer",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authorization:  "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token",
			authorization:  "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				claims, ok := ClaimsFromContext(r.Context())
				if !ok {
					t.Fatal("claims not found in context")
				}

				if claims.Email != user.Email {
					t.Errorf("expected email %q, got %q", user.Email, claims.Email)
				}

				if claims.Role != user.Role {
					t.Errorf("expected role %q, got %q", user.Role, claims.Role)
				}

				if claims.Subject != user.ID {
					t.Errorf("expected subject %q, got %q", user.ID, claims.Subject)
				}

				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.Authenticate(next)

			req := httptest.NewRequest(http.MethodGet, "/api/admin/messages", nil)

			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}

			if tt.expectedStatus == http.StatusOK && !nextCalled {
				t.Error("expected next handler to be called")
			}

			if tt.expectedStatus == http.StatusUnauthorized && nextCalled {
				t.Error("next handler should not be called")
			}
		})
	}
}
