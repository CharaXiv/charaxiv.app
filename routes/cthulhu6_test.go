package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"charaxiv/models"
)

// setupTestRouter creates a test router with a fresh store
func setupTestRouter() (chi.Router, *models.Store) {
	store := models.NewStore()
	r := chi.NewRouter()
	r.Mount("/cthulhu6", Cthulhu6(store))
	return r, store
}

// TestSheetPage tests the main sheet page renders
func TestSheetPage(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/cthulhu6/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected text/html content type, got %s", w.Header().Get("Content-Type"))
	}

	// Check that the page contains expected elements
	body := w.Body.String()
	if !strings.Contains(body, "CharaXiv") {
		t.Error("Expected page to contain CharaXiv")
	}
	if !strings.Contains(body, "/cthulhu6/api/") {
		t.Error("Expected page to contain /cthulhu6/api/ routes")
	}
}

// TestPreviewModeOn tests enabling preview mode
func TestPreviewModeOn(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/cthulhu6/api/preview/on", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	// Preview mode should return memo fragments in read-only mode
	if !strings.Contains(body, "memo-group") {
		t.Error("Expected response to contain memo-group")
	}
}

// TestPreviewModeOff tests disabling preview mode
func TestPreviewModeOff(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/cthulhu6/api/preview/off", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestStatusAdjust tests adjusting status variables
func TestStatusAdjust(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		delta    string
		wantCode int
	}{
		{"STR +1", "status-STR", "1", http.StatusOK},
		{"STR -1", "status-STR", "-1", http.StatusOK},
		{"STR +5", "status-STR", "5", http.StatusOK},
		{"CON +1", "status-CON", "1", http.StatusOK},
		{"DEX +1", "status-DEX", "1", http.StatusOK},
		{"INT +1", "status-INT", "1", http.StatusOK},
		{"EDU +1", "status-EDU", "1", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := setupTestRouter()

			req := httptest.NewRequest("POST", "/cthulhu6/api/status/"+tt.key+"/adjust?delta="+tt.delta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

// TestStatusSet tests setting status variables directly
func TestStatusSet(t *testing.T) {
	r, store := setupTestRouter()

	// Get initial STR value
	initialSTR := store.GetStatus().Variables["STR"].Base

	// Set STR to a new value
	form := url.Values{}
	form.Set("status_STR", "15")
	req := httptest.NewRequest("POST", "/cthulhu6/api/status/status-STR/set", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify the value changed
	newSTR := store.GetStatus().Variables["STR"].Base
	if newSTR == initialSTR {
		t.Error("Expected STR to change")
	}
	if newSTR != 15 {
		t.Errorf("Expected STR to be 15, got %d", newSTR)
	}
}

// TestStatusSetNoChange tests that unchanged values return 204
func TestStatusSetNoChange(t *testing.T) {
	r, store := setupTestRouter()

	// Get initial STR value
	initialSTR := store.GetStatus().Variables["STR"].Base

	// Set STR to the same value
	form := url.Values{}
	form.Set("status_STR", string(rune('0'+initialSTR))) // Convert to string
	req := httptest.NewRequest("POST", "/cthulhu6/api/status/status-STR/set", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for unchanged value, got %d", w.Code)
	}
}

// TestParameterAdjust tests adjusting parameters (HP, MP, SAN)
func TestParameterAdjust(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		delta string
	}{
		{"HP +1", "param-HP", "1"},
		{"HP -1", "param-HP", "-1"},
		{"MP +1", "param-MP", "1"},
		{"SAN +1", "param-SAN", "1"},
		{"SAN -5", "param-SAN", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := setupTestRouter()

			req := httptest.NewRequest("POST", "/cthulhu6/api/status/"+tt.key+"/adjust?delta="+tt.delta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

// TestExtraPointsAdjust tests adjusting extra skill points
func TestExtraPointsAdjust(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		delta string
	}{
		{"Extra Job +1", "extra-job", "1"},
		{"Extra Job +5", "extra-job", "5"},
		{"Extra Hobby +1", "extra-hobby", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := setupTestRouter()

			req := httptest.NewRequest("POST", "/cthulhu6/api/status/"+tt.key+"/adjust?delta="+tt.delta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

// TestMemoSet tests setting memo content
func TestMemoSet(t *testing.T) {
	r, store := setupTestRouter()

	form := url.Values{}
	form.Set("public-memo", "Test memo content")
	req := httptest.NewRequest("POST", "/cthulhu6/api/memo/public-memo/set", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify the memo was saved
	if store.GetMemo("public-memo") != "Test memo content" {
		t.Error("Expected memo to be saved")
	}
}

// TestMemoSetAllTypes tests all memo types
func TestMemoSetAllTypes(t *testing.T) {
	memoIDs := []string{"public-memo", "secret-memo", "scenario-public-memo", "scenario-secret-memo"}

	for _, id := range memoIDs {
		t.Run(id, func(t *testing.T) {
			r, store := setupTestRouter()

			form := url.Values{}
			form.Set(id, "Content for "+id)
			req := httptest.NewRequest("POST", "/cthulhu6/api/memo/"+id+"/set", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusNoContent {
				t.Errorf("Expected status 204, got %d", w.Code)
			}

			if store.GetMemo(id) != "Content for "+id {
				t.Error("Expected memo to be saved")
			}
		})
	}
}

// TestSkillGrow tests toggling skill grow flag
func TestSkillGrow(t *testing.T) {
	r, store := setupTestRouter()

	// Get initial grow state for 回避
	skill, _ := store.GetSkill("回避")
	initialGrow := skill.Single.Grow

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/grow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify grow was toggled
	skill, _ = store.GetSkill("回避")
	if skill.Single.Grow == initialGrow {
		t.Error("Expected grow flag to be toggled")
	}
}

// TestSkillAdjust tests adjusting skill fields
func TestSkillAdjust(t *testing.T) {
	tests := []struct {
		name  string
		skill string
		field string
		delta string
	}{
		{"回避 job +1", "回避", "job", "1"},
		{"回避 hobby +5", "回避", "hobby", "5"},
		{"回避 perm +1", "回避", "perm", "1"},
		{"回避 temp +1", "回避", "temp", "1"},
		{"キック job +1", "キック", "job", "1"},
		{"目星 hobby +10", "目星", "hobby", "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := setupTestRouter()

			req := httptest.NewRequest("POST", "/cthulhu6/api/skill/"+url.PathEscape(tt.skill)+"/"+tt.field+"/adjust?delta="+tt.delta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Verify response contains OOB fragments
			body := w.Body.String()
			if !strings.Contains(body, "hx-swap-oob") {
				t.Error("Expected response to contain OOB swap fragments")
			}
		})
	}
}

// TestSkillAdjustChangesValue tests that skill adjustment actually changes the value
func TestSkillAdjustChangesValue(t *testing.T) {
	r, store := setupTestRouter()

	// Get initial value
	skill, _ := store.GetSkill("回避")
	initialJob := skill.Single.Job

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/job/adjust?delta=5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify value changed
	skill, _ = store.GetSkill("回避")
	if skill.Single.Job != initialJob+5 {
		t.Errorf("Expected job to be %d, got %d", initialJob+5, skill.Single.Job)
	}
}

// TestSkillAdjustNonNegative tests that job/hobby can't go below 0
func TestSkillAdjustNonNegative(t *testing.T) {
	r, store := setupTestRouter()

	// Try to decrease job below 0
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/job/adjust?delta=-100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify value is clamped to 0
	skill, _ := store.GetSkill("回避")
	if skill.Single.Job < 0 {
		t.Error("Expected job to not go below 0")
	}
}

// TestGenreAdd tests adding a genre to a multi-skill
func TestGenreAdd(t *testing.T) {
	r, store := setupTestRouter()

	// Get initial genre count for 芸術
	skill, _ := store.GetSkill("芸術")
	initialCount := len(skill.Multi.Genres)

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify genre was added
	skill, _ = store.GetSkill("芸術")
	if len(skill.Multi.Genres) != initialCount+1 {
		t.Errorf("Expected %d genres, got %d", initialCount+1, len(skill.Multi.Genres))
	}
}

// TestGenreDelete tests deleting a genre from a multi-skill
func TestGenreDelete(t *testing.T) {
	r, store := setupTestRouter()

	// First add a genre
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	skill, _ := store.GetSkill("芸術")
	countAfterAdd := len(skill.Multi.Genres)

	// Now delete it
	req = httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/0/delete", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify genre was deleted
	skill, _ = store.GetSkill("芸術")
	if len(skill.Multi.Genres) != countAfterAdd-1 {
		t.Errorf("Expected %d genres after delete, got %d", countAfterAdd-1, len(skill.Multi.Genres))
	}
}

// TestGenreGrow tests toggling genre grow flag
func TestGenreGrow(t *testing.T) {
	r, store := setupTestRouter()

	// First add a genre to 芸術
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Get initial grow state
	skill, _ := store.GetSkill("芸術")
	initialGrow := skill.Multi.Genres[0].Grow

	// Toggle grow
	req = httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/0/grow", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify grow was toggled
	skill, _ = store.GetSkill("芸術")
	if skill.Multi.Genres[0].Grow == initialGrow {
		t.Error("Expected grow flag to be toggled")
	}
}

// TestGenreLabel tests setting genre label
func TestGenreLabel(t *testing.T) {
	r, store := setupTestRouter()

	// First add a genre to 芸術
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Set label
	form := url.Values{}
	form.Set("label", "絵画")
	req = httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/0/label", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify label was set
	skill, _ := store.GetSkill("芸術")
	if skill.Multi.Genres[0].Label != "絵画" {
		t.Errorf("Expected label to be '絵画', got '%s'", skill.Multi.Genres[0].Label)
	}
}

// TestGenreAdjust tests adjusting genre fields
func TestGenreAdjust(t *testing.T) {
	r, store := setupTestRouter()

	// First add a genre to 芸術
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	tests := []struct {
		field string
		delta string
	}{
		{"job", "5"},
		{"hobby", "10"},
		{"perm", "3"},
		{"temp", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/0/"+tt.field+"/adjust?delta="+tt.delta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}

	// Verify values were set
	skill, _ := store.GetSkill("芸術")
	genre := skill.Multi.Genres[0]
	if genre.Job != 5 {
		t.Errorf("Expected job to be 5, got %d", genre.Job)
	}
	if genre.Hobby != 10 {
		t.Errorf("Expected hobby to be 10, got %d", genre.Hobby)
	}
	if genre.Perm != 3 {
		t.Errorf("Expected perm to be 3, got %d", genre.Perm)
	}
	if genre.Temp != 2 {
		t.Errorf("Expected temp to be 2, got %d", genre.Temp)
	}
}

// TestInvalidSkill tests handling of invalid skill keys
func TestInvalidSkill(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/nonexistent/grow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return OK with empty body (Empty component)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestInvalidGenreIndex tests handling of invalid genre index
func TestInvalidGenreIndex(t *testing.T) {
	r, _ := setupTestRouter()

	// Try to delete non-existent genre
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/芸術/genre/999/delete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return OK with empty body
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestOOBFragmentsHaveCorrectBasePath tests that OOB fragments have correct API paths
func TestOOBFragmentsHaveCorrectBasePath(t *testing.T) {
	r, _ := setupTestRouter()

	// Test skill grow toggle response
	req := httptest.NewRequest("POST", "/cthulhu6/api/skill/回避/grow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// All hx-post attributes should have /cthulhu6 prefix
	if strings.Contains(body, `hx-post="/api/`) {
		t.Error("OOB fragment contains /api/ without /cthulhu6 prefix")
	}

	// Should contain correctly prefixed routes
	if !strings.Contains(body, `/cthulhu6/api/skill/`) {
		t.Error("OOB fragment missing /cthulhu6/api/skill/ routes")
	}
}

// TestStatusAdjustOOBFragments tests that status adjust returns correct OOB fragments
func TestStatusAdjustOOBFragments(t *testing.T) {
	r, _ := setupTestRouter()

	// DEX changes should include skills panel (回避 depends on DEX)
	req := httptest.NewRequest("POST", "/cthulhu6/api/status/status-DEX/adjust?delta=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "skills-panel") {
		t.Error("DEX change should include skills panel in OOB response")
	}

	// INT changes should include points display
	req = httptest.NewRequest("POST", "/cthulhu6/api/status/status-INT/adjust?delta=1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body = w.Body.String()
	if !strings.Contains(body, "points-display") {
		t.Error("INT change should include points display in OOB response")
	}
}

// TestAllRoutesRegistered verifies all expected routes are registered
func TestAllRoutesRegistered(t *testing.T) {
	r, _ := setupTestRouter()

	expectedRoutes := map[string]bool{
		"GET /cthulhu6/":                                              false,
		"POST /cthulhu6/api/preview/on":                               false,
		"POST /cthulhu6/api/preview/off":                              false,
		"POST /cthulhu6/api/status/{key}/set":                         false,
		"POST /cthulhu6/api/status/{key}/adjust":                      false,
		"POST /cthulhu6/api/memo/{id}/set":                            false,
		"POST /cthulhu6/api/skill/{key}/grow":                         false,
		"POST /cthulhu6/api/skill/{key}/{field}/adjust":               false,
		"POST /cthulhu6/api/skill/{key}/genre/add":                    false,
		"POST /cthulhu6/api/skill/{key}/genre/{index}/delete":         false,
		"POST /cthulhu6/api/skill/{key}/genre/{index}/grow":           false,
		"POST /cthulhu6/api/skill/{key}/genre/{index}/label":          false,
		"POST /cthulhu6/api/skill/{key}/genre/{index}/{field}/adjust": false,
	}

	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		key := method + " " + route
		if _, ok := expectedRoutes[key]; ok {
			expectedRoutes[key] = true
		}
		return nil
	})

	for route, found := range expectedRoutes {
		if !found {
			t.Errorf("Expected route not registered: %s", route)
		}
	}
}
