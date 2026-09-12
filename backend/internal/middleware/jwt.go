package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/maxvast/contact-form-app/backend/internal/auth"
)

type contextKey string

const jwtClaimsKey contextKey = "jwt_claims"

type JWTMiddleware struct {
	jwtService *auth.JWTService
}

func NewJWTMiddleware(jwtService *auth.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))

		if header == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(header)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), jwtClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(jwtClaimsKey).(*auth.Claims)

	return claims, ok
}
