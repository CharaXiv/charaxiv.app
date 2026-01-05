package cthulhu6

import (
	"charaxiv/storage/coalesce"
	"charaxiv/systems/cthulhu6"
)

// Store wraps coalesce.Store with cthulhu6-specific typed access.
type Store struct {
	coalesce *coalesce.Store
}

// NewStore creates a new cthulhu6 store backed by coalesce storage.
func NewStore(c *coalesce.Store) *Store {
	return &Store{coalesce: c}
}

// load reads character data from coalesce (which flushes pending writes)
func (s *Store) load(charID string) (*cthulhu6.Status, *cthulhu6.Skills, map[string]string) {
	data, err := s.coalesce.Read(charID)
	if err != nil {
		return cthulhu6.NewStatus(), cthulhu6.NewSkills(), make(map[string]string)
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

	// Merge skills
	if skillsData, ok := data["skills"].(map[string]any); ok {
		if catsData, ok := skillsData["categories"].(map[string]any); ok {
			for catName, catData := range catsData {
				cat := cthulhu6.SkillCategory(catName)
				if defaultCat, ok := skills.Categories[cat]; ok {
					if catMap, ok := catData.(map[string]any); ok {
						if skillsMap, ok := catMap["skills"].(map[string]any); ok {
							for skillKey, skillData := range skillsMap {
								if skillMap, ok := skillData.(map[string]any); ok {
									skill := parseSkill(skillMap)
									defaultCat.Skills[skillKey] = skill
								}
							}
						}
					}
					skills.Categories[cat] = defaultCat
				}
			}
		}
		if extraData, ok := skillsData["extra"].(map[string]any); ok {
			if job, ok := extraData["job"].(float64); ok {
				skills.Extra.Job = int(job)
			}
			if hobby, ok := extraData["hobby"].(float64); ok {
				skills.Extra.Hobby = int(hobby)
			}
		}
	}

	// Parse memos
	if memosData, ok := data["memos"].(map[string]any); ok {
		for k, v := range memosData {
			if str, ok := v.(string); ok {
				memos[k] = str
			}
		}
	}

	return status, skills, memos
}

// parseSkill parses a skill from map[string]any
func parseSkill(data map[string]any) cthulhu6.Skill {
	skill := cthulhu6.Skill{}
	if order, ok := data["order"].(float64); ok {
		skill.Order = int(order)
	}
	if singleData, ok := data["single"].(map[string]any); ok {
		skill.Single = &cthulhu6.SingleSkill{}
		if v, ok := singleData["job"].(float64); ok {
			skill.Single.Job = int(v)
		}
		if v, ok := singleData["hobby"].(float64); ok {
			skill.Single.Hobby = int(v)
		}
		if v, ok := singleData["perm"].(float64); ok {
			skill.Single.Perm = int(v)
		}
		if v, ok := singleData["temp"].(float64); ok {
			skill.Single.Temp = int(v)
		}
		if v, ok := singleData["grow"].(bool); ok {
			skill.Single.Grow = v
		}
	}
	if multiData, ok := data["multi"].(map[string]any); ok {
		skill.Multi = &cthulhu6.MultiSkill{}
		if genresData, ok := multiData["genres"].([]any); ok {
			for _, gd := range genresData {
				if gm, ok := gd.(map[string]any); ok {
					genre := cthulhu6.SkillGenre{}
					if v, ok := gm["label"].(string); ok {
						genre.Label = v
					}
					if v, ok := gm["job"].(float64); ok {
						genre.Job = int(v)
					}
					if v, ok := gm["hobby"].(float64); ok {
						genre.Hobby = int(v)
					}
					if v, ok := gm["perm"].(float64); ok {
						genre.Perm = int(v)
					}
					if v, ok := gm["temp"].(float64); ok {
						genre.Temp = int(v)
					}
					if v, ok := gm["grow"].(bool); ok {
						genre.Grow = v
					}
					skill.Multi.Genres = append(skill.Multi.Genres, genre)
				}
			}
		}
	}
	return skill
}

// GetStatus returns the status for a character
func (s *Store) GetStatus(charID string) *cthulhu6.Status {
	status, _, _ := s.load(charID)
	return status
}

// GetSkills returns the skills for a character
func (s *Store) GetSkills(charID string) *cthulhu6.Skills {
	_, skills, _ := s.load(charID)
	return skills
}

// GetMemo returns a memo value
func (s *Store) GetMemo(charID, memoID string) string {
	_, _, memos := s.load(charID)
	return memos[memoID]
}

// SetMemo sets a memo value
func (s *Store) SetMemo(charID, memoID, value string) bool {
	_, _, memos := s.load(charID)
	if memos[memoID] == value {
		return false
	}
	s.coalesce.Write(charID, "memos."+memoID, value)
	return true
}

// SetVariableBase sets a variable's base value
func (s *Store) SetVariableBase(charID, key string, value int) *cthulhu6.Variable {
	status, _, _ := s.load(charID)
	if v, ok := status.Variables[key]; ok {
		if value < v.Min {
			value = v.Min
		}
		if value > v.Max {
			value = v.Max
		}
		v.Base = value
		s.coalesce.Write(charID, "status.variables."+key+".base", value)
		return &v
	}
	return nil
}

// UpdateVariableBase updates a variable's base by delta
func (s *Store) UpdateVariableBase(charID, key string, delta int) *cthulhu6.Variable {
	status, _, _ := s.load(charID)
	if v, ok := status.Variables[key]; ok {
		newBase := v.Base + delta
		if newBase < v.Min {
			newBase = v.Min
		}
		if newBase > v.Max {
			newBase = v.Max
		}
		v.Base = newBase
		s.coalesce.Write(charID, "status.variables."+key+".base", newBase)
		return &v
	}
	return nil
}

// UpdateParameter updates a parameter by delta
func (s *Store) UpdateParameter(charID, key string, delta int) int {
	status, _, _ := s.load(charID)
	current := status.EffectiveParameter(key)
	newVal := current + delta
	if newVal < 0 {
		newVal = 0
	}
	s.coalesce.Write(charID, "status.parameters."+key, newVal)
	return newVal
}

// SetDamageBonus sets the damage bonus
func (s *Store) SetDamageBonus(charID, db string) {
	s.coalesce.Write(charID, "status.db", db)
}

// GetSkill returns a skill by key
func (s *Store) GetSkill(charID, key string) (cthulhu6.Skill, bool) {
	_, skills, _ := s.load(charID)
	for _, catData := range skills.Categories {
		if skill, ok := catData.Skills[key]; ok {
			return skill, true
		}
	}
	return cthulhu6.Skill{}, false
}

// UpdateSkill updates a skill
func (s *Store) UpdateSkill(charID, key string, skill cthulhu6.Skill) {
	_, skills, _ := s.load(charID)
	for cat, catData := range skills.Categories {
		if _, ok := catData.Skills[key]; ok {
			s.coalesce.Write(charID, "skills.categories."+string(cat)+".skills."+key, skill)
			return
		}
	}
}

// SetSkillExtra sets extra skill points
func (s *Store) SetSkillExtra(charID string, job, hobby int) {
	s.coalesce.Write(charID, "skills.extra", cthulhu6.SkillExtra{Job: job, Hobby: hobby})
}
