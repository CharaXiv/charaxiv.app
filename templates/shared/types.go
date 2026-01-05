package shared

// StatusVariable represents a single ability score with base, perm, and temp values
type StatusVariable struct {
	Key  string
	Base int
	Perm int
	Temp int
	Min  int
	Max  int
}

// Sum returns the total value of the variable
func (v StatusVariable) Sum() int {
	return v.Base + v.Perm + v.Temp
}

// ComputedValue represents a derived value calculated from variables
type ComputedValue struct {
	Key   string
	Value int
}

// StatusParameter represents an editable parameter with a default value
type StatusParameter struct {
	Key          string
	Value        *int // nil means use default
	DefaultValue int
}

// EffectiveValue returns the value or default if nil
func (p StatusParameter) EffectiveValue() int {
	if p.Value != nil {
		return *p.Value
	}
	return p.DefaultValue
}

// SkillPoints represents remaining/consumed skill points
type SkillPoints struct {
	Job   int // 職業ポイント (occupation points)
	Hobby int // 興味ポイント (hobby points)
}

// SkillExtra represents bonus skill points
type SkillExtra struct {
	Job   int
	Hobby int
}

// Skill represents a single skill with its point allocations
// SkillGenre represents a single genre/specialty within a multi-genre skill
type SkillGenre struct {
	Index int    // index within the skill's genres
	Label string // genre label (e.g., "自動車" for 運転)
	Init  int    // initial value (same as parent skill)
	Job   int
	Hobby int
	Perm  int
	Temp  int
	Grow  bool
}

// Total returns the total value for this genre
func (g SkillGenre) Total() int {
	return g.Init + g.Job + g.Hobby + g.Perm + g.Temp
}

// Allocated returns total allocated points
func (g SkillGenre) Allocated() int {
	return g.Job + g.Hobby + g.Perm + g.Temp
}

// Skill represents a single skill (may be single or multi-genre)
type Skill struct {
	Key      string       // skill key/name
	Category string       // skill category (e.g., "戦闘技能")
	Init     int          // initial value (from character stats)
	Job      int          // occupation points allocated (for single skills)
	Hobby    int          // hobby points allocated (for single skills)
	Perm     int          // permanent increase (for single skills)
	Temp     int          // temporary increase (for single skills)
	Grow     bool         // marked for growth check (for single skills)
	Order    int          // display order within category
	Multi    bool         // true if this is a multi-genre skill
	Genres   []SkillGenre // genres for multi-genre skills
}

// Total returns the total skill value (init + all bonuses) for single skills
func (s Skill) Total() int {
	return s.Init + s.Job + s.Hobby + s.Perm + s.Temp
}

// Allocated returns total allocated points (job + hobby + perm + temp)
func (s Skill) Allocated() int {
	return s.Job + s.Hobby + s.Perm + s.Temp
}

// StatusState holds all status-related data for rendering
type StatusState struct {
	Variables   []StatusVariable
	Computed    []ComputedValue
	Parameters  []StatusParameter
	DamageBonus string
}

// SkillCategory represents a group of skills
type SkillCategory struct {
	Name   string  // category name (e.g., "戦闘技能")
	Skills []Skill // skills in this category, sorted by order
}

// SkillsState holds all skills-related data for rendering
type SkillsState struct {
	Categories []SkillCategory // categories in display order
	Extra      SkillExtra
	Remaining  SkillPoints
}

// SheetState holds all data needed to render a character sheet
type SheetState struct {
	PC     PageContext
	Status StatusState
	Skills SkillsState
}
