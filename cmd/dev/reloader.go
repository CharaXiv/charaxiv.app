package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/google/uuid"
)

func reloader(wg *sync.WaitGroup, sigCh <-chan os.Signal) {
	defer wg.Done()

	targets := make(map[string]chan bool)
	var mu sync.Mutex

	mux := http.NewServeMux()

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s[RELOADER] Browser connected...%s\n", colorBlue, colorReset)
		id := uuid.New().String()
		mu.Lock()
		targets[id] = make(chan bool)
		mu.Unlock()

		websocket.Handler(func(ws *websocket.Conn) {
			defer func() {
				mu.Lock()
				close(targets[id])
				delete(targets, id)
				mu.Unlock()
				ws.Close()
			}()

			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-targets[id]:
					fmt.Printf("%s[RELOADER] Notifying browser to reload...%s\n", colorBlue, colorReset)
					ws.Write([]byte("reload"))
				case <-ticker.C:
					ws.Write([]byte("ping"))
				}
			}
		}).ServeHTTP(w, r)
	})

	mux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s[RELOADER] Server is ready, notifying browsers...%s\n", colorBlue, colorReset)
		mu.Lock()
		for _, target := range targets {
			target <- true
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
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
