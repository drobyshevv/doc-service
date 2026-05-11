package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/drobyshevv/doc-service/internal/auth/config"
	authHandler "github.com/drobyshevv/doc-service/internal/auth/handler"
	"github.com/drobyshevv/doc-service/internal/auth/jwt"
	authMiddleware "github.com/drobyshevv/doc-service/internal/auth/middleware"
	authService "github.com/drobyshevv/doc-service/internal/auth/service"
	"github.com/drobyshevv/doc-service/internal/auth/storage/postgres"
)

func main() {
	cfg, err := config.LoadConfig("configs/auth.yaml")
	if err != nil {
		log.Fatal("config error:", err)
	}

	db, err := sql.Open("postgres", cfg.DBConnStr())
	if err != nil {
		log.Fatal("db error:", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("db ping error:", err)
	}

	userRepo := postgres.NewUserRepository(db)
	refreshRepo := postgres.NewRefreshTokenRepository(db)

	jwtManager := jwt.NewManager(
		cfg.JWT.AccessSecret,
		time.Duration(cfg.JWT.AccessTTLMinutes)*time.Minute,
		time.Duration(cfg.JWT.RefreshTTLDays)*24*time.Hour,
	)

	authSvc := authService.NewAuthService(userRepo, refreshRepo, jwtManager)
	userSvc := authService.NewUserService(userRepo)

	authH := authHandler.NewAuthHandler(authSvc)
	userH := authHandler.NewUserHandler(userSvc)

	authMW := authMiddleware.NewAuthMiddleware(jwtManager)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMW.Middleware)

		r.Route("/users", func(r chi.Router) {
			r.Get("/{id}", userH.GetByID)
		})
	})

	addr := cfg.HTTPAddr()
	log.Println("auth service running on", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
