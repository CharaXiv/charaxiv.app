package main

import (
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"charaxiv/templates"
)

// HTML wraps a templ.Component handler, setting the Content-Type header.
func HTML(c func(r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		c(r).Render(r.Context(), w)
	}
}

func main() {
	// Trigger live reload when server starts (dev mode)
	if os.Getenv("DEV") == "1" {
		go http.Get("http://localhost:8001/reload")
	}

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

	// Character sheet
	r.Get("/", HTML(func(r *http.Request) templ.Component {
		ctx := templates.NewPageContext()
		return templates.CharacterSheet(ctx)
	}))

	// Preview mode toggle - returns targeted fragments with OOB swaps
	r.Post("/api/preview/on", HTML(func(r *http.Request) templ.Component {
		ctx := templates.PageContext{IsOwner: false}
		return templates.PreviewModeFragments(ctx)
	}))

	r.Post("/api/preview/off", HTML(func(r *http.Request) templ.Component {
		ctx := templates.PageContext{IsOwner: true}
		return templates.PreviewModeFragments(ctx)
	}))

	port := "8000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
