package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if err := createMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Prevent multiple application instances from running migrations
	// at the same time.
	if _, err := pool.Exec(ctx, `SELECT pg_advisory_lock(7483921)`); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = pool.Exec(context.Background(), `SELECT pg_advisory_unlock(7483921)`)
	}()

	entries, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		if err := applyMigration(ctx, pool, entry.Name()); err != nil {
			return err
		}
	}

	return nil
}

func createMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`

	_, err := pool.Exec(ctx, query)

	return err
}

func applyMigration(
	ctx context.Context,
	pool *pgxpool.Pool,
	filename string,
) error {
	var applied bool

	err := pool.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE version = $1
		)`,
		filename,
	).Scan(&applied)

	if err != nil {
		return fmt.Errorf(
			"check migration %s: %w",
			filename,
			err,
		)
	}

	if applied {
		fmt.Printf("migration %s already applied\n", filename)
		return nil
	}

	sqlBytes, err := os.ReadFile(
		"migrations/" + filename,
	)
	if err != nil {
		return fmt.Errorf(
			"read migration %s: %w",
			filename,
			err,
		)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin migration %s: %w",
			filename,
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf(
			"execute migration %s: %w",
			filename,
			err,
		)
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO schema_migrations (version)
		 VALUES ($1)`,
		filename,
	); err != nil {
		return fmt.Errorf(
			"record migration %s: %w",
			filename,
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit migration %s: %w",
			filename,
			err,
		)
	}

	fmt.Printf("migration %s applied\n", filename)

	return nil
}
