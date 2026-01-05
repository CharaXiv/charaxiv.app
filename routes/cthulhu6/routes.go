package cthulhu6

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

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

// buildPageContext creates a PageContext with memos loaded from the store
func buildPageContext(store *Store, charID, basePath string) shared.PageContext {
	ctx := shared.NewPageContext()
	ctx.BasePath = basePath
	// Load all memos
	memoIDs := []string{"public-memo", "secret-memo", "scenario-public-memo", "scenario-secret-memo"}
	for _, id := range memoIDs {
		ctx.Memos[id] = store.GetMemo(charID, id)
	}
	return ctx
}

// Routes returns a chi.Router with all cthulhu6-specific routes.
func Routes(store *Store) chi.Router {
	r := chi.NewRouter()

	// For now, use a fixed character ID until we have proper routing
	const charID = "demo"
	const basePath = "/cthulhu6"

	// Character sheet
	r.Get("/", html(func(r *http.Request) templ.Component {
		pc := buildPageContext(store, charID, basePath)
		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return pages.Cthulhu6Sheet(state)
	}))

	// Preview mode toggle - returns targeted fragments with OOB swaps
	r.Post("/api/preview/on", html(func(r *http.Request) templ.Component {
		ctx := buildPageContext(store, charID, basePath)
		ctx.Preview = true
		return pages.Cthulhu6PreviewModeFragments(ctx)
	}))

	r.Post("/api/preview/off", html(func(r *http.Request) templ.Component {
		ctx := buildPageContext(store, charID, basePath)
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
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var value int
		_, err := fmt.Sscanf(valueStr, "%d", &value)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Check if value actually changed
		status := store.GetStatus(charID)
		if v, ok := status.Variables[key]; ok {
			clampedValue := value
			if clampedValue < v.Min {
				clampedValue = v.Min
			}
			if clampedValue > v.Max {
				clampedValue = v.Max
			}
			if v.Base == clampedValue {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		updated := store.SetVariableBase(charID, key, value)
		if updated == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		pc := buildPageContext(store, charID, basePath)
		status = store.GetStatus(charID)
		skills := store.GetSkills(charID)
		state := cthulhu6.BuildSheetState(pc, status, skills)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

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
		r.ParseForm()
		value := r.FormValue(id)

		if !store.SetMemo(charID, id, value) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Skill grow toggle
	r.Post("/api/skill/{key}/grow", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		skill, ok := store.GetSkill(charID, key)
		if !ok || !skill.IsSingle() {
			return shared.Empty()
		}

		skill.Single.Grow = !skill.Single.Grow
		store.UpdateSkill(charID, key, skill)

		pc := buildPageContext(store, charID, basePath)
		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
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

		skill, ok := store.GetSkill(charID, key)
		if !ok || !skill.IsSingle() {
			return shared.Empty()
		}

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

		store.UpdateSkill(charID, key, skill)

		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
		updatedSkill, _ := store.GetSkill(charID, key)

		templSkill := cthulhu6.BuildSkill(status, key, updatedSkill)
		remaining := cthulhu6.BuildRemainingPoints(status, skills)

		return components.Cthulhu6SkillUpdateFragments(templSkill, field, remaining, basePath)
	}))

	// Add genre to multi-skill
	r.Post("/api/skill/{key}/genre/add", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")

		skill, ok := store.GetSkill(charID, key)
		if !ok || !skill.IsMulti() {
			return shared.Empty()
		}

		skill.Multi.Genres = append(skill.Multi.Genres, cthulhu6.SkillGenre{})
		store.UpdateSkill(charID, key, skill)

		pc := buildPageContext(store, charID, basePath)
		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Delete genre from multi-skill
	r.Post("/api/skill/{key}/genre/{index}/delete", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		skill, ok := store.GetSkill(charID, key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		skill.Multi.Genres = append(skill.Multi.Genres[:index], skill.Multi.Genres[index+1:]...)
		store.UpdateSkill(charID, key, skill)

		pc := buildPageContext(store, charID, basePath)
		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Toggle grow flag for genre
	r.Post("/api/skill/{key}/genre/{index}/grow", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		skill, ok := store.GetSkill(charID, key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		skill.Multi.Genres[index].Grow = !skill.Multi.Genres[index].Grow
		store.UpdateSkill(charID, key, skill)

		pc := buildPageContext(store, charID, basePath)
		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
		state := cthulhu6.BuildSheetState(pc, status, skills)
		return components.Cthulhu6SkillsPanel(state, true)
	}))

	// Update genre label
	r.Post("/api/skill/{key}/genre/{index}/label", html(func(r *http.Request) templ.Component {
		key := chi.URLParam(r, "key")
		indexStr := chi.URLParam(r, "index")
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		skill, ok := store.GetSkill(charID, key)
		if !ok || !skill.IsMulti() || index < 0 || index >= len(skill.Multi.Genres) {
			return shared.Empty()
		}

		r.ParseForm()
		label := r.FormValue("label")
		skill.Multi.Genres[index].Label = label
		store.UpdateSkill(charID, key, skill)

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

		skill, ok := store.GetSkill(charID, key)
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

		store.UpdateSkill(charID, key, skill)

		pc := buildPageContext(store, charID, basePath)
		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
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

			skills := store.GetSkills(charID)
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
			store.SetSkillExtra(charID, skills.Extra.Job, skills.Extra.Hobby)

			pc := buildPageContext(store, charID, basePath)
			status := store.GetStatus(charID)
			skills = store.GetSkills(charID)
			state := cthulhu6.BuildSheetState(pc, status, skills)
			return components.Cthulhu6SkillsPanelWithPoints(state)
		}

		// Handle parameters (HP, MP, SAN)
		if strings.HasPrefix(key, "param-") {
			paramKey := strings.TrimPrefix(key, "param-")
			deltaStr := r.URL.Query().Get("delta")
			delta := 0
			fmt.Sscanf(deltaStr, "%d", &delta)

			store.UpdateParameter(charID, paramKey, delta)

			pc := buildPageContext(store, charID, basePath)
			status := store.GetStatus(charID)
			skills := store.GetSkills(charID)
			state := cthulhu6.BuildSheetState(pc, status, skills)
			return components.Cthulhu6StatusPanel(state, true)
		}

		// Original status variable handling
		key = strings.TrimPrefix(key, "status-")

		deltaStr := r.URL.Query().Get("delta")
		delta := 0
		fmt.Sscanf(deltaStr, "%d", &delta)

		pc := buildPageContext(store, charID, basePath)
		updated := store.UpdateVariableBase(charID, key, delta)
		if updated == nil {
			return shared.Empty()
		}

		status := store.GetStatus(charID)
		skills := store.GetSkills(charID)
		state := cthulhu6.BuildSheetState(pc, status, skills)

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
