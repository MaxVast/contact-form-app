package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingToken = errors.New("missing token")
)

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTService(secret string, expiration time.Duration) (*JWTService, error) {
	if secret == "" {
		return nil, errors.New("JWT secret is required")
	}

	if expiration <= 0 {
		return nil, errors.New("JWT expiration must be greater than zero")
	}

	return &JWTService{
		secret:     []byte(secret),
		expiration: expiration,
	}, nil
}

func (s *JWTService) GenerateToken(user *model.AdminUser) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}

	if user.ID == "" {
		return "", errors.New("user ID is required")
	}

	if user.Email == "" {
		return "", errors.New("user email is required")
	}

	if user.Role == "" {
		return "", errors.New("user role is required")
	}

	now := time.Now()

	claims := Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signedToken, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf(
					"%w: unexpected signing method %s",
					ErrInvalidToken,
					token.Method.Alg(),
				)
			}

			return s.secret, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("%w: missing subject", ErrInvalidToken)
	}

	if claims.Role != model.AdminRole {
		return nil, fmt.Errorf("%w: invalid role", ErrInvalidToken)
	}

	return claims, nil
}

func (s *JWTService) Expiration() time.Duration {
	return s.expiration
}
