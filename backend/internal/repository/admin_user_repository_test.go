package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"

	"testing"
	"time"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

func TestAdminUserRepositoryCreate(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAdminUserRepository(db)

	user := &model.AdminUser{
		Email:        "admin@example.com",
		PasswordHash: "hashed-password",
		Role:         model.AdminRole,
	}

	createdAt := time.Now()
	updatedAt := createdAt

	db.ExpectQuery(`INSERT INTO admin_users`).
		WithArgs(
			user.Email,
			user.PasswordHash,
			user.Role,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				"550e8400-e29b-41d4-a716-446655440000",
				createdAt,
				updatedAt,
			),
		)

	err = repo.Create(context.Background(), user)

	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	if user.ID == "" {
		t.Error("Create() should set ID")
	}

	if user.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("Create() should set UpdatedAt")
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestAdminUserRepositoryFindByEmail(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAdminUserRepository(db)

	createdAt := time.Now()
	updatedAt := createdAt

	db.ExpectQuery(`SELECT\s+id,\s+email,\s+password_hash,\s+role,\s+created_at,\s+updated_at\s+FROM admin_users`).
		WithArgs("admin@example.com").
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"email",
				"password_hash",
				"role",
				"created_at",
				"updated_at",
			}).AddRow(
				"550e8400-e29b-41d4-a716-446655440000",
				"admin@example.com",
				"hashed-password",
				"admin",
				createdAt,
				updatedAt,
			),
		)

	user, err := repo.FindByEmail(
		context.Background(),
		"admin@example.com",
	)

	if err != nil {
		t.Fatalf("FindByEmail() unexpected error: %v", err)
	}

	if user == nil {
		t.Fatal("FindByEmail() returned nil user")
	}

	if user.Email != "admin@example.com" {
		t.Errorf(
			"Email = %q, want %q",
			user.Email,
			"admin@example.com",
		)
	}

	if user.Role != model.AdminRole {
		t.Errorf(
			"Role = %q, want %q",
			user.Role,
			model.AdminRole,
		)
	}

	if user.PasswordHash != "hashed-password" {
		t.Errorf("PasswordHash was not correctly loaded")
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestAdminUserRepositoryFindByEmailNotFound(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAdminUserRepository(db)

	db.ExpectQuery(`SELECT\s+id,\s+email,\s+password_hash,\s+role,\s+created_at,\s+updated_at\s+FROM admin_users`).
		WithArgs("unknown@example.com").
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.FindByEmail(
		context.Background(),
		"unknown@example.com",
	)

	if user != nil {
		t.Error("FindByEmail() should return nil user when not found")
	}

	if !errors.Is(err, ErrAdminUserNotFound) {
		t.Errorf(
			"error = %v, want ErrAdminUserNotFound",
			err,
		)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestAdminUserRepositoryFindByID(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAdminUserRepository(db)

	createdAt := time.Now()
	updatedAt := createdAt

	id := "550e8400-e29b-41d4-a716-446655440000"

	db.ExpectQuery(`SELECT\s+id,\s+email,\s+password_hash,\s+role,\s+created_at,\s+updated_at\s+FROM admin_users`).
		WithArgs(id).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"email",
				"password_hash",
				"role",
				"created_at",
				"updated_at",
			}).AddRow(
				id,
				"admin@example.com",
				"hashed-password",
				model.AdminRole,
				createdAt,
				updatedAt,
			),
		)

	user, err := repo.FindByID(context.Background(), id)

	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}

	if user == nil {
		t.Fatal("FindByID() returned nil user")
	}

	if user.ID != id {
		t.Errorf("ID = %q, want %q", user.ID, id)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}
