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
	"charaxiv/systems/cthulhu6"
	"charaxiv/templates/components"
	"charaxiv/templates/pages"
	"charaxiv/templates/shared"
)

// HTML wraps a templ.Component handler, setting the Content-Type header.
func HTML(c func(r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		c(r).Render(r.Context(), w)
	}
}

// buildPageContext creates a PageContext with memos loaded from the store
func buildPageContext(store *models.Store) shared.PageContext {
	ctx := shared.NewPageContext()
	// Load all memos
	memoIDs := []string{"public-memo", "secret-memo", "scenario-public-memo", "scenario-secret-memo"}
	for _, id := range memoIDs {
		ctx.Memos[id] = store.GetMemo(id)
	}
	return ctx
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
		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return pages.Cthulhu6Sheet(state)
	}))

	// Preview mode toggle - returns targeted fragments with OOB swaps
	r.Post("/api/preview/on", HTML(func(r *http.Request) templ.Component {
		ctx := buildPageContext(charStore)
		ctx.Preview = true
		return pages.Cthulhu6PreviewModeFragments(ctx)
	}))

	r.Post("/api/preview/off", HTML(func(r *http.Request) templ.Component {
		ctx := buildPageContext(charStore)
		ctx.Preview = false
		return pages.Cthulhu6PreviewModeFragments(ctx)
	}))

	// Status variable set (direct value from input)
	r.Post("/api/status/{key}/set", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		// Remove "status-" prefix if present
		key = strings.TrimPrefix(key, "status-")

		// Parse form value
		r.ParseForm()
		valueStr := r.FormValue("status_" + key)
		if valueStr == "" {
			// Empty input, no update needed
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var value int
		_, err := fmt.Sscanf(valueStr, "%d", &value)
		if err != nil {
			// Invalid number, no update needed
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Check if value actually changed
		status := charStore.GetStatus()
		if v, ok := status.Variables[key]; ok {
			// Clamp to bounds for comparison
			clampedValue := value
			if clampedValue < v.Min {
				clampedValue = v.Min
			}
			if clampedValue > v.Max {
				clampedValue = v.Max
			}
			if v.Base == clampedValue {
				// No change, no update needed
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		updated := charStore.SetVariableBase(key, value)
		if updated == nil {
			// Key not found
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Return the full status panel
		pc := shared.NewPageContext()
		status = charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// For DEX/EDU changes, also update skills panel (回避 depends on DEX, 母国語 depends on EDU)
		// For INT/EDU changes, also update skill points display
		switch key {
		case "DEX", "EDU":
			components.Cthulhu6StatusPanelWithSkills(state).Render(r.Context(), w)
		case "INT":
			components.Cthulhu6StatusPanelWithPoints(state).Render(r.Context(), w)
		default:
			components.Cthulhu6StatusPanel(state, true).Render(r.Context(), w)
		}
	})

	// Memo update endpoint
	r.Post("/api/memo/{id}/set", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		// Parse form value
		r.ParseForm()
		value := r.FormValue(id)

		// Update memo, return 204 if unchanged
		if !charStore.SetMemo(id, value) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Return 204 - memo saved, no content update needed
		// (the editor already has the correct value)
		w.WriteHeader(http.StatusNoContent)
	})

	// Skill grow toggle
	r.Post("/api/skill/{key}/grow", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsSingle() {
			return shared.Empty()
		}

		// Toggle grow flag
		skill.Single.Grow = !skill.Single.Grow
		charStore.UpdateSkill(key, skill)

		pc := shared.NewPageContext()
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Skill field adjustment (job, hobby, perm, temp)
	r.Post("/api/skill/{key}/{field}/adjust", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		field := chi.URLParam(r, "field")

		deltaStr := r.URL.Query().Get("delta")
		delta := 0
		fmt.Sscanf(deltaStr, "%d", &delta)

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsSingle() {
			return shared.Empty()
		}

		// Apply delta to the appropriate field
		switch field {
		case "job":
			skill.Single.Job += delta
			if skill.Single.Job < 0 {
				skill.Single.Job = 0
			}
		case "hobby":
			skill.Single.Hobby += delta
			if skill.Single.Hobby < 0 {
				skill.Single.Hobby = 0
			}
		case "perm":
			skill.Single.Perm += delta
		case "temp":
			skill.Single.Temp += delta
		}

		charStore.UpdateSkill(key, skill)

		// Build the skill for template
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		updatedSkill, _ := charStore.GetSkill(key)

		templSkill := cthulhu6.BuildSkill(status, key, updatedSkill)
		remaining := cthulhu6.BuildRemainingPoints(status, skills)

		return components.Cthulhu6SkillUpdateFragments(templSkill, field, remaining)
	}))

	// Add genre to multi-skill
	r.Post("/api/skill/{key}/genre/add", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsMulti() {
			return shared.Empty()
		}

		// Add a new empty genre
		skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
		charStore.UpdateSkill(key, skill)

		pc := shared.NewPageContext()
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Delete genre from multi-skill
	r.Post("/api/skill/{key}/genre/{index}/delete", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		// Remove the genre at index
		skill.Multi.Genres = append(skill.Multi.Genres[:index], skill.Multi.Genres[index+1:]...)
		charStore.UpdateSkill(key, skill)

		pc := shared.NewPageContext()
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Toggle grow flag for genre
	r.Post("/api/skill/{key}/genre/{index}/grow", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		skill.Multi.Genres[index].Grow = !skill.Multi.Genres[index].Grow
		charStore.UpdateSkill(key, skill)

		pc := shared.NewPageContext()
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Update genre label
	r.Post("/api/skill/{key}/genre/{index}/label", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		r.ParseForm()
		label := r.FormValue("label")
		skill.Multi.Genres[index].Label = label
		charStore.UpdateSkill(key, skill)

		return shared.Empty()
	}))

	// Adjust genre field (job, hobby, perm, temp)
	r.Post("/api/skill/{key}/genre/{index}/{field}/adjust", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		field := chi.URLParam(r, "field")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		deltaStr := r.URL.Query().Get("delta")
		delta := 0
		fmt.Sscanf(deltaStr, "%d", &delta)

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		genre := &skill.Multi.Genres[index]
		switch field {
		case "job":
			genre.Job += delta
			if genre.Job < 0 {
				genre.Job = 0
			}
		case "hobby":
			genre.Hobby += delta
			if genre.Hobby < 0 {
				genre.Hobby = 0
			}
		case "perm":
			genre.Perm += delta
		case "temp":
			genre.Temp += delta
		}

		charStore.UpdateSkill(key, skill)

		pc := shared.NewPageContext()
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Extra points adjustment
	r.Post("/api/status/{key}/adjust", HTML(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		// Handle extra-job and extra-hobby
		if key == "extra-job" || key == "extra-hobby" {
			deltaStr := r.URL.Query().Get("delta")
			delta := 0
			fmt.Sscanf(deltaStr, "%d", &delta)

			skills := charStore.GetSkills()
			if key == "extra-job" {
				skills.Extra.Job += delta
				if skills.Extra.Job < 0 {
					skills.Extra.Job = 0
				}
			} else {
				skills.Extra.Hobby += delta
				if skills.Extra.Hobby < 0 {
					skills.Extra.Hobby = 0
				}
			}
			charStore.SetSkillExtra(skills.Extra.Job, skills.Extra.Hobby)

			pc := shared.NewPageContext()
			status := charStore.GetStatus()
			skills = charStore.GetSkills()
			state := cthulhu6.BuildSheetState(pc, status, skills)
			return components.Cthulhu6SkillsPanelWithPoints(state)
		}

		// Handle parameters (HP, MP, SAN)
		if strings.HasPrefix(key, "param-") {
			paramKey := strings.TrimPrefix(key, "param-")
			deltaStr := r.URL.Query().Get("delta")
			delta := 0
			fmt.Sscanf(deltaStr, "%d", &delta)

			charStore.UpdateParameter(paramKey, delta)

			pc := shared.NewPageContext()
			status := charStore.GetStatus()
			skills := charStore.GetSkills()
			state := cthulhu6.BuildSheetState(pc, status, skills)
			return components.Cthulhu6StatusPanel(state, true)
		}

		// Original status variable handling
		key = strings.TrimPrefix(key, "status-")

		deltaStr := r.URL.Query().Get("delta")
		delta := 0
		fmt.Sscanf(deltaStr, "%d", &delta)

		pc := shared.NewPageContext()
		updated := charStore.UpdateVariableBase(key, delta)
		if updated == nil {
			return shared.Empty()
		}

		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)

		// For DEX/EDU changes, also update skills panel (回避 depends on DEX, 母国語 depends on EDU)
		// For INT/EDU changes, also update skill points display
		switch key {
		case "DEX", "EDU":
			return components.Cthulhu6StatusPanelWithSkills(state)
		case "INT":
			return components.Cthulhu6StatusPanelWithPoints(state)
		default:
			return components.Cthulhu6StatusPanel(state, true)
		}
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
