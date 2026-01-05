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
// SingleSkillData holds point allocations for a single skill
type SingleSkillData struct {
	Job   int
	Hobby int
	Perm  int
	Temp  int
	Grow  bool
}

// Total returns total allocated points
func (s *SingleSkillData) Total() int {
	return s.Job + s.Hobby + s.Perm + s.Temp
}

// SkillGenre represents one specialty within a multi-genre skill
type SkillGenre struct {
	Label string // genre label (e.g., "自動車" for 運転)
	Job   int
	Hobby int
	Perm  int
	Temp  int
	Grow  bool
}

// Total returns total allocated points for this genre
func (g SkillGenre) Total() int {
	return g.Job + g.Hobby + g.Perm + g.Temp
}

// MultiSkillData holds genres for a multi-genre skill
type MultiSkillData struct {
	Genres []SkillGenre
}

// Skill represents a skill (exactly one of Single or Multi will be non-nil)
type Skill struct {
	Key       string           // skill key/name
	Category  string           // skill category (e.g., "戦闘技能")
	Init      int              // initial value (from character stats)
	Order     int              // display order within category
	Essential bool             // true for important/essential skills (shown bold)
	Single    *SingleSkillData // non-nil for single skills
	Multi     *MultiSkillData  // non-nil for multi-genre skills
}

// IsSingle returns true if this is a single skill
func (s Skill) IsSingle() bool {
	return s.Single != nil
}

// IsMulti returns true if this is a multi-genre skill
func (s Skill) IsMulti() bool {
	return s.Multi != nil
}

// Total returns the total skill value (init + allocated) for single skills
func (s Skill) Total() int {
	if s.Single != nil {
		return s.Init + s.Single.Total()
	}
	return s.Init
}

// Allocated returns total allocated points (job + hobby + perm + temp)
// Allocated returns total allocated points for single skills
func (s Skill) Allocated() int {
	if s.Single != nil {
		return s.Single.Total()
	}
	return 0
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
	Name         string  // category name (e.g., "戦闘技能")
	SingleSkills []Skill // single skills in this category, sorted by order
	MultiSkills  []Skill // multi-genre skills in this category, sorted by order
}

// AllSkills returns all skills merged and sorted by Order
func (c SkillCategory) AllSkills() []Skill {
	all := make([]Skill, 0, len(c.SingleSkills)+len(c.MultiSkills))
	i, j := 0, 0
	for i < len(c.SingleSkills) && j < len(c.MultiSkills) {
		if c.SingleSkills[i].Order <= c.MultiSkills[j].Order {
			all = append(all, c.SingleSkills[i])
			i++
		} else {
			all = append(all, c.MultiSkills[j])
			j++
		}
	}
	all = append(all, c.SingleSkills[i:]...)
	all = append(all, c.MultiSkills[j:]...)
	return all
}

// CustomSkill represents a user-defined skill
type CustomSkill struct {
	Name  string // user-defined skill name
	Job   int
	Hobby int
	Perm  int
	Temp  int
	Grow  bool
}

// Total returns total allocated points for this custom skill
func (c CustomSkill) Total() int {
	return c.Job + c.Hobby + c.Perm + c.Temp
}

// IsActive returns true if the skill has any points allocated or grow checked
func (c CustomSkill) IsActive() bool {
	return c.Total() > 0 || c.Grow
}

// SkillsState holds all skills-related data for rendering
type SkillsState struct {
	Categories   []SkillCategory // categories in display order
	CustomSkills []CustomSkill   // user-defined custom skills
	Extra        SkillExtra
	Remaining    SkillPoints
}

// SheetState holds all data needed to render a character sheet
type SheetState struct {
	PC     PageContext
	Status StatusState
	Skills SkillsState
}
