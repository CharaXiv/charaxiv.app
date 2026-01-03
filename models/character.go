package models

// Variable represents an ability score with base, perm, and temp modifiers
type Variable struct {
	Base int `json:"base"`
	Perm int `json:"perm"`
	Temp int `json:"temp"`
	Min  int `json:"min"`
	Max  int `json:"max"`
}

// Sum returns the total value of the variable
func (v Variable) Sum() int {
	return v.Base + v.Perm + v.Temp
}

// Cthulhu6Status represents the status section for CoC 6th edition
type Cthulhu6Status struct {
	Variables  map[string]Variable `json:"variables"`
	Parameters map[string]*int     `json:"parameters"` // nil means use default
	DB         string              `json:"db"`
}

// NewCthulhu6Status creates a new status with default values
func NewCthulhu6Status() *Cthulhu6Status {
	return &Cthulhu6Status{
		Variables: map[string]Variable{
			"STR": {Base: 10, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"CON": {Base: 12, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"POW": {Base: 11, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"DEX": {Base: 13, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"APP": {Base: 9, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"SIZ": {Base: 14, Perm: 0, Temp: 0, Min: 8, Max: 18},
			"INT": {Base: 15, Perm: 0, Temp: 0, Min: 8, Max: 18},
			"EDU": {Base: 16, Perm: 0, Temp: 0, Min: 6, Max: 21},
		},
		Parameters: map[string]*int{
			"HP":  nil,
			"MP":  nil,
			"SAN": nil,
		},
		DB: "",
	}
}

// ComputedValues returns the derived values from variables
func (s *Cthulhu6Status) ComputedValues() map[string]int {
	pow := s.Variables["POW"].Sum()
	inT := s.Variables["INT"].Sum()
	edu := s.Variables["EDU"].Sum()

	return map[string]int{
		"初期SAN": pow * 5,
		"アイデア":  inT * 5,
		"幸運":    pow * 5,
		"知識":    edu * 5,
		"職業P":   edu * 20,
		"興味P":   inT * 10,
	}
}

// DefaultParameters returns the default parameter values derived from variables
func (s *Cthulhu6Status) DefaultParameters() map[string]int {
	con := s.Variables["CON"].Sum()
	pow := s.Variables["POW"].Sum()
	siz := s.Variables["SIZ"].Sum()

	return map[string]int{
		"HP":  (con + siz + 1) / 2,
		"MP":  pow,
		"SAN": pow * 5,
	}
}

// DamageBonus calculates the damage bonus from STR+SIZ
func (s *Cthulhu6Status) DamageBonus() string {
	if s.DB != "" {
		return s.DB
	}
	str := s.Variables["STR"].Sum()
	siz := s.Variables["SIZ"].Sum()
	sum := str + siz

	switch {
	case sum < 13:
		return "-1d6"
	case sum < 17:
		return "-1d4"
	case sum < 25:
		return "+0"
	case sum < 33:
		return "+1d4"
	default:
		return "+1d6"
	}
}

// Indefinite calculates the indefinite insanity threshold
func (s *Cthulhu6Status) Indefinite() int {
	san := s.EffectiveParameter("SAN")
	return (san * 4) / 5
}

// EffectiveParameter returns the parameter value or its default
func (s *Cthulhu6Status) EffectiveParameter(key string) int {
	if val := s.Parameters[key]; val != nil {
		return *val
	}
	return s.DefaultParameters()[key]
}
