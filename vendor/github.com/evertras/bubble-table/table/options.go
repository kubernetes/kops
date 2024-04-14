package table

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// WithHighlightedRow sets the highlighted row to the given index.
func (m Model) WithHighlightedRow(index int) Model {
	m.rowCursorIndex = index

	if m.rowCursorIndex >= len(m.GetVisibleRows()) {
		m.rowCursorIndex = len(m.GetVisibleRows()) - 1
	}

	if m.rowCursorIndex < 0 {
		m.rowCursorIndex = 0
	}

	m.currentPage = m.expectedPageForRowIndex(m.rowCursorIndex)

	return m
}

// HeaderStyle sets the style to apply to the header text, such as color or bold.
func (m Model) HeaderStyle(style lipgloss.Style) Model {
	m.headerStyle = style.Copy()

	return m
}

// WithRows sets the rows to show as data in the table.
func (m Model) WithRows(rows []Row) Model {
	m.rows = rows
	m.visibleRowCacheUpdated = false

	if m.rowCursorIndex >= len(m.rows) {
		m.rowCursorIndex = len(m.rows) - 1
	}

	if m.rowCursorIndex < 0 {
		m.rowCursorIndex = 0
	}

	if m.pageSize != 0 {
		maxPage := m.MaxPages()

		// MaxPages is 1-index, currentPage is 0 index
		if maxPage <= m.currentPage {
			m.pageLast()
		}
	}

	return m
}

// WithKeyMap sets the key map to use for controls when focused.
func (m Model) WithKeyMap(keyMap KeyMap) Model {
	m.keyMap = keyMap

	return m
}

// KeyMap returns a copy of the current key map in use.
func (m Model) KeyMap() KeyMap {
	return m.keyMap
}

// SelectableRows sets whether or not rows are selectable.  If set, adds a column
// in the front that acts as a checkbox and responds to controls if Focused.
func (m Model) SelectableRows(selectable bool) Model {
	m.selectableRows = selectable

	hasSelectColumn := len(m.columns) > 0 && m.columns[0].key == columnKeySelect

	if hasSelectColumn != selectable {
		if selectable {
			m.columns = append([]Column{
				NewColumn(columnKeySelect, m.selectedText, len([]rune(m.selectedText))),
			}, m.columns...)
		} else {
			m.columns = m.columns[1:]
		}
	}

	m.recalculateWidth()

	return m
}

// HighlightedRow returns the full Row that's currently highlighted by the user.
func (m Model) HighlightedRow() Row {
	if len(m.GetVisibleRows()) > 0 {
		return m.GetVisibleRows()[m.rowCursorIndex]
	}

	// TODO: Better way to do this without pointers/nil?  Or should it be nil?
	return Row{}
}

// SelectedRows returns all rows that have been set as selected by the user.
func (m Model) SelectedRows() []Row {
	selectedRows := []Row{}

	for _, row := range m.GetVisibleRows() {
		if row.selected {
			selectedRows = append(selectedRows, row)
		}
	}

	return selectedRows
}

// HighlightStyle sets a custom style to use when the row is being highlighted
// by the cursor.
func (m Model) HighlightStyle(style lipgloss.Style) Model {
	m.highlightStyle = style

	return m
}

// Focused allows the table to show highlighted rows and take in controls of
// up/down/space/etc to let the user navigate the table and interact with it.
func (m Model) Focused(focused bool) Model {
	m.focused = focused

	return m
}

// Filtered allows the table to show rows that match the filter.
func (m Model) Filtered(filtered bool) Model {
	m.filtered = filtered
	m.visibleRowCacheUpdated = false

	return m
}

// StartFilterTyping focuses the text input to allow user typing to filter.
func (m Model) StartFilterTyping() Model {
	m.filterTextInput.Focus()

	return m
}

// WithStaticFooter adds a footer that only displays the given text.
func (m Model) WithStaticFooter(footer string) Model {
	m.staticFooter = footer

	return m
}

// WithPageSize enables pagination using the given page size.  This can be called
// again at any point to resize the height of the table.
func (m Model) WithPageSize(pageSize int) Model {
	m.pageSize = pageSize

	maxPages := m.MaxPages()

	if m.currentPage >= maxPages {
		m.currentPage = maxPages - 1
	}

	return m
}

// WithNoPagination disables pagination in the table.
func (m Model) WithNoPagination() Model {
	m.pageSize = 0

	return m
}

// WithPaginationWrapping sets whether to wrap around from the beginning to the
// end when navigating through pages.  Defaults to true.
func (m Model) WithPaginationWrapping(wrapping bool) Model {
	m.paginationWrapping = wrapping

	return m
}

// WithSelectedText describes what text to show when selectable rows are enabled.
// The selectable column header will use the selected text string.
func (m Model) WithSelectedText(unselected, selected string) Model {
	m.selectedText = selected
	m.unselectedText = unselected

	if len(m.columns) > 0 && m.columns[0].key == columnKeySelect {
		m.columns[0] = NewColumn(columnKeySelect, m.selectedText, len([]rune(m.selectedText)))
		m.recalculateWidth()
	}

	return m
}

// WithBaseStyle applies a base style as the default for everything in the table.
// This is useful for border colors, default alignment, default color, etc.
func (m Model) WithBaseStyle(style lipgloss.Style) Model {
	m.baseStyle = style

	return m
}

// WithTargetWidth sets the total target width of the table, including borders.
// This only takes effect when using flex columns.  When using flex columns,
// columns will stretch to fill out to the total width given here.
func (m Model) WithTargetWidth(totalWidth int) Model {
	m.targetTotalWidth = totalWidth

	m.recalculateWidth()

	return m
}

// PageDown goes to the next page of a paginated table, wrapping to the first
// page if the table is already on the last page.
func (m Model) PageDown() Model {
	m.pageDown()

	return m
}

// PageUp goes to the previous page of a paginated table, wrapping to the
// last page if the table is already on the first page.
func (m Model) PageUp() Model {
	m.pageUp()

	return m
}

// PageLast goes to the last page of a paginated table.
func (m Model) PageLast() Model {
	m.pageLast()

	return m
}

// PageFirst goes to the first page of a paginated table.
func (m Model) PageFirst() Model {
	m.pageFirst()

	return m
}

// WithCurrentPage sets the current page (1 as the first page) of a paginated
// table, bounded to the total number of pages.  The current selected row will
// be set to the top row of the page if the page changed.
func (m Model) WithCurrentPage(currentPage int) Model {
	if m.pageSize == 0 || currentPage == m.CurrentPage() {
		return m
	}

	if currentPage < 1 {
		currentPage = 1
	} else {
		maxPages := m.MaxPages()

		if currentPage > maxPages {
			currentPage = maxPages
		}
	}

	m.currentPage = currentPage - 1
	m.rowCursorIndex = m.currentPage * m.pageSize

	return m
}

// WithColumns sets the visible columns for the table, so that columns can be
// added/removed/resized or headers rewritten.
func (m Model) WithColumns(columns []Column) Model {
	// Deep copy to avoid edits
	m.columns = make([]Column, len(columns))
	copy(m.columns, columns)

	m.recalculateWidth()

	return m
}

// WithFilterInput makes the table use the provided text input bubble for
// filtering rather than using the built-in default.  This allows for external
// text input controls to be used.
func (m Model) WithFilterInput(input textinput.Model) Model {
	if m.filterTextInput.Value() != input.Value() {
		m.pageFirst()
	}

	m.filterTextInput = input
	m.visibleRowCacheUpdated = false

	return m
}

// WithFilterInputValue sets the filter value to the given string, immediately
// applying it as if the user had typed it in.  Useful for external filter inputs
// that are not necessarily a text input.
func (m Model) WithFilterInputValue(value string) Model {
	if m.filterTextInput.Value() != value {
		m.pageFirst()
	}

	m.filterTextInput.SetValue(value)
	m.filterTextInput.Blur()
	m.visibleRowCacheUpdated = false

	return m
}

// WithFooterVisibility sets the visibility of the footer.
func (m Model) WithFooterVisibility(visibility bool) Model {
	m.footerVisible = visibility

	return m
}

// WithHeaderVisibility sets the visibility of the header.
func (m Model) WithHeaderVisibility(visibility bool) Model {
	m.headerVisible = visibility

	return m
}

// WithMaxTotalWidth sets the maximum total width that the table should render.
// If this width is exceeded by either the target width or by the total width
// of all the columns (including borders!), anything extra will be treated as
// overflow and horizontal scrolling will be enabled to see the rest.
func (m Model) WithMaxTotalWidth(maxTotalWidth int) Model {
	m.maxTotalWidth = maxTotalWidth

	m.recalculateWidth()

	return m
}

// WithHorizontalFreezeColumnCount freezes the given number of columns to the
// left side.  This is useful for things like ID or Name columns that should
// always be visible even when scrolling.
func (m Model) WithHorizontalFreezeColumnCount(columnsToFreeze int) Model {
	m.horizontalScrollFreezeColumnsCount = columnsToFreeze

	m.recalculateWidth()

	return m
}

// ScrollRight moves one column to the right.  Use with WithMaxTotalWidth.
func (m Model) ScrollRight() Model {
	m.scrollRight()

	return m
}

// ScrollLeft moves one column to the left.  Use with WithMaxTotalWidth.
func (m Model) ScrollLeft() Model {
	m.scrollLeft()

	return m
}

// WithMissingDataIndicator sets an indicator to use when data for a column is
// not found in a given row.  Note that this is for completely missing data,
// an empty string or other zero value that is explicitly set is not considered
// to be missing.
func (m Model) WithMissingDataIndicator(str string) Model {
	m.missingDataIndicator = str

	return m
}

// WithMissingDataIndicatorStyled sets a styled indicator to use when data for
// a column is not found in a given row.  Note that this is for completely
// missing data, an empty string or other zero value that is explicitly set is
// not considered to be missing.
func (m Model) WithMissingDataIndicatorStyled(styled StyledCell) Model {
	m.missingDataIndicator = styled

	return m
}

// WithAllRowsDeselected deselects any rows that are currently selected.
func (m Model) WithAllRowsDeselected() Model {
	rows := m.GetVisibleRows()

	for i, row := range rows {
		if row.selected {
			rows[i] = row.Selected(false)
		}
	}

	m.rows = rows

	return m
}
