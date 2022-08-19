package table

import "github.com/charmbracelet/lipgloss"

// StyledCell represents a cell in the table that has a particular style applied.
// The cell style takes highest precedence and will overwrite more general styles
// from the row, column, or table as a whole.  This style should be generally
// limited to colors, font style, and alignments - spacing style such as margin
// will break the table format.
type StyledCell struct {
	Data  interface{}
	Style lipgloss.Style
}

// NewStyledCell creates an entry that can be set in the row data and show as
// styled with the given style.
func NewStyledCell(data interface{}, style lipgloss.Style) StyledCell {
	return StyledCell{data, style}
}
