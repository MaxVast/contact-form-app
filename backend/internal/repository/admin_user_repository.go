package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

var ErrAdminUserNotFound = errors.New("admin user not found")

type PostgresAdminUserRepository struct {
	db DB
}

func NewAdminUserRepository(db DB) *PostgresAdminUserRepository {
	return &PostgresAdminUserRepository{
		db: db,
	}
}

func (r *PostgresAdminUserRepository) Create(ctx context.Context, user *model.AdminUser) error {
	query := `
		INSERT INTO admin_users (email,password_hash,role)
				VALUES ($1, $2, $3)
				RETURNING id, created_at, updated_at
				`

	err := r.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	return nil
}

func (r *PostgresAdminUserRepository) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM admin_users
		WHERE email = $1
		`

	var user model.AdminUser

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAdminUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find admin user by email: %w", err)
	}

	return &user, nil
}

func (r *PostgresAdminUserRepository) FindByID(ctx context.Context, id string) (*model.AdminUser, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM admin_users
		WHERE id = $1
		`

	var user model.AdminUser

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAdminUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find admin user by id: %w", err)
	}

	return &user, nil
}
