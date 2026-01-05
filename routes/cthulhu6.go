package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"charaxiv/models"
	"charaxiv/systems/cthulhu6"
	"charaxiv/templates/components"
	"charaxiv/templates/pages"
	"charaxiv/templates/shared"
)

// html wraps a templ.Component handler, setting the Content-Type header.
func html(c func(r *http.Request) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		c(r).Render(r.Context(), w)
	}
}

const basePath = "/cthulhu6"

// buildPageContext creates a PageContext with memos loaded from the store
func buildPageContext(store *models.Store) shared.PageContext {
	ctx := shared.NewPageContext()
	ctx.BasePath = basePath
	// Load all memos
	memoIDs := []string{"public-memo", "secret-memo", "scenario-public-memo", "scenario-secret-memo"}
	for _, id := range memoIDs {
		ctx.Memos[id] = store.GetMemo(id)
	}
	return ctx
}

// Cthulhu6 returns a chi.Router with all cthulhu6-specific routes.
func Cthulhu6(charStore *models.Store) chi.Router {
	r := chi.NewRouter()

	// Character sheet
	r.Get("/", html(func(r *http.Request) templ.Component {
		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return pages.Cthulhu6Sheet(state)
	}))

	// Preview mode toggle - returns targeted fragments with OOB swaps
	r.Post("/api/preview/on", html(func(r *http.Request) templ.Component {
		ctx := buildPageContext(charStore)
		ctx.Preview = true
		return pages.Cthulhu6PreviewModeFragments(ctx)
	}))

	r.Post("/api/preview/off", html(func(r *http.Request) templ.Component {
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
		pc := buildPageContext(charStore)
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
	r.Post("/api/skill/{key}/grow", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsSingle() {
			return shared.Empty()
		}

		// Toggle grow flag
		skill.Single.Grow = !skill.Single.Grow
		charStore.UpdateSkill(key, skill)

		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Skill field adjustment (job, hobby, perm, temp)
	r.Post("/api/skill/{key}/{field}/adjust", html(func(r *http.Request) templ.Component {
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

		return components.Cthulhu6SkillUpdateFragments(templSkill, field, remaining, basePath)
	}))

	// Add genre to multi-skill
	r.Post("/api/skill/{key}/genre/add", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		skill, ok := charStore.GetSkill(key)
		if !ok || !skill.IsMulti() {
			return shared.Empty()
		}

		// Add a new empty genre
		skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
		charStore.UpdateSkill(key, skill)

		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Delete genre from multi-skill
	r.Post("/api/skill/{key}/genre/{index}/delete", html(func(r *http.Request) templ.Component {
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

		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Toggle grow flag for genre
	r.Post("/api/skill/{key}/genre/{index}/grow", html(func(r *http.Request) templ.Component {
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

		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Update genre label
	r.Post("/api/skill/{key}/genre/{index}/label", html(func(r *http.Request) templ.Component {
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
	r.Post("/api/skill/{key}/genre/{index}/{field}/adjust", html(func(r *http.Request) templ.Component {
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

		pc := buildPageContext(charStore)
		status := charStore.GetStatus()
		skills := charStore.GetSkills()
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Extra points adjustment
	r.Post("/api/status/{key}/adjust", html(func(r *http.Request) templ.Component {
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

			pc := buildPageContext(charStore)
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

			pc := buildPageContext(charStore)
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

		pc := buildPageContext(charStore)
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

	return r
}
