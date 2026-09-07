package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

type ContactRepository interface {
	Save(ctx context.Context, c *model.ContactMessage) error
	List(ctx context.Context, limit int) ([]model.ContactMessage, error)
	Ping(ctx context.Context) error
}

type AdminUserRepository interface {
	Create(ctx context.Context, user *model.AdminUser) error
	FindByEmail(ctx context.Context, email string) (*model.AdminUser, error)
	FindByID(ctx context.Context, id string) (*model.AdminUser, error)
}

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
}
