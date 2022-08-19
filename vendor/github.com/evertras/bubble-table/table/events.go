package table

// UserEvent is some state change that has occurred due to user input.  These will
// ONLY be generated when a user has interacted directly with the table.  These
// will NOT be generated when code programmatically changes values in the table.
type UserEvent interface{}

func (m *Model) appendUserEvent(e UserEvent) {
	m.lastUpdateUserEvents = append(m.lastUpdateUserEvents, e)
}

func (m *Model) clearUserEvents() {
	m.lastUpdateUserEvents = nil
}

// GetLastUpdateUserEvents returns a list of events that happened due to user
// input in the last Update call.  This is useful to look for triggers such as
// whether the user moved to a new highlighted row.
func (m *Model) GetLastUpdateUserEvents() []UserEvent {
	// Most common case
	if len(m.lastUpdateUserEvents) == 0 {
		return nil
	}

	returned := make([]UserEvent, len(m.lastUpdateUserEvents))

	// Slightly wasteful but helps guarantee immutability, and this should only
	// have data very rarely so this is fine
	copy(returned, m.lastUpdateUserEvents)

	return returned
}

// UserEventHighlightedIndexChanged indicates that the user has scrolled to a new
// row.
type UserEventHighlightedIndexChanged struct {
	// PreviousRow is the row that was selected before the change.
	PreviousRowIndex int

	// SelectedRow is the row index that is now selected
	SelectedRowIndex int
}

// UserEventRowSelectToggled indicates that the user has either selected or
// deselected a row by toggling the selection.  The event contains information
// about which row index was selected and whether it was selected or deselected.
type UserEventRowSelectToggled struct {
	RowIndex   int
	IsSelected bool
}
