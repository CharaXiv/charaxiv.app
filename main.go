package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"charaxiv/storage"
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

	// Initialize storage
	var store storage.Storage
	if bucket := os.Getenv("GCS_BUCKET"); bucket != "" {
		ctx := context.Background()
		client, err := storage.NewGCSClient(ctx, storage.GCSConfig{
			Bucket: bucket,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize GCS client: %v", err)
			log.Printf("Falling back to in-memory storage")
			store = storage.NewMemoryStorage()
		} else {
			store = client
			log.Printf("GCS storage initialized for bucket: %s", bucket)
		}
	} else {
		log.Printf("GCS_BUCKET not set, using in-memory storage")
		store = storage.NewMemoryStorage()
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
		ctx := templates.PageContext{IsOwner: true, Preview: true}
		return templates.PreviewModeFragments(ctx)
	}))

	r.Post("/api/preview/off", HTML(func(r *http.Request) templ.Component {
		ctx := templates.PageContext{IsOwner: true, Preview: false}
		return templates.PreviewModeFragments(ctx)
	}))

	// Storage test endpoint
	r.Get("/api/storage/test", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		testKey := "test/hello.txt"
		testContent := "Hello from storage! " + time.Now().Format(time.RFC3339)

		// Upload
		err := store.Upload(ctx, testKey, strings.NewReader(testContent), "text/plain")
		if err != nil {
			http.Error(w, "Upload failed: "+err.Error(), 500)
			return
		}

		// Download
		reader, err := store.Download(ctx, testKey)
		if err != nil {
			http.Error(w, "Download failed: "+err.Error(), 500)
			return
		}
		defer reader.Close()

		downloaded, _ := io.ReadAll(reader)

		// Generate signed URL (may not be supported by all implementations)
		signedURL, signedErr := store.SignedURL(ctx, testKey, 15*time.Minute)
		if signedErr != nil {
			signedURL = "(not supported: " + signedErr.Error() + ")"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":     "ok",
			"uploaded":   testContent,
			"downloaded": string(downloaded),
			"signedURL":  signedURL,
			"publicURL":  store.PublicURL(testKey),
		})
	})

	port := "8000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
