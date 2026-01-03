package models

import "sync"

// Store holds the in-memory character data
type Store struct {
	mu     sync.RWMutex
	status *Cthulhu6Status
	memos  map[string]string
}

// NewStore creates a new store with default data
func NewStore() *Store {
	return &Store{
		status: NewCthulhu6Status(),
		memos:  make(map[string]string),
	}
}

// GetStatus returns the current status
func (s *Store) GetStatus() *Cthulhu6Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// UpdateVariableBase updates a variable's base value
func (s *Store) UpdateVariableBase(key string, delta int) *Variable {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.status.Variables[key]; ok {
		newBase := v.Base + delta
		if newBase < v.Min {
			newBase = v.Min
		}
		if newBase > v.Max {
			newBase = v.Max
		}
		v.Base = newBase
		s.status.Variables[key] = v
		return &v
	}
	return nil
}

// SetVariableBase sets a variable's base value directly
func (s *Store) SetVariableBase(key string, value int) *Variable {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.status.Variables[key]; ok {
		if value < v.Min {
			value = v.Min
		}
		if value > v.Max {
			value = v.Max
		}
		v.Base = value
		s.status.Variables[key] = v
		return &v
	}
	return nil
}

// UpdateParameter updates a parameter value
func (s *Store) UpdateParameter(key string, delta int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.status.EffectiveParameter(key)
	newVal := current + delta
	if newVal < 0 {
		newVal = 0
	}
	s.status.Parameters[key] = &newVal
	return newVal
}

// SetParameter sets a parameter value directly
func (s *Store) SetParameter(key string, value int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if value < 0 {
		value = 0
	}
	s.status.Parameters[key] = &value
}

// SetDamageBonus sets the damage bonus
func (s *Store) SetDamageBonus(db string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.DB = db
}

// GetMemo returns a memo by ID
func (s *Store) GetMemo(id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.memos[id]
}

// SetMemo sets a memo value, returns true if changed
func (s *Store) SetMemo(id string, value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.memos[id] == value {
		return false
	}
	s.memos[id] = value
	return true
}
