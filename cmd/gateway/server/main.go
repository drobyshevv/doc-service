package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/drobyshevv/doc-service/internal/auth/jwt"
	authmw "github.com/drobyshevv/doc-service/internal/auth/middleware"
)

func newProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// original host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Origin-Host", u.Host)
	}

	return proxy
}

func main() {
	authProxy := newProxy("http://auth:8082")
	mainProxy := newProxy("http://mainserv:8081")

	jwtManager := jwt.NewManager(
		"super-secret-access",
		15*time.Minute,
		30*24*time.Hour,
	)

	authMiddleware := authmw.NewAuthMiddleware(jwtManager)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// PUBLIC AUTH
	r.Mount("/auth", authProxy)

	// PUBLIC SEARCH
	r.Group(func(r chi.Router) {
		r.Handle("/search", mainProxy)
		r.Handle("/search/", mainProxy)
		r.Handle("/search/phrase", mainProxy)
		r.Handle("/search/suggest", mainProxy)
		r.Handle("/health", mainProxy)
	})

	// PROTECTED ROUTES
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)

		r.Handle("GET /documents", mainProxy)

		r.Handle("/documents/*", mainProxy)
		r.Handle("/search/owner", mainProxy)
	})

	log.Println("gateway started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
