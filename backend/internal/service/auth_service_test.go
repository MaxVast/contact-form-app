package service

import (
	"context"
	"errors"
	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/repository"
	"strings"
	"testing"
)

type fakeAdminUserRepository struct {
	user  *model.AdminUser
	err   error
	email string
}

func (f *fakeAdminUserRepository) Create(ctx context.Context, user *model.AdminUser) error {
	return nil
}

func (f *fakeAdminUserRepository) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	f.email = email

	if f.err != nil {
		return nil, f.err
	}

	return f.user, nil
}

func (f *fakeAdminUserRepository) FindByID(ctx context.Context, id string) (*model.AdminUser, error) {
	return f.user, nil
}

func TestAuthenticate_Success(t *testing.T) {
	password := "SuperSecret123!"

	passwordHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	expectedUser := &model.AdminUser{
		ID:           "admin-id",
		Email:        "admin@example.com",
		PasswordHash: passwordHash,
		Role:         model.AdminRole,
	}

	repo := &fakeAdminUserRepository{
		user: expectedUser,
	}

	service := NewAuthService(repo)

	user, err := service.Authenticate(
		context.Background(),
		" ADMIN@EXAMPLE.COM ",
		password,
	)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if user != expectedUser {
		t.Fatalf("Authenticate() user = %v, want %v", user, expectedUser)
	}

	if repo.email != "admin@example.com" {
		t.Fatalf(
			"repository email = %q, want %q",
			repo.email,
			"admin@example.com",
		)
	}
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	passwordHash, err := HashPassword("CorrectPassword123!")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	repo := &fakeAdminUserRepository{
		user: &model.AdminUser{
			ID:           "admin-id",
			Email:        "admin@example.com",
			PasswordHash: passwordHash,
			Role:         model.AdminRole,
		},
	}

	service := NewAuthService(repo)

	user, err := service.Authenticate(
		context.Background(),
		"admin@example.com",
		"WrongPassword123!",
	)

	if user != nil {
		t.Fatalf("Authenticate() user = %v, want nil", user)
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"Authenticate() error = %v, want ErrInvalidCredentials",
			err,
		)
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	repo := &fakeAdminUserRepository{
		err: repository.ErrAdminUserNotFound,
	}

	service := NewAuthService(repo)

	user, err := service.Authenticate(
		context.Background(),
		"unknown@example.com",
		"SomePassword123!",
	)

	if user != nil {
		t.Fatalf("Authenticate() user = %v, want nil", user)
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"Authenticate() error = %v, want ErrInvalidCredentials",
			err,
		)
	}
}

func TestAuthenticate_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")

	repo := &fakeAdminUserRepository{
		err: repositoryErr,
	}

	service := NewAuthService(repo)

	user, err := service.Authenticate(
		context.Background(),
		"admin@example.com",
		"SomePassword123!",
	)

	if user != nil {
		t.Fatalf("Authenticate() user = %v, want nil", user)
	}

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"Authenticate() error = %v, want repository error",
			err,
		)
	}
}

func TestAuthenticate_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "empty email",
			email:    "",
			password: "Password123!",
		},
		{
			name:     "empty password",
			email:    "admin@example.com",
			password: "",
		},
		{
			name:     "both empty",
			email:    "",
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAdminUserRepository{}

			service := NewAuthService(repo)

			user, err := service.Authenticate(
				context.Background(),
				tt.email,
				tt.password,
			)

			if user != nil {
				t.Fatalf("Authenticate() user = %v, want nil", user)
			}

			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf(
					"Authenticate() error = %v, want ErrInvalidCredentials",
					err,
				)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "SuperSecretPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned an empty hash")
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected Argon2id hash, got %q", hash)
	}

	if hash == password {
		t.Fatal("password must not be stored in plain text")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SuperSecretPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if !valid {
		t.Fatal("expected password to be valid")
	}
}

func TestVerifyPassword_InvalidPassword(t *testing.T) {
	password := "SuperSecretPassword123!"
	wrongPassword := "WrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	valid, err := VerifyPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if valid {
		t.Fatal("expected password to be invalid")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "SuperSecretPassword123!"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("expected different hashes because salts should be random")
	}
}
