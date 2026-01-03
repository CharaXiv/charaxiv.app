package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev mode
	},
}

func reloader(wg *sync.WaitGroup, sigCh <-chan os.Signal) {
	defer wg.Done()

	targets := make(map[string]chan bool)
	var mu sync.Mutex

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
		ch := make(chan bool, 1)

		mu.Lock()
		targets[id] = ch
		mu.Unlock()

		defer func() {
			mu.Lock()
			delete(targets, id)
			mu.Unlock()
			close(ch)
		}()

		// Ping ticker to keep connection alive
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ch:
				fmt.Printf("%s[RELOADER] Notifying browser to reload...%s\n", colorBlue, colorReset)
				if err := conn.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
					return
				}
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	})

	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s[RELOADER] Reload requested, polling server...%s\n", colorBlue, colorReset)

		client := http.Client{Timeout: 1 * time.Second}
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		// Poll the main server's health endpoint until it responds
		for {
			select {
			case <-ticker.C:
				resp, err := client.Get("http://localhost:8000/health")
				if err != nil {
					continue
				}
				resp.Body.Close()

				if resp.StatusCode == 200 {
					fmt.Printf("%s[RELOADER] Server is ready, notifying browsers...%s\n", colorBlue, colorReset)
					mu.Lock()
					for _, target := range targets {
						select {
						case target <- true:
						default:
						}
					}
					mu.Unlock()
					w.WriteHeader(http.StatusOK)
					return
				}
			}
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
