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
	"path/filepath"
	"strings"
	"time"

	"charaxiv/storage"
	"charaxiv/storage/coalesce"
)

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

	// Initialize coalesce store for character data
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	cs, err := coalesce.New(coalesce.Config{
		DBPath:  filepath.Join(dataDir, "buffer.db"),
		DataDir: filepath.Join(dataDir, "characters"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize coalesce store: %v", err)
	}
	defer cs.Close()
	log.Printf("Coalesce store initialized: %s", dataDir)

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

	// Build server with application routes
	r := NewServer(cs)

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
