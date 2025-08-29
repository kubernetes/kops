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

	fmtString string
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

// WithFormatString sets the format string used by fmt.Sprintf to display the data.
// If not set, the default is "%v" for all data types.  Intended mainly for
// numeric formatting.
//
// Since data is of the interface{} type, make sure that all data in the column
// is of the expected type or the format may fail.  For example, hardcoding '3'
// instead of '3.0' and using '%.2f' will fail because '3' is an integer.
func (c Column) WithFormatString(fmtString string) Column {
	c.fmtString = fmtString

	return c
}

func (c *Column) isFlex() bool {
	return c.flexFactor != 0
}

// Title returns the title of the column.
func (c Column) Title() string {
	return c.title
}

// Key returns the key of the column.
func (c Column) Key() string {
	return c.key
}

// Width returns the width of the column.
func (c Column) Width() int {
	return c.width
}

// FlexFactor returns the flex factor of the column.
func (c Column) FlexFactor() int {
	return c.flexFactor
}

// IsFlex returns whether the column is a flex column.
func (c Column) IsFlex() bool {
	return c.isFlex()
}

// Filterable returns whether the column is filterable.
func (c Column) Filterable() bool {
	return c.filterable
}

// Style returns the style of the column.
func (c Column) Style() lipgloss.Style {
	return c.style
}

// FmtString returns the format string of the column.
func (c Column) FmtString() string {
	return c.fmtString
}
