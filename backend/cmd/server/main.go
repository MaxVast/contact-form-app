package main

import (
	"context"
	"errors"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/maxvast/contact-form-app/backend/internal/auth"
	"github.com/maxvast/contact-form-app/backend/internal/config"
	"github.com/maxvast/contact-form-app/backend/internal/database"
	"github.com/maxvast/contact-form-app/backend/internal/handler"
	"github.com/maxvast/contact-form-app/backend/internal/middleware"
	"github.com/maxvast/contact-form-app/backend/internal/observability"
	"github.com/maxvast/contact-form-app/backend/internal/repository"
	"github.com/maxvast/contact-form-app/backend/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connexion à la base impossible: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migration impossible: %v", err)
	}

	jwtExpiration, err := time.ParseDuration(cfg.JWTExpiration)
	if err != nil {
		log.Fatalf("durée JWT invalide: %v", err)
	}

	jwtService, err := auth.NewJWTService(cfg.JWTSecret, jwtExpiration)
	if err != nil {
		log.Fatalf("configuration JWT invalide: %v", err)
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	contactRepository := repository.NewContactRepository(pool)
	contactService := service.NewContactService(contactRepository)
	contactHandler := handler.NewContactHandler(contactService)

	adminUserRepository := repository.NewAdminUserRepository(pool)
	authService := service.NewAuthService(adminUserRepository)

	authHandler := handler.NewAuthHandler(authService, jwtService)

	rateLimiter := middleware.NewRateLimiter(2.0/60.0, 2)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(observability.Metrics)
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{cfg.CORSOrigin},
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
		r.Get("/health", contactHandler.Health)
		r.Get("/ready", contactHandler.Ready)
		r.Route("/api/contact", func(r chi.Router) {
			r.With(rateLimiter.Middleware).Post("/", contactHandler.Create)
		})
		r.Route("/api/admin", func(r chi.Router) {
			r.With(rateLimiter.Middleware).Post("/login", authHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(jwtMiddleware.Authenticate)
				r.Get("/me", authHandler.Me)
				r.Get("/messages", contactHandler.List)
			})
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("contact-service démarré sur le port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("erreur serveur: %v", err)
		}
	}()

	// Arrêt propre sur SIGTERM (important pour Kubernetes : le pod reçoit
	// un SIGTERM avant d'être tué, il faut finir les requêtes en cours).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("arrêt en cours...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
