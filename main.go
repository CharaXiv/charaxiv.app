package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"charaxiv/models"
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

// statusToTemplates converts model status to template types
func statusToTemplates(status *models.Cthulhu6Status) ([]templates.StatusVariable, []templates.ComputedValue, []templates.StatusParameter, string) {
	// Variables in display order
	varOrder := []string{"STR", "CON", "POW", "DEX", "APP", "SIZ", "INT", "EDU"}
	variables := make([]templates.StatusVariable, 0, len(varOrder))
	for _, key := range varOrder {
		v := status.Variables[key]
		variables = append(variables, templates.StatusVariable{
			Key:  key,
			Base: v.Base,
			Perm: v.Perm,
			Temp: v.Temp,
			Min:  v.Min,
			Max:  v.Max,
		})
	}

	// Computed values in display order
	computedOrder := []string{"初期SAN", "アイデア", "幸運", "知識", "職業P", "興味P"}
	computedMap := status.ComputedValues()
	computed := make([]templates.ComputedValue, 0, len(computedOrder))
	for _, key := range computedOrder {
		computed = append(computed, templates.ComputedValue{
			Key:   key,
			Value: computedMap[key],
		})
	}

	// Parameters
	paramOrder := []string{"HP", "MP", "SAN"}
	defaults := status.DefaultParameters()
	parameters := make([]templates.StatusParameter, 0, len(paramOrder))
	for _, key := range paramOrder {
		var val *int
		if v := status.Parameters[key]; v != nil {
			val = v
		}
		parameters = append(parameters, templates.StatusParameter{
			Key:          key,
			Value:        val,
			DefaultValue: defaults[key],
		})
	}

	return variables, computed, parameters, status.DamageBonus()
}

// proxyWebSocket proxies a WebSocket connection to the target host.
func proxyWebSocket(w http.ResponseWriter, r *http.Request, targetHost string) {
	// Hijack the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "WebSocket hijack not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Connect to target
	targetConn, err := net.Dial("tcp", targetHost)
	if err != nil {
		return
	}
	defer targetConn.Close()

	// Forward the original request with path stripped
	path := strings.TrimPrefix(r.URL.Path, "/dev")
	if path == "" {
		path = "/"
	}

	// Write the HTTP upgrade request to target
	fmt.Fprintf(targetConn, "%s %s HTTP/1.1\r\n", r.Method, path)
	fmt.Fprintf(targetConn, "Host: %s\r\n", targetHost)
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(targetConn, "%s: %s\r\n", key, value)
		}
	}
	fmt.Fprintf(targetConn, "\r\n")

	// Bidirectional copy
	done := make(chan struct{})
	go func() {
		io.Copy(targetConn, clientConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, targetConn)
		done <- struct{}{}
	}()
	<-done
}

func main() {
	// Dev mode: notify reloader when server is ready (after ListenAndServe starts)
	devMode := os.Getenv("DEV") == "1"

	// Initialize character store (in-memory for now)
	charStore := models.NewStore()

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

	// Dev mode: proxy to reloader (including WebSocket)
	if os.Getenv("DEV") == "1" {
		reloaderURL, _ := url.Parse("http://localhost:8001")
		reloaderProxy := httputil.NewSingleHostReverseProxy(reloaderURL)

		// Custom director to strip /dev prefix
		originalDirector := reloaderProxy.Director
		reloaderProxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/dev")
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
		}

		// Handle WebSocket upgrade by copying headers
		reloaderProxy.ModifyResponse = func(resp *http.Response) error {
			return nil
		}

		r.HandleFunc("/dev/*", func(w http.ResponseWriter, r *http.Request) {
			// For WebSocket, we need to handle upgrade
			if r.Header.Get("Upgrade") == "websocket" {
				proxyWebSocket(w, r, "localhost:8001")
				return
			}
			reloaderProxy.ServeHTTP(w, r)
		})
	}

	// Character sheet
	r.Get("/", HTML(func(r *http.Request) templ.Component {
		ctx := templates.NewPageContext()
		status := charStore.GetStatus()
		vars, computed, params, db := statusToTemplates(status)
		return templates.CharacterSheetWithStatus(ctx, vars, computed, params, db)
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

	// Status variable adjustment (e.g., STR, CON, etc.)
	r.Post("/api/status/{key}/adjust", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		// Remove "status-" prefix if present
		key = strings.TrimPrefix(key, "status-")

		deltaStr := r.URL.Query().Get("delta")
		delta := 0
		fmt.Sscanf(deltaStr, "%d", &delta)

		ctx := templates.NewPageContext()
		updated := charStore.UpdateVariableBase(key, delta)
		if updated == nil {
			// Key not found, return empty
			return templates.Empty()
		}

		// Return the full status panel (for now - could optimize to just return the row)
		status := charStore.GetStatus()
		vars, computed, params, db := statusToTemplates(status)
		return templates.StatusPanelOOB(ctx, vars, computed, params, db)
	}))

	// Status variable set (direct value from input)
	r.Post("/api/status/{key}/set", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		// Remove "status-" prefix if present
		key = strings.TrimPrefix(key, "status-")

		// Parse form value
		r.ParseForm()
		valueStr := r.FormValue("status_" + key)
		value := 0
		fmt.Sscanf(valueStr, "%d", &value)

		ctx := templates.NewPageContext()
		updated := charStore.SetVariableBase(key, value)
		if updated == nil {
			// Key not found, return empty
			return templates.Empty()
		}

		// Return the full status panel
		status := charStore.GetStatus()
		vars, computed, params, db := statusToTemplates(status)
		return templates.StatusPanelOOB(ctx, vars, computed, params, db)
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

	if devMode {
		// Use a listener so we can trigger reload AFTER server starts listening
		ln, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Server listening, triggering reload...")
		// Now trigger reload - server is definitely listening
		go http.Get("http://localhost:8001/reload")
		log.Fatal(http.Serve(ln, r))
	} else {
		log.Fatal(http.ListenAndServe(":"+port, r))
	}
}
