package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

func TestPostgresContactRepositorySave(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	c := &model.ContactMessage{
		Name:    "John Doe",
		Email:   "john@example.com",
		Subject: "Hello",
		Message: "Test message",
	}

	createdAt := time.Now()
	expectedID := "550e8400-e29b-41d4-a716-446655440000"

	db.ExpectQuery(`INSERT INTO contact_messages`).
		WithArgs(
			c.Name,
			c.Email,
			c.Subject,
			c.Message,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"created_at",
			}).AddRow(
				expectedID,
				createdAt,
			),
		)

	err = repo.Save(context.Background(), c)

	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if c.ID == "" {
		t.Errorf("Save() should set ID")
	}

	if c.ID != expectedID {
		t.Errorf("Save() ID = %q, want %q", c.ID, expectedID)
	}

	if c.CreatedAt.IsZero() {
		t.Error("Save() should set CreatedAt")
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositorySaveQueryError(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	c := &model.ContactMessage{
		Name:    "Jane Doe",
		Email:   "jane@example.com",
		Subject: "Hi",
		Message: "Another message",
	}

	db.ExpectQuery(`INSERT INTO contact_messages`).
		WithArgs(
			c.Name,
			c.Email,
			c.Subject,
			c.Message,
		).
		WillReturnError(errors.New("database unavailable"))

	err = repo.Save(context.Background(), c)

	if err == nil {
		t.Fatal("Save() expected error, got nil")
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositoryList(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	now := time.Now()
	id1 := "550e8400-e29b-41d4-a716-446655440001"
	id2 := "550e8400-e29b-41d4-a716-446655440002"

	db.ExpectQuery(`SELECT\s+id,\s+name,\s+email,\s+subject,\s+message,\s+created_at\s+FROM contact_messages`).
		WithArgs(2).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"name",
				"email",
				"subject",
				"message",
				"created_at",
			}).AddRow(
				id2,
				"Bob",
				"bob@example.com",
				"Subject2",
				"Message2",
				now,
			).AddRow(
				id1,
				"Alice",
				"alice@example.com",
				"Subject1",
				"Message1",
				now,
			),
		)

	messages, err := repo.List(context.Background(), 2)

	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("List() returned %d messages, want 2", len(messages))
	}

	if messages[0].ID != id2 || messages[0].Name != "Bob" {
		t.Errorf("List()[0] = %+v, unexpected", messages[0])
	}

	if messages[1].ID != id1 || messages[1].Name != "Alice" {
		t.Errorf("List()[1] = %+v, unexpected", messages[1])
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositoryListEmpty(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	db.ExpectQuery(`SELECT\s+id,\s+name,\s+email,\s+subject,\s+message,\s+created_at\s+FROM contact_messages`).
		WithArgs(10).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"name",
				"email",
				"subject",
				"message",
				"created_at",
			}),
		)

	messages, err := repo.List(context.Background(), 10)

	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(messages) != 0 {
		t.Fatalf("List() returned %d messages, want 0", len(messages))
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositoryListQueryError(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	db.ExpectQuery(`SELECT\s+id,\s+name,\s+email,\s+subject,\s+message,\s+created_at\s+FROM contact_messages`).
		WithArgs(5).
		WillReturnError(errors.New("connection lost"))

	messages, err := repo.List(context.Background(), 5)

	if err == nil {
		t.Fatal("List() expected error, got nil")
	}

	if messages != nil {
		t.Errorf("List() messages = %+v, want nil", messages)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositoryListScanError(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	db.ExpectQuery(`SELECT\s+id,\s+name,\s+email,\s+subject,\s+message,\s+created_at\s+FROM contact_messages`).
		WithArgs(1).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id",
				"name",
				"email",
				"subject",
				"message",
				"created_at",
			}).AddRow(
				1,
				"Alice",
				"alice@example.com",
				"Subject1",
				"Message1",
				time.Now(),
			).RowError(0, errors.New("scan failure")),
		)

	messages, err := repo.List(context.Background(), 1)

	if err == nil {
		t.Fatal("List() expected error, got nil")
	}

	if messages != nil {
		t.Errorf("List() messages = %+v, want nil", messages)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositoryPing(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	db.ExpectPing()

	err = repo.Ping(context.Background())

	if err != nil {
		t.Fatalf("Ping() unexpected error: %v", err)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresContactRepositoryPingError(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewContactRepository(db)

	db.ExpectPing().WillReturnError(errors.New("no connection"))

	err = repo.Ping(context.Background())

	if err == nil {
		t.Fatal("Ping() expected error, got nil")
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}
