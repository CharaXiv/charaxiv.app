package cthulhu6

import (
	"encoding/json"
	"sync"

	"charaxiv/storage/coalesce"
	"charaxiv/systems/cthulhu6"
)

// Store wraps coalesce.Store with cthulhu6-specific typed access.
type Store struct {
	coalesce *coalesce.Store
	mu       sync.RWMutex

	// In-memory cache (loaded on first access per character)
	status map[string]*cthulhu6.Status
	skills map[string]*cthulhu6.Skills
	memos  map[string]map[string]string
}

// NewStore creates a new cthulhu6 store backed by coalesce storage.
func NewStore(c *coalesce.Store) *Store {
	return &Store{
		coalesce: c,
		status:   make(map[string]*cthulhu6.Status),
		skills:   make(map[string]*cthulhu6.Skills),
		memos:    make(map[string]map[string]string),
	}
}

// load ensures character data is loaded into memory
func (s *Store) load(charID string) error {
	if _, ok := s.status[charID]; ok {
		return nil // already loaded
	}

	data, err := s.coalesce.Read(charID)
	if err != nil {
		return err
	}

	// Start with defaults
	status := cthulhu6.NewStatus()
	skills := cthulhu6.NewSkills()
	memos := make(map[string]string)

	// Merge status variables (only override fields that exist in loaded data)
	if statusData, ok := data["status"].(map[string]any); ok {
		if varsData, ok := statusData["variables"].(map[string]any); ok {
			for key, varData := range varsData {
				if v, ok := status.Variables[key]; ok {
					if vm, ok := varData.(map[string]any); ok {
						if base, ok := vm["base"].(float64); ok {
							v.Base = int(base)
						}
						if perm, ok := vm["perm"].(float64); ok {
							v.Perm = int(perm)
						}
						if temp, ok := vm["temp"].(float64); ok {
							v.Temp = int(temp)
						}
						status.Variables[key] = v
					}
				}
			}
		}
		// Merge parameters
		if paramsData, ok := statusData["parameters"].(map[string]any); ok {
			for key, val := range paramsData {
				if v, ok := val.(float64); ok {
					intVal := int(v)
					status.Parameters[key] = &intVal
				}
			}
		}
		// Merge DB
		if db, ok := statusData["db"].(string); ok {
			status.DB = db
		}
	}
	s.status[charID] = status

	// Merge skills (for now, just unmarshal if present)
	if skillsData, ok := data["skills"].(map[string]any); ok {
		if b, err := json.Marshal(skillsData); err == nil {
			// Create a temp skills to unmarshal into, then merge
			var loadedSkills cthulhu6.Skills
			if json.Unmarshal(b, &loadedSkills) == nil {
				// Merge categories
				for cat, loadedCat := range loadedSkills.Categories {
					if defaultCat, ok := skills.Categories[cat]; ok {
						for skillKey, loadedSkill := range loadedCat.Skills {
							defaultCat.Skills[skillKey] = loadedSkill
						}
						skills.Categories[cat] = defaultCat
					}
				}
				skills.Extra = loadedSkills.Extra
				skills.Custom = loadedSkills.Custom
			}
		}
	}
	s.skills[charID] = skills

	// Parse memos
	if memosData, ok := data["memos"].(map[string]any); ok {
		for k, v := range memosData {
			if str, ok := v.(string); ok {
				memos[k] = str
			}
		}
	}
	s.memos[charID] = memos

	return nil
}

// GetStatus returns the status for a character
func (s *Store) GetStatus(charID string) *cthulhu6.Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return cthulhu6.NewStatus()
	}
	return s.status[charID]
}

// GetSkills returns the skills for a character
func (s *Store) GetSkills(charID string) *cthulhu6.Skills {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return cthulhu6.NewSkills()
	}
	return s.skills[charID]
}

// GetMemo returns a memo value
func (s *Store) GetMemo(charID, memoID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return ""
	}
	if m, ok := s.memos[charID]; ok {
		return m[memoID]
	}
	return ""
}

// SetMemo sets a memo value
func (s *Store) SetMemo(charID, memoID, value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return false
	}

	if s.memos[charID] == nil {
		s.memos[charID] = make(map[string]string)
	}
	if s.memos[charID][memoID] == value {
		return false
	}
	s.memos[charID][memoID] = value
	s.coalesce.Write(charID, "memos."+memoID, value)
	return true
}

// SetVariableBase sets a variable's base value
func (s *Store) SetVariableBase(charID, key string, value int) *cthulhu6.Variable {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return nil
	}

	status := s.status[charID]
	if v, ok := status.Variables[key]; ok {
		if value < v.Min {
			value = v.Min
		}
		if value > v.Max {
			value = v.Max
		}
		v.Base = value
		status.Variables[key] = v
		s.coalesce.Write(charID, "status.variables."+key+".base", value)
		return &v
	}
	return nil
}

// UpdateVariableBase updates a variable's base by delta
func (s *Store) UpdateVariableBase(charID, key string, delta int) *cthulhu6.Variable {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return nil
	}

	status := s.status[charID]
	if v, ok := status.Variables[key]; ok {
		newBase := v.Base + delta
		if newBase < v.Min {
			newBase = v.Min
		}
		if newBase > v.Max {
			newBase = v.Max
		}
		v.Base = newBase
		status.Variables[key] = v
		s.coalesce.Write(charID, "status.variables."+key+".base", newBase)
		return &v
	}
	return nil
}

// UpdateParameter updates a parameter by delta
func (s *Store) UpdateParameter(charID, key string, delta int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return 0
	}

	status := s.status[charID]
	current := status.EffectiveParameter(key)
	newVal := current + delta
	if newVal < 0 {
		newVal = 0
	}
	status.Parameters[key] = &newVal
	s.coalesce.Write(charID, "status.parameters."+key, newVal)
	return newVal
}

// SetDamageBonus sets the damage bonus
func (s *Store) SetDamageBonus(charID, db string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return
	}
	s.status[charID].DB = db
	s.coalesce.Write(charID, "status.db", db)
}

// GetSkill returns a skill by key
func (s *Store) GetSkill(charID, key string) (cthulhu6.Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.load(charID); err != nil {
		return cthulhu6.Skill{}, false
	}

	skills := s.skills[charID]
	for _, catData := range skills.Categories {
		if skill, ok := catData.Skills[key]; ok {
			return skill, true
		}
	}
	return cthulhu6.Skill{}, false
}

// UpdateSkill updates a skill
func (s *Store) UpdateSkill(charID, key string, skill cthulhu6.Skill) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return
	}

	skills := s.skills[charID]
	for cat, catData := range skills.Categories {
		if _, ok := catData.Skills[key]; ok {
			catData.Skills[key] = skill
			skills.Categories[cat] = catData
			// Write the entire skill object
			s.coalesce.Write(charID, "skills.categories."+string(cat)+".skills."+key, skill)
			return
		}
	}
}

// SetSkillExtra sets extra skill points
func (s *Store) SetSkillExtra(charID string, job, hobby int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.load(charID); err != nil {
		return
	}

	skills := s.skills[charID]
	skills.Extra.Job = job
	skills.Extra.Hobby = hobby
	s.coalesce.Write(charID, "skills.extra", skills.Extra)
}
