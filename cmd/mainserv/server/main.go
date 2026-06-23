package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/drobyshevv/doc-service/internal/mainserv/config"
	"github.com/drobyshevv/doc-service/internal/mainserv/handler"
	"github.com/drobyshevv/doc-service/internal/mainserv/service"
	"github.com/drobyshevv/doc-service/internal/mainserv/storage/postgres"
	redisstorage "github.com/drobyshevv/doc-service/internal/mainserv/storage/redis"
	s3storage "github.com/drobyshevv/doc-service/internal/mainserv/storage/s3"
)

// main является точкой входа mainserv.
//
// Сервис отвечает за:
//   - загрузку и хранение документов
//   - построение поискового индекса
//   - полнотекстовый поиск по документам
//
// При запуске инициализируются:
//   - PostgreSQL connection pool
//   - S3 storage client
//   - repositories
//   - services
//   - HTTP handlers
//
// HTTP сервер запускается с поддержкой graceful shutdown.
func main() {
	cfg, err := config.LoadConfig("configs/mainserv.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	s3Client, err := s3storage.NewClient(cfg.S3)
	if err != nil {
		log.Fatalf("init s3 client: %v", err)
	}

	storage := s3storage.NewStorage(s3Client)

	redisClient := redisstorage.NewClient(config.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("redis not available: %v", err)
	}

	documentRepo := postgres.NewDocumentRepository(db)
	searchRepo := postgres.NewSearchRepository(db)

	documentService := service.NewDocumentService(
		documentRepo,
		searchRepo,
		storage,
		redisClient,
	)

	searchService := service.NewSearchService(
		searchRepo,
		documentRepo,
		redisClient,
	)

	documentHandler := handler.NewDocumentHandler(documentService)
	searchHandler := handler.NewSearchHandler(searchService)

	router := chi.NewRouter()

	router.Get("/health", handler.Health)

	router.Route("/documents", func(r chi.Router) {
		r.Post("/upload", documentHandler.UploadDocument)
		r.Get("/{id}", documentHandler.GetDocument)
		r.Delete("/{id}", documentHandler.DeleteDocument)

		r.Post("/{id}/versions", documentHandler.UploadNewVersion)
		r.Get("/{id}/versions", documentHandler.ListVersions)
		r.Get("/{id}/versions/{version}", documentHandler.GetVersion)
		r.Post("/{id}/versions/{version}/rollback", documentHandler.RollbackVersion)

		r.Get("/", documentHandler.ListDocuments)
		r.Put("/{id}", documentHandler.UpdateMetadata)
		r.Get("/{id}/meta", documentHandler.GetMetadata)
	})

	router.Route("/search", func(r chi.Router) {
		r.Get("/", searchHandler.Search)
		r.Get("/title", searchHandler.SearchByTitle)
		r.Get("/owner", searchHandler.SearchByOwner)
		r.Get("/suggest", searchHandler.Suggest)
	})

	router.Get("/public/documents", documentHandler.ListPublicDocuments)

	server := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("mainserv started on %s", server.Addr)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop
	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	log.Println("server stopped gracefully")
}
