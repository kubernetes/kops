package table

// GetColumnSorting returns the current sorting rules for the table as a list of
// SortColumns, which are applied from first to last.  This means that data will
// be grouped by the later elements in the list.  The returned list is a copy
// and modifications will have no effect.
func (m *Model) GetColumnSorting() []SortColumn {
	c := make([]SortColumn, len(m.sortOrder))

	copy(c, m.sortOrder)

	return c
}

// GetCanFilter returns true if the table enables filtering at all.  This does
// not say whether a filter is currently active, only that the feature is enabled.
func (m *Model) GetCanFilter() bool {
	return m.filtered
}

// GetIsFilterActive returns true if the table is currently being filtered.  This
// does not say whether the table CAN be filtered, only whether or not a filter
// is actually currently being applied.
func (m *Model) GetIsFilterActive() bool {
	return m.filterTextInput.Value() != ""
}

// GetCurrentFilter returns the current filter text being applied, or an empty
// string if none is applied.
func (m *Model) GetCurrentFilter() string {
	return m.filterTextInput.Value()
}

// GetVisibleRows returns sorted and filtered rows.
func (m Model) GetVisibleRows() []Row {
	rows := make([]Row, len(m.rows))
	copy(rows, m.rows)
	if m.filtered {
		rows = m.getFilteredRows(rows)
	}
	rows = getSortedRows(m.sortOrder, rows)

	return rows
}

// GetHighlightedRowIndex returns the index of the Row that's currently highlighted
// by the user.
func (m *Model) GetHighlightedRowIndex() int {
	return m.rowCursorIndex
}

// GetFocused returns whether or not the table is focused and is receiving inputs.
func (m *Model) GetFocused() bool {
	return m.focused
}

// GetHorizontalScrollColumnOffset returns how many columns to the right the table
// has been scrolled.  0 means the table is all the way to the left, which is
// the starting default.
func (m *Model) GetHorizontalScrollColumnOffset() int {
	return m.horizontalScrollOffsetCol
}

// GetHeaderVisibility returns true if the header has been set to visible (default)
// or false if the header has been set to hidden.
func (m *Model) GetHeaderVisibility() bool {
	return m.headerVisible
}

// GetPaginationWrapping returns true if pagination wrapping is enabled, or false
// if disabled.  If disabled, navigating through pages will stop at the first
// and last pages.
func (m *Model) GetPaginationWrapping() bool {
	return m.paginationWrapping
}
