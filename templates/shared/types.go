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
type Skill struct {
	Key   string // skill key/name
	Init  int    // initial value (from character stats)
	Job   int    // occupation points allocated
	Hobby int    // hobby points allocated
	Perm  int    // permanent increase
	Temp  int    // temporary increase
	Grow  bool   // marked for growth check
	Order int    // display order
}

// Total returns the total skill value (init + all bonuses)
func (s Skill) Total() int {
	return s.Init + s.Job + s.Hobby + s.Perm + s.Temp
}

// Allocated returns total allocated points (job + hobby + perm + temp)
func (s Skill) Allocated() int {
	return s.Job + s.Hobby + s.Perm + s.Temp
}
