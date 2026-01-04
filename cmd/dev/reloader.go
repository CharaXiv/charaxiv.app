package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev mode
	},
}

// buildID increments on each reload to help clients detect stale connections
var buildID atomic.Uint64

func reloader(wg *sync.WaitGroup, sigCh <-chan os.Signal) {
	defer wg.Done()

	targets := make(map[string]chan string)
	var mu sync.Mutex

	// Debounce reload requests
	var reloadDebounce *time.Timer
	var debounceMu sync.Mutex

	mux := http.NewServeMux()

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("%s[RELOADER] WebSocket upgrade failed: %v%s\n", colorRed, err, colorReset)
			return
		}
		defer conn.Close()

		fmt.Printf("%s[RELOADER] Browser connected...%s\n", colorBlue, colorReset)
		id := uuid.New().String()
		ch := make(chan string, 1)

		mu.Lock()
		targets[id] = ch
		mu.Unlock()

		defer func() {
			mu.Lock()
			delete(targets, id)
			mu.Unlock()
			close(ch)
		}()

		// Send initial build ID so client knows current version
		initialMsg := fmt.Sprintf("build:%d", buildID.Load())
		if err := conn.WriteMessage(websocket.TextMessage, []byte(initialMsg)); err != nil {
			return
		}

		// Heartbeat ticker - sends heartbeat every 5 seconds
		heartbeatTicker := time.NewTicker(5 * time.Second)
		defer heartbeatTicker.Stop()

		// Ping ticker to keep connection alive at protocol level
		pingTicker := time.NewTicker(30 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case msg := <-ch:
				fmt.Printf("%s[RELOADER] Notifying browser: %s%s\n", colorBlue, msg, colorReset)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					return
				}
			case <-heartbeatTicker.C:
				if err := conn.WriteMessage(websocket.TextMessage, []byte("heartbeat")); err != nil {
					return
				}
			case <-pingTicker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	})

	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		// Debounce: if another reload comes within 100ms, reset the timer
		debounceMu.Lock()
		if reloadDebounce != nil {
			reloadDebounce.Stop()
		}

		done := make(chan bool, 1)
		reloadDebounce = time.AfterFunc(100*time.Millisecond, func() {
			performReload(&mu, targets)
			select {
			case done <- true:
			default:
			}
		})
		debounceMu.Unlock()

		// Wait for debounced reload to complete (with timeout)
		select {
		case <-done:
			w.WriteHeader(http.StatusOK)
		case <-time.After(35 * time.Second):
			http.Error(w, "reload timeout", http.StatusGatewayTimeout)
		}
	})

	server := &http.Server{Addr: ":8001", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("%s[RELOADER] Server error: %v%s\n", colorRed, err, colorReset)
		}
	}()

	fmt.Printf("%s[RELOADER] Server started on :8001%s\n", colorBlue, colorReset)

	<-sigCh
	fmt.Printf("%s[RELOADER] Shutting down...%s\n", colorBlue, colorReset)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	fmt.Printf("%s[RELOADER] Shutdown complete%s\n", colorBlue, colorReset)
}

func performReload(mu *sync.Mutex, targets map[string]chan string) {
	fmt.Printf("%s[RELOADER] Reload requested, polling server...%s\n", colorBlue, colorReset)

	client := http.Client{Timeout: 1 * time.Second}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// 30 second timeout for health check polling
	timeout := time.After(30 * time.Second)

	// Poll the main server's health endpoint until it responds
	for {
		select {
		case <-timeout:
			fmt.Printf("%s[RELOADER] Timeout waiting for server health%s\n", colorRed, colorReset)
			return
		case <-ticker.C:
			resp, err := client.Get("http://localhost:8000/health")
			if err != nil {
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == 200 {
				// Increment build ID
				newBuildID := buildID.Add(1)
				fmt.Printf("%s[RELOADER] Server is ready (build %d), notifying browsers...%s\n", colorBlue, newBuildID, colorReset)

				msg := fmt.Sprintf("reload:%d", newBuildID)
				mu.Lock()
				for _, target := range targets {
					select {
					case target <- msg:
					default:
					}
				}
				mu.Unlock()
				return
			}
		}
	}
}
