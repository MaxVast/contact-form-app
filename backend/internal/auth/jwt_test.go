package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

func testAdminUser() *model.AdminUser {
	return &model.AdminUser{
		ID:    "994c3ed3-bd88-4b54-bb0e-2b560e34a3b1",
		Email: "max@exemple.com",
		Role:  model.AdminRole,
	}
}

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			secret:     "super-secret-key",
			expiration: time.Hour,
		},
		{
			name:       "empty secret",
			secret:     "",
			expiration: time.Hour,
			wantErr:    true,
		},
		{
			name:       "zero expiration",
			secret:     "super-secret-key",
			expiration: 0,
			wantErr:    true,
		},
		{
			name:       "negative expiration",
			secret:     "super-secret-key",
			expiration: -time.Hour,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewJWTService(tt.secret, tt.expiration)

			if tt.wantErr {
				if err == nil {
					t.Fatal("NewJWTService() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewJWTService() error = %v", err)
			}

			if service == nil {
				t.Fatal("NewJWTService() returned nil service")
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	user := testAdminUser()

	tokenString, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if tokenString == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	claims, err := service.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.Subject != user.ID {
		t.Errorf("Subject = %q, want %q", claims.Subject, user.ID)
	}

	if claims.Email != user.Email {
		t.Errorf("Email = %q, want %q", claims.Email, user.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("Role = %q, want %q", claims.Role, user.Role)
	}

	if claims.IssuedAt == nil {
		t.Fatal("IssuedAt is nil")
	}

	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}

	if !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		t.Error("ExpiresAt must be after IssuedAt")
	}
}

func TestGenerateTokenNilUser(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	_, err = service.GenerateToken(nil)

	if err == nil {
		t.Fatal("GenerateToken(nil) expected error")
	}
}

func TestGenerateTokenMissingUserID(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	user := testAdminUser()
	user.ID = ""

	_, err = service.GenerateToken(user)

	if err == nil {
		t.Fatal("GenerateToken() expected error for missing user ID")
	}
}

func TestGenerateTokenMissingEmail(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	user := testAdminUser()
	user.Email = ""

	_, err = service.GenerateToken(user)

	if err == nil {
		t.Fatal("GenerateToken() expected error for missing email")
	}
}

func TestGenerateTokenMissingRole(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	user := testAdminUser()
	user.Role = ""

	_, err = service.GenerateToken(user)

	if err == nil {
		t.Fatal("GenerateToken() expected error for missing role")
	}
}

func TestValidateToken(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	user := testAdminUser()

	tokenString, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := service.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.Subject != user.ID {
		t.Errorf("Subject = %q, want %q", claims.Subject, user.ID)
	}
}

func TestValidateTokenMissingToken(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	_, err = service.ValidateToken("")

	if err == nil {
		t.Fatal("ValidateToken() expected error")
	}

	if !strings.Contains(err.Error(), ErrMissingToken.Error()) {
		t.Errorf("error = %v, want %q", err, ErrMissingToken)
	}
}

func TestValidateTokenInvalidToken(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	_, err = service.ValidateToken("not-a-valid-token")

	if err == nil {
		t.Fatal("ValidateToken() expected error")
	}

	if !strings.Contains(err.Error(), ErrInvalidToken.Error()) {
		t.Errorf("error = %v, want %q", err, ErrInvalidToken)
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	otherService, err := NewJWTService("another-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	tokenString, err := service.GenerateToken(testAdminUser())
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = otherService.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() expected error with wrong secret")
	}
}

func TestValidateTokenExpired(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Millisecond)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	tokenString, err := service.GenerateToken(testAdminUser())
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() expected expiration error")
	}
}

func TestValidateTokenInvalidRole(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	user := testAdminUser()
	user.Role = "user"

	claims := Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("super-secret-key"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() expected invalid role error")
	}
}

func TestValidateTokenMissingSubject(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	claims := Claims{
		Email: "max@exemple.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("super-secret-key"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() expected missing subject error")
	}
}

func TestValidateTokenWrongSigningMethod(t *testing.T) {
	service, err := NewJWTService("super-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	claims := Claims{
		Email: "max@exemple.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   testAdminUser().ID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)

	tokenString, err := token.SignedString([]byte("super-secret-key"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() expected signing method error")
	}
}

func TestJWTService_Expiration(t *testing.T) {
	expiration := 2 * time.Hour

	service, err := NewJWTService("test-secret", expiration)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	if got := service.Expiration(); got != expiration {
		t.Errorf("Expiration() = %v, want %v", got, expiration)
	}
}

func TestJWTService_GenerateToken_InvalidUser(t *testing.T) {
	service, err := NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	tests := []struct {
		name string
		user *model.AdminUser
	}{
		{
			name: "nil user",
			user: nil,
		},
		{
			name: "missing user ID",
			user: &model.AdminUser{
				Email: "admin@example.com",
				Role:  model.AdminRole,
			},
		},
		{
			name: "missing user email",
			user: &model.AdminUser{
				ID:   "user-id",
				Role: model.AdminRole,
			},
		},
		{
			name: "missing user role",
			user: &model.AdminUser{
				ID:    "user-id",
				Email: "admin@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.user)

			if err == nil {
				t.Fatalf("GenerateToken() error = nil, want error")
			}

			if token != "" {
				t.Errorf("GenerateToken() token = %q, want empty token", token)
			}
		})
	}
}

func TestJWTService_ValidateToken_MissingSubject(t *testing.T) {
	service, err := NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	now := time.Now()

	claims := Claims{
		Email: "admin@example.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() error = nil, want error")
	}

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestJWTService_ValidateToken_InvalidRole(t *testing.T) {
	service, err := NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	token, err := service.GenerateToken(&model.AdminUser{
		ID:    "user-id",
		Email: "admin@example.com",
		Role:  "user",
	})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = service.ValidateToken(token)

	if err == nil {
		t.Fatal("ValidateToken() error = nil, want error")
	}

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestJWTService_ValidateToken_UnexpectedSigningMethod(t *testing.T) {
	service, err := NewJWTService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	claims := Claims{
		Email: "admin@example.com",
		Role:  model.AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-id",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = service.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() error = nil, want error")
	}

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ValidateToken() error = %v, want ErrInvalidToken", err)
	}
}
