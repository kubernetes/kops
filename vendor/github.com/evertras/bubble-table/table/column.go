package table

import (
	"github.com/charmbracelet/lipgloss"
)

// Column is a column in the table.
type Column struct {
	title string
	key   string
	width int

	flexFactor int

	filterable bool
	style      lipgloss.Style
}

// NewColumn creates a new fixed-width column with the given information.
func NewColumn(key, title string, width int) Column {
	return Column{
		key:   key,
		title: title,
		width: width,

		filterable: false,
	}
}

// NewFlexColumn creates a new flexible width column that tries to fill in the
// total table width.  If multiple flex columns exist, each will measure against
// each other depending on their flexFactor.  For example, if both have a flexFactor
// of 1, they will have equal width.  If one has a flexFactor of 1 and the other
// has a flexFactor of 3, the second will be 3 times larger than the first.  You
// must use WithTargetWidth if you have any flex columns, so that the table knows
// how much width it should fill.
func NewFlexColumn(key, title string, flexFactor int) Column {
	return Column{
		key:   key,
		title: title,

		flexFactor: max(flexFactor, 1),
	}
}

// WithStyle applies a style to the column as a whole.
func (c Column) WithStyle(style lipgloss.Style) Column {
	c.style = style.Copy().Width(c.width)

	return c
}

// WithFiltered sets whether the column should be considered for filtering (true)
// or not (false).
func (c Column) WithFiltered(filterable bool) Column {
	c.filterable = filterable

	return c
}

func (c *Column) isFlex() bool {
	return c.flexFactor != 0
}
