package cthulhu6

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"charaxiv/storage/coalesce"
	"charaxiv/systems/cthulhu6"
)

// RouteTest defines a test case for a route
type RouteTest struct {
	Method       string
	Route        string
	Desc         string
	TestURL      string       // Actual URL to test (with params filled in)
	Form         url.Values   // Form data to send
	Query        string       // Query string (e.g., "delta=1")
	Setup        func(*Store) // Optional setup before test
	WantCode     int
	WantContains []string // Strings that should be in response body
}

// routeTests defines all routes with their test cases
var routeTests = []RouteTest{
	{
		Method:       "GET",
		Route:        "/cthulhu6/",
		Desc:         "Sheet page",
		TestURL:      "/cthulhu6/",
		WantCode:     http.StatusOK,
		WantContains: []string{"CharaXiv", "/cthulhu6/api/", "text/html"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/preview/on",
		Desc:         "Enable preview mode",
		TestURL:      "/cthulhu6/api/preview/on",
		WantCode:     http.StatusOK,
		WantContains: []string{"memo-group"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/preview/off",
		Desc:         "Disable preview mode",
		TestURL:      "/cthulhu6/api/preview/off",
		WantCode:     http.StatusOK,
		WantContains: []string{"memo-group"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/status/{key}/set",
		Desc:         "Set status variable directly",
		TestURL:      "/cthulhu6/api/status/status-STR/set",
		Form:         url.Values{"status_STR": {"15"}},
		WantCode:     http.StatusOK,
		WantContains: []string{"status-panel"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/status/{key}/adjust",
		Desc:         "Adjust status variable",
		TestURL:      "/cthulhu6/api/status/status-STR/adjust",
		Query:        "delta=1",
		WantCode:     http.StatusOK,
		WantContains: []string{"status-panel"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/status/{key}/adjust",
		Desc:         "Adjust parameter (HP)",
		TestURL:      "/cthulhu6/api/status/param-HP/adjust",
		Query:        "delta=1",
		WantCode:     http.StatusOK,
		WantContains: []string{"status-panel"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/status/{key}/adjust",
		Desc:         "Adjust extra job points",
		TestURL:      "/cthulhu6/api/status/extra-job/adjust",
		Query:        "delta=1",
		WantCode:     http.StatusOK,
		WantContains: []string{"skills-panel"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/memo/{id}/set",
		Desc:         "Set memo content",
		TestURL:      "/cthulhu6/api/memo/public-memo/set",
		Form:         url.Values{"public-memo": {"Test content"}},
		WantCode:     http.StatusNoContent,
		WantContains: nil,
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/skill/{key}/grow",
		Desc:         "Toggle skill grow flag",
		TestURL:      "/cthulhu6/api/skill/回避/grow",
		WantCode:     http.StatusOK,
		WantContains: []string{"skills-panel", "/cthulhu6/api/skill/"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/skill/{key}/{field}/adjust",
		Desc:         "Adjust skill field",
		TestURL:      "/cthulhu6/api/skill/回避/job/adjust",
		Query:        "delta=5",
		WantCode:     http.StatusOK,
		WantContains: []string{"hx-swap-oob", "/cthulhu6/api/"},
	},
	{
		Method:       "POST",
		Route:        "/cthulhu6/api/skill/{key}/genre/add",
		Desc:         "Add genre to multi-skill",
		TestURL:      "/cthulhu6/api/skill/芸術/genre/add",
		WantCode:     http.StatusOK,
		WantContains: []string{"skills-panel"},
	},
	{
		Method:  "POST",
		Route:   "/cthulhu6/api/skill/{key}/genre/{index}/delete",
		Desc:    "Delete genre from multi-skill",
		TestURL: "/cthulhu6/api/skill/芸術/genre/0/delete",
		Setup: func(s *Store) {
			// Add a genre first so we can delete it
			skill, _ := s.GetSkill("demo", "芸術")
			skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
			s.UpdateSkill("demo", "芸術", skill)
		},
		WantCode:     http.StatusOK,
		WantContains: []string{"skills-panel"},
	},
	{
		Method:  "POST",
		Route:   "/cthulhu6/api/skill/{key}/genre/{index}/grow",
		Desc:    "Toggle genre grow flag",
		TestURL: "/cthulhu6/api/skill/芸術/genre/0/grow",
		Setup: func(s *Store) {
			skill, _ := s.GetSkill("demo", "芸術")
			skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
			s.UpdateSkill("demo", "芸術", skill)
		},
		WantCode:     http.StatusOK,
		WantContains: []string{"skills-panel"},
	},
	{
		Method:  "POST",
		Route:   "/cthulhu6/api/skill/{key}/genre/{index}/label",
		Desc:    "Set genre label",
		TestURL: "/cthulhu6/api/skill/芸術/genre/0/label",
		Form:    url.Values{"label": {"絵画"}},
		Setup: func(s *Store) {
			skill, _ := s.GetSkill("demo", "芸術")
			skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
			s.UpdateSkill("demo", "芸術", skill)
		},
		WantCode:     http.StatusOK,
		WantContains: nil, // Returns empty
	},
	{
		Method:  "POST",
		Route:   "/cthulhu6/api/skill/{key}/genre/{index}/{field}/adjust",
		Desc:    "Adjust genre field",
		TestURL: "/cthulhu6/api/skill/芸術/genre/0/job/adjust",
		Query:   "delta=5",
		Setup: func(s *Store) {
			skill, _ := s.GetSkill("demo", "芸術")
			skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
			s.UpdateSkill("demo", "芸術", skill)
		},
		WantCode:     http.StatusOK,
		WantContains: []string{"skills-panel"},
	},
}

// TestRoutes runs all route tests from the table
func TestRoutes(t *testing.T) {
	for _, tt := range routeTests {
		t.Run(tt.Desc, func(t *testing.T) {
			r, store, cleanup := setupTestRouter(t)
			defer cleanup()

			// Run setup if provided
			if tt.Setup != nil {
				tt.Setup(store)
			}

			// Build URL with query string
			testURL := tt.TestURL
			if tt.Query != "" {
				testURL += "?" + tt.Query
			}

			// Build request
			var req *http.Request
			if tt.Form != nil {
				req = httptest.NewRequest(tt.Method, testURL, strings.NewReader(tt.Form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.Method, testURL, nil)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.WantCode {
				t.Errorf("Expected status %d, got %d", tt.WantCode, w.Code)
			}

			// Check response contains expected strings
			body := w.Body.String()
			contentType := w.Header().Get("Content-Type")
			fullResponse := body + contentType

			for _, want := range tt.WantContains {
				if !strings.Contains(fullResponse, want) {
					t.Errorf("Expected response to contain %q", want)
				}
			}
		})
	}
}

// TestAllRoutesHaveTests verifies every registered route has a test case
func TestAllRoutesHaveTests(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// Build set of routes that have tests
	testedRoutes := make(map[string]bool)
	for _, tt := range routeTests {
		key := tt.Method + " " + tt.Route
		testedRoutes[key] = true
	}

	// Walk registered routes and check each has a test
	var untestedRoutes []string
	var registeredRoutes []string

	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		key := method + " " + route
		registeredRoutes = append(registeredRoutes, key)
		if !testedRoutes[key] {
			untestedRoutes = append(untestedRoutes, key)
		}
		return nil
	})

	// Report untested routes
	for _, route := range untestedRoutes {
		t.Errorf("Route has no test: %s", route)
	}

	// Check for tests that don't match any registered route
	registeredSet := make(map[string]bool)
	for _, route := range registeredRoutes {
		registeredSet[route] = true
	}

	for _, tt := range routeTests {
		key := tt.Method + " " + tt.Route
		if !registeredSet[key] {
			t.Errorf("Test exists for unregistered route: %s (%s)", key, tt.Desc)
		}
	}
}

// setupTestRouter creates a test router with a fresh store.
func setupTestRouter(t *testing.T) (chi.Router, *Store, func()) {
	t.Helper()

	// Create temp directory for test data
	tmpDir, err := os.MkdirTemp("", "cthulhu6-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cs, err := coalesce.New(coalesce.Config{
		DBPath:  filepath.Join(tmpDir, "buffer.db"),
		DataDir: filepath.Join(tmpDir, "characters"),
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create coalesce store: %v", err)
	}

	store := NewStore(cs)
	r := chi.NewRouter()
	r.Mount("/cthulhu6", Routes(store))

	cleanup := func() {
		cs.Close()
		os.RemoveAll(tmpDir)
	}

	return r, store, cleanup
}

// Additional focused tests for specific behaviors

// TestStatusSetNoChange tests that unchanged values return 204
func TestStatusSetNoChange(t *testing.T) {
	r, store, cleanup := setupTestRouter(t)
	defer cleanup()

	initialSTR := store.GetStatus("demo").Variables["STR"].Base

	form := url.Values{}
	form.Set("status_STR", fmt.Sprintf("%d", initialSTR))
	req := httptest.NewRequest("POST", "/cthulhu6/api/status/status-STR/set", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for unchanged value, got %d", w.Code)
	}
}

// TestSkillAdjustChangesValue tests that skill adjustment actually changes the value
func TestSkillAdjustChangesValue(t *testing.T) {
	r, store, cleanup := setupTestRouter(t)
	defer cleanup()

	skill, _ := store.GetSkill("demo", "回避")
	initialJob := skill.Single.Job

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/job/adjust?delta=5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	skill, _ = store.GetSkill("demo", "回避")
	if skill.Single.Job != initialJob+5 {
		t.Errorf("Expected job to be %d, got %d", initialJob+5, skill.Single.Job)
	}
}

// TestSkillAdjustNonNegative tests that job/hobby can't go below 0
func TestSkillAdjustNonNegative(t *testing.T) {
	r, store, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/job/adjust?delta=-100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	skill, _ := store.GetSkill("demo", "回避")
	if skill.Single.Job < 0 {
		t.Error("Expected job to not go below 0")
	}
}

// TestMemoSetChangesValue tests that memo is actually saved
func TestMemoSetChangesValue(t *testing.T) {
	r, store, cleanup := setupTestRouter(t)
	defer cleanup()

	form := url.Values{}
	form.Set("public-memo", "Test memo content")
	req := httptest.NewRequest("POST", "/cthulhu6/api/memo/public-memo/set", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if store.GetMemo("demo", "public-memo") != "Test memo content" {
		t.Error("Expected memo to be saved")
	}
}

// TestGenreAddChangesCount tests that adding a genre increases the count
func TestGenreAddChangesCount(t *testing.T) {
	r, store, cleanup := setupTestRouter(t)
	defer cleanup()

	skill, _ := store.GetSkill("demo", "芸術")
	initialCount := len(skill.Multi.Genres)

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	skill, _ = store.GetSkill("demo", "芸術")
	if len(skill.Multi.Genres) != initialCount+1 {
		t.Errorf("Expected %d genres, got %d", initialCount+1, len(skill.Multi.Genres))
	}
}

// TestInvalidSkillReturnsEmpty tests handling of invalid skill keys
func TestInvalidSkillReturnsEmpty(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/nonexistent/grow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	// Empty component returns minimal HTML
	if len(w.Body.String()) > 100 {
		t.Error("Expected minimal/empty response for invalid skill")
	}
}

// TestOOBFragmentsHaveCorrectBasePath tests that OOB fragments have correct API paths
func TestOOBFragmentsHaveCorrectBasePath(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/grow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()

	if strings.Contains(body, `hx-post="/api/`) {
		t.Error("OOB fragment contains /api/ without /cthulhu6 prefix")
	}

	if !strings.Contains(body, `/cthulhu6/api/skill/`) {
		t.Error("OOB fragment missing /cthulhu6/api/skill/ routes")
	}
}

// TestDEXChangeIncludesSkillsPanel tests that DEX changes return skills panel OOB
func TestDEXChangeIncludesSkillsPanel(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/cthulhu6/api/status/status-DEX/adjust?delta=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "skills-panel") {
		t.Error("DEX change should include skills panel in OOB response")
	}
}

// TestINTChangeIncludesPointsDisplay tests that INT changes return points display OOB
func TestINTChangeIncludesPointsDisplay(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/cthulhu6/api/status/status-INT/adjust?delta=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "points-display") {
		t.Error("INT change should include points display in OOB response")
	}
}
