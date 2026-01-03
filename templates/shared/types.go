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
