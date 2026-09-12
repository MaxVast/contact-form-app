package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/maxvast/contact-form-app/backend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maxvast/contact-form-app/backend/internal/auth"
	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/service"
)

type fakeAuthService struct {
	user     *model.AdminUser
	err      error
	email    string
	password string
}

func (f *fakeAuthService) Authenticate(
	_ context.Context,
	email string,
	password string,
) (*model.AdminUser, error) {
	f.email = email
	f.password = password

	return f.user, f.err
}

func newTestJWTService(t *testing.T) *auth.JWTService {
	t.Helper()

	jwtService, err := auth.NewJWTService("test-secret", time.Hour)
	require.NoError(t, err)

	return jwtService
}

func TestAuthHandler_Login_Success(t *testing.T) {
	authService := &fakeAuthService{
		user: &model.AdminUser{
			ID:    "123",
			Email: "john@example.com",
			Role:  model.AdminRole,
		},
	}

	handler := NewAuthHandler(
		authService,
		newTestJWTService(t),
	)

	body := `{
		"email": "john@example.com",
		"password": "password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(body),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response loginResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Equal(t, "authentication successful", response.Message)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.NotEmpty(t, response.AccessToken)
	assert.Equal(t, int64(time.Hour.Seconds()), response.ExpiresIn)

	assert.Equal(t, "john@example.com", authService.email)
	assert.Equal(t, "password", authService.password)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	authService := &fakeAuthService{}

	handler := NewAuthHandler(
		authService,
		newTestJWTService(t),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`invalid json`),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "invalid request body\n", rec.Body.String())
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	authService := &fakeAuthService{
		err: service.ErrInvalidCredentials,
	}

	handler := NewAuthHandler(
		authService,
		newTestJWTService(t),
	)

	body := `{
		"email": "john@example.com",
		"password": "wrong-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(body),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "invalid credentials\n", rec.Body.String())
}

func TestAuthHandler_Login_AuthServiceError(t *testing.T) {
	authService := &fakeAuthService{
		err: errors.New("database error"),
	}

	handler := NewAuthHandler(
		authService,
		newTestJWTService(t),
	)

	body := `{
		"email": "john@example.com",
		"password": "password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(body),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal server error\n", rec.Body.String())
}

func TestAuthHandler_Login_JWTError(t *testing.T) {
	authService := &fakeAuthService{
		user: &model.AdminUser{
			ID:    "", // GenerateToken() doit échouer ici
			Email: "john@example.com",
			Role:  model.AdminRole,
		},
	}

	handler := NewAuthHandler(
		authService,
		newTestJWTService(t),
	)

	body := `{
		"email": "john@example.com",
		"password": "password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(body),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal server error\n", rec.Body.String())
}

func TestAuthHandler_Me_Success(t *testing.T) {
	handler := NewAuthHandler(
		&fakeAuthService{},
		newTestJWTService(t),
	)

	claims := &auth.Claims{
		Email: "john@example.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-123",
		},
	}

	ctx := middleware.ContextWithClaims(
		context.Background(),
		claims,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/me",
		nil,
	).WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Equal(t, "user-123", response.ID)
	assert.Equal(t, "john@example.com", response.Email)
	assert.Equal(t, model.AdminRole, response.Role)
}

func TestAuthHandler_Me_ReturnsClaimsValues(t *testing.T) {
	handler := NewAuthHandler(
		&fakeAuthService{},
		newTestJWTService(t),
	)

	claims := &auth.Claims{
		Email: "admin@test.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "admin-456",
		},
	}

	ctx := middleware.ContextWithClaims(
		context.Background(),
		claims,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/me",
		nil,
	).WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Equal(t, "admin-456", response["id"])
	assert.Equal(t, "admin@test.com", response["email"])
	assert.Equal(t, model.AdminRole, response["role"])
}

func TestAuthHandler_Me_EmptySubject(t *testing.T) {
	handler := NewAuthHandler(
		&fakeAuthService{},
		newTestJWTService(t),
	)

	claims := &auth.Claims{
		Email: "john@example.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "",
		},
	}

	ctx := middleware.ContextWithClaims(
		context.Background(),
		claims,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/me",
		nil,
	).WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Empty(t, response.ID)
	assert.Equal(t, "john@example.com", response.Email)
	assert.Equal(t, model.AdminRole, response.Role)
}

func TestAuthHandler_Me_EmptyEmail(t *testing.T) {
	handler := NewAuthHandler(
		&fakeAuthService{},
		newTestJWTService(t),
	)

	claims := &auth.Claims{
		Email: "",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-123",
		},
	}

	ctx := middleware.ContextWithClaims(
		context.Background(),
		claims,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/me",
		nil,
	).WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Equal(t, "user-123", response.ID)
	assert.Empty(t, response.Email)
	assert.Equal(t, model.AdminRole, response.Role)
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	handler := NewAuthHandler(
		&fakeAuthService{},
		newTestJWTService(t),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "unauthorized\n", rec.Body.String())
}

func TestAuthHandler_Me_WithValidJWT(t *testing.T) {
	jwtService := newTestJWTService(t)

	user := &model.AdminUser{
		ID:    "user-123",
		Email: "john@example.com",
		Role:  model.AdminRole,
	}

	token, err := jwtService.GenerateToken(user)
	require.NoError(t, err)

	claims, err := jwtService.ValidateToken(token)
	require.NoError(t, err)

	ctx := middleware.ContextWithClaims(
		context.Background(),
		claims,
	)

	handler := NewAuthHandler(
		&fakeAuthService{},
		jwtService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/me",
		nil,
	).WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.Role, response.Role)
}
