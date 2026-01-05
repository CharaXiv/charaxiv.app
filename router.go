package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"charaxiv/routes/cthulhu6"
	"charaxiv/storage/coalesce"
)

// AppRouter creates the application router with all game system routes mounted.
// This is the single source of truth for application route configuration.
// Infrastructure routes (health, static, dev proxy) are added separately in main.
func AppRouter(cs *coalesce.Store) chi.Router {
	r := chi.NewRouter()

	// Mount cthulhu6 system routes
	cthulhu6Store := cthulhu6.NewStore(cs)
	r.Mount("/cthulhu6", cthulhu6.Routes(cthulhu6Store))

	return r
}

// NewServer creates a fully configured HTTP server with all routes.
// This includes both application routes and infrastructure routes.
func NewServer(cs *coalesce.Store) chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Mount application routes
	r.Mount("/", AppRouter(cs))

	return r
}
