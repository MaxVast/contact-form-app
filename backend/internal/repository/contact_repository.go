package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

type ContactRepository struct {
	db *pgxpool.Pool
}

func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Save(ctx context.Context, c *model.ContactMessage) error {
	query := `
		INSERT INTO contact_messages (name, email, subject, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.db.QueryRow(ctx, query, c.Name, c.Email, c.Subject, c.Message).
		Scan(&c.ID, &c.CreatedAt)
}

func (r *ContactRepository) List(ctx context.Context, limit int) ([]model.ContactMessage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, email, subject, message, created_at
		FROM contact_messages
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.ContactMessage
	for rows.Next() {
		var m model.ContactMessage
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Subject, &m.Message, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *ContactRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
