package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/maxvast/contact-form-app/backend/internal/config"
	"github.com/maxvast/contact-form-app/backend/internal/database"
	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/repository"
	"github.com/maxvast/contact-form-app/backend/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connexion à la base impossible: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	email := strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")

	if email == "" {
		fmt.Fprintln(os.Stderr, "ADMIN_EMAIL est requis")
		os.Exit(1)
	}

	if password == "" {
		fmt.Fprintln(os.Stderr, "ADMIN_PASSWORD est requis")
		os.Exit(1)
	}

	passwordHash, err := service.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash du mot de passe impossible: %v\n", err)
		os.Exit(1)
	}

	user := &model.AdminUser{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         model.AdminRole,
	}

	if err := user.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "administrateur invalide: %v\n", err)
		os.Exit(1)
	}

	adminRepository := repository.NewAdminUserRepository(pool)

	if err := adminRepository.Create(ctx, user); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			fmt.Printf(
				"seed-admin: admin %s existe déjà, aucune modification\n",
				email,
			)
			return
		}

		fmt.Fprintf(
			os.Stderr,
			"création de l'administrateur impossible: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Printf("Admin user created successfully: %s\n", user.Email)
}
