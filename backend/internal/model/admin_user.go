package model

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

var ErrAdminUserValidation = errors.New("admin user validation")

type AdminUser struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const (
	AdminRole = "admin"
)

func (u *AdminUser) Validate() error {
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.Role = strings.TrimSpace(strings.ToLower(u.Role))

	if u.Email == "" {
		return fmt.Errorf("%w: email requis", ErrAdminUserValidation)
	}

	if _, err := mail.ParseAddress(u.Email); err != nil {
		return fmt.Errorf("%w: email invalide", ErrAdminUserValidation)
	}

	if len(u.Email) > 150 {
		return fmt.Errorf("%w: email trop long", ErrAdminUserValidation)
	}

	if !strings.Contains(u.Email, "@") {
		return fmt.Errorf("%w: email invalide", ErrAdminUserValidation)
	}

	if u.PasswordHash == "" {
		return fmt.Errorf("%w: password hash requis", ErrAdminUserValidation)
	}

	if u.Role == "" {
		u.Role = AdminRole
	}

	if u.Role != AdminRole {
		return fmt.Errorf("%w: rôle invalide", ErrAdminUserValidation)
	}

	return nil
}
