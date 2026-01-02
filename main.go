package main

import (
	"log"
	"net/http"
	"os"

	"charaxiv/templates"
)

func main() {
	// Trigger live reload when server starts (dev mode)
	if os.Getenv("DEV") == "1" {
		go http.Get("http://localhost:8001/reload")
	}

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Character sheet
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		templates.CharacterSheet().Render(r.Context(), w)
	})

	port := "8000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
