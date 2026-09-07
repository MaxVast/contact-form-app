package model

import (
	"strings"
	"testing"
	"time"
)

func TestAdminUserValidate(t *testing.T) {
	tests := []struct {
		name      string
		user      AdminUser
		wantError bool
	}{
		{
			name: "valid admin user",
			user: AdminUser{
				Email:        "admin@example.com",
				PasswordHash: "hashed-password",
				Role:         AdminRole,
			},
			wantError: false,
		},
		{
			name: "email is trimmed and normalized",
			user: AdminUser{
				Email:        "  ADMIN@EXAMPLE.COM  ",
				PasswordHash: "hashed-password",
				Role:         AdminRole,
			},
			wantError: false,
		},
		{
			name: "empty email",
			user: AdminUser{
				Email:        "",
				PasswordHash: "hashed-password",
				Role:         AdminRole,
			},
			wantError: true,
		},
		{
			name: "invalid email",
			user: AdminUser{
				Email:        "admin-example.com",
				PasswordHash: "hashed-password",
				Role:         AdminRole,
			},
			wantError: true,
		},
		{
			name: "email too long",
			user: AdminUser{
				Email:        strings.Repeat("a", 151),
				PasswordHash: "hashed-password",
				Role:         AdminRole,
			},
			wantError: true,
		},
		{
			name: "empty password hash",
			user: AdminUser{
				Email:        "admin@example.com",
				PasswordHash: "",
				Role:         AdminRole,
			},
			wantError: true,
		},
		{
			name: "empty role defaults to admin",
			user: AdminUser{
				Email:        "admin@example.com",
				PasswordHash: "hashed-password",
				Role:         "",
			},
			wantError: false,
		},
		{
			name: "invalid role",
			user: AdminUser{
				Email:        "admin@example.com",
				PasswordHash: "hashed-password",
				Role:         "user",
			},
			wantError: true,
		},
		{
			name: "role is normalized",
			user: AdminUser{
				Email:        "admin@example.com",
				PasswordHash: "hashed-password",
				Role:         " ADMIN ",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.wantError && err == nil {
				t.Fatalf("Validate() expected an error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestAdminUserValidateNormalizesEmail(t *testing.T) {
	user := AdminUser{
		Email:        "  ADMIN@EXAMPLE.COM  ",
		PasswordHash: "hashed-password",
		Role:         AdminRole,
	}

	if err := user.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if user.Email != "admin@example.com" {
		t.Errorf(
			"Email = %q, want %q",
			user.Email,
			"admin@example.com",
		)
	}
}

func TestAdminUserValidateDefaultsRole(t *testing.T) {
	user := AdminUser{
		Email:        "admin@example.com",
		PasswordHash: "hashed-password",
	}

	if err := user.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if user.Role != AdminRole {
		t.Errorf(
			"Role = %q, want %q",
			user.Role,
			AdminRole,
		)
	}
}

func TestAdminUserMetadata(t *testing.T) {
	now := time.Now()

	user := AdminUser{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "admin@example.com",
		PasswordHash: "hashed-password",
		Role:         AdminRole,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if user.ID == "" {
		t.Error("ID should not be empty")
	}

	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}
}
