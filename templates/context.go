package templates

// PageContext holds rendering context for the character sheet
type PageContext struct {
	// IsOwner is true if the current user owns this character
	IsOwner bool
}

// NewPageContext creates a PageContext with default values
func NewPageContext() PageContext {
	return PageContext{
		IsOwner: true, // Hard-coded for development
	}
}
