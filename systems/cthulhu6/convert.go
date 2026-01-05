package cthulhu6

import (
	"sort"

	"charaxiv/templates/shared"
)

// BuildSheetState creates a SheetState from CoC6 data and page context
func BuildSheetState(pc shared.PageContext, status *Status, skills *Skills) shared.SheetState {
	vars, computed, params, db, skillCategories, skillExtra, skillPoints := convertToTemplates(status, skills)
	return shared.SheetState{
		PC: pc,
		Status: shared.StatusState{
			Variables:   vars,
			Computed:    computed,
			Parameters:  params,
			DamageBonus: db,
		},
		Skills: shared.SkillsState{
			Categories: skillCategories,
			Extra:      skillExtra,
			Remaining:  skillPoints,
		},
	}
}

// convertToTemplates converts CoC6 types to template types
func convertToTemplates(status *Status, skills *Skills) ([]shared.StatusVariable, []shared.ComputedValue, []shared.StatusParameter, string, []shared.SkillCategory, shared.SkillExtra, shared.SkillPoints) {
	// Variables in display order
	varOrder := []string{"STR", "CON", "POW", "DEX", "APP", "SIZ", "INT", "EDU"}
	variables := make([]shared.StatusVariable, 0, len(varOrder))
	for _, key := range varOrder {
		v := status.Variables[key]
		variables = append(variables, shared.StatusVariable{
			Key:  key,
			Base: v.Base,
			Perm: v.Perm,
			Temp: v.Temp,
			Min:  v.Min,
			Max:  v.Max,
		})
	}

	// Computed values in display order
	computedOrder := []string{"初期SAN", "アイデア", "幸運", "知識", "職業P", "興味P"}
	computedMap := status.ComputedValues()
	computed := make([]shared.ComputedValue, 0, len(computedOrder))
	for _, key := range computedOrder {
		computed = append(computed, shared.ComputedValue{
			Key:   key,
			Value: computedMap[key],
		})
	}

	// Parameters
	paramOrder := []string{"HP", "MP", "SAN"}
	defaults := status.DefaultParameters()
	parameters := make([]shared.StatusParameter, 0, len(paramOrder))
	for _, key := range paramOrder {
		var val *int
		if v := status.Parameters[key]; v != nil {
			val = v
		}
		parameters = append(parameters, shared.StatusParameter{
			Key:          key,
			Value:        val,
			DefaultValue: defaults[key],
		})
	}

	// Skills - build from categories
	skillCategories := make([]shared.SkillCategory, 0, len(CategoryOrder))
	for _, cat := range CategoryOrder {
		catData := skills.Categories[cat]
		var singleSkills, multiSkills []shared.Skill
		for key, s := range catData.Skills {
			init := status.SkillInitialValue(key)
			skill := shared.Skill{
				Key:       key,
				Category:  string(cat),
				Init:      init,
				Order:     s.Order,
				Essential: IsEssentialSkill(key),
			}
			if s.IsMulti() {
				genres := make([]shared.SkillGenre, len(s.Multi.Genres))
				for i, g := range s.Multi.Genres {
					genres[i] = shared.SkillGenre{
						Label: g.Label,
						Job:   g.Job,
						Hobby: g.Hobby,
						Perm:  g.Perm,
						Temp:  g.Temp,
						Grow:  g.Grow,
					}
				}
				skill.Multi = &shared.MultiSkillData{Genres: genres}
				multiSkills = append(multiSkills, skill)
			} else if s.IsSingle() {
				skill.Single = &shared.SingleSkillData{
					Job:   s.Single.Job,
					Hobby: s.Single.Hobby,
					Perm:  s.Single.Perm,
					Temp:  s.Single.Temp,
					Grow:  s.Single.Grow,
				}
				singleSkills = append(singleSkills, skill)
			}
		}
		sort.Slice(singleSkills, func(i, j int) bool {
			return singleSkills[i].Order < singleSkills[j].Order
		})
		sort.Slice(multiSkills, func(i, j int) bool {
			return multiSkills[i].Order < multiSkills[j].Order
		})
		skillCategories = append(skillCategories, shared.SkillCategory{
			Name:         string(cat),
			SingleSkills: singleSkills,
			MultiSkills:  multiSkills,
		})
	}

	skillExtra := shared.SkillExtra{
		Job:   skills.Extra.Job,
		Hobby: skills.Extra.Hobby,
	}

	remJob, remHobby := status.RemainingPoints(skills)
	skillPoints := shared.SkillPoints{
		Job:   remJob,
		Hobby: remHobby,
	}

	return variables, computed, parameters, status.DamageBonus(), skillCategories, skillExtra, skillPoints
}

// BuildSkill creates a template Skill from CoC6 skill data
func BuildSkill(status *Status, key string, skill Skill) shared.Skill {
	init := status.SkillInitialValue(key)
	templSkill := shared.Skill{
		Key:       key,
		Init:      init,
		Order:     skill.Order,
		Essential: IsEssentialSkill(key),
	}
	if skill.IsSingle() {
		templSkill.Single = &shared.SingleSkillData{
			Job:   skill.Single.Job,
			Hobby: skill.Single.Hobby,
			Perm:  skill.Single.Perm,
			Temp:  skill.Single.Temp,
			Grow:  skill.Single.Grow,
		}
	}
	return templSkill
}

// BuildRemainingPoints creates template SkillPoints from CoC6 data
func BuildRemainingPoints(status *Status, skills *Skills) shared.SkillPoints {
	remJob, remHobby := status.RemainingPoints(skills)
	return shared.SkillPoints{Job: remJob, Hobby: remHobby}
}
