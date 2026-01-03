package shared

// PageContext holds rendering context for the character sheet
type PageContext struct {
	// IsOwner is true if the current user owns this character
	IsOwner bool
	// Preview is true when the owner is previewing as a visitor
	Preview bool
	// Memos holds memo content by ID
	Memos map[string]string
}

// IsReadOnly returns true if the character should be displayed in read-only mode.
// This is true when the user is not the owner OR when preview mode is active.
func (pc PageContext) IsReadOnly() bool {
	return !pc.IsOwner || pc.Preview
}

// NewPageContext creates a PageContext with default values
func NewPageContext() PageContext {
	return PageContext{
		IsOwner: true, // Hard-coded for development
		Preview: false,
		Memos:   make(map[string]string),
	}
}

// GetMemo returns the memo content for the given ID
func (pc PageContext) GetMemo(id string) string {
	if pc.Memos == nil {
		return ""
	}
	return pc.Memos[id]
}
