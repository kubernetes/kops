package table

import "github.com/charmbracelet/lipgloss"

// Border defines the borders in and around the table.
type Border struct {
	Top         string
	Left        string
	Right       string
	Bottom      string
	TopRight    string
	TopLeft     string
	BottomRight string
	BottomLeft  string

	TopJunction    string
	LeftJunction   string
	RightJunction  string
	BottomJunction string

	InnerJunction string

	InnerDivider string

	// Styles for 2x2 tables and larger
	styleMultiTopLeft     lipgloss.Style
	styleMultiTop         lipgloss.Style
	styleMultiTopRight    lipgloss.Style
	styleMultiRight       lipgloss.Style
	styleMultiBottomRight lipgloss.Style
	styleMultiBottom      lipgloss.Style
	styleMultiBottomLeft  lipgloss.Style
	styleMultiLeft        lipgloss.Style
	styleMultiInner       lipgloss.Style

	// Styles for a single column table
	styleSingleColumnTop    lipgloss.Style
	styleSingleColumnInner  lipgloss.Style
	styleSingleColumnBottom lipgloss.Style

	// Styles for a single row table
	styleSingleRowLeft  lipgloss.Style
	styleSingleRowInner lipgloss.Style
	styleSingleRowRight lipgloss.Style

	// Style for a table with only one cell
	styleSingleCell lipgloss.Style

	// Style for the footer
	styleFooter lipgloss.Style
}

var (
	// https://www.w3.org/TR/xml-entity-names/025.html

	borderDefault = Border{
		Top:    "━",
		Left:   "┃",
		Right:  "┃",
		Bottom: "━",

		TopRight:    "┓",
		TopLeft:     "┏",
		BottomRight: "┛",
		BottomLeft:  "┗",

		TopJunction:    "┳",
		LeftJunction:   "┣",
		RightJunction:  "┫",
		BottomJunction: "┻",
		InnerJunction:  "╋",

		InnerDivider: "┃",
	}

	borderRounded = Border{
		Top:    "─",
		Left:   "│",
		Right:  "│",
		Bottom: "─",

		TopRight:    "╮",
		TopLeft:     "╭",
		BottomRight: "╯",
		BottomLeft:  "╰",

		TopJunction:    "┬",
		LeftJunction:   "├",
		RightJunction:  "┤",
		BottomJunction: "┴",
		InnerJunction:  "┼",

		InnerDivider: "│",
	}
)

func init() {
	borderDefault.generateStyles()
	borderRounded.generateStyles()
}

func (b *Border) generateStyles() {
	b.generateMultiStyles()
	b.generateSingleColumnStyles()
	b.generateSingleRowStyles()
	b.generateSingleCellStyle()

	// The footer is a single cell with the top taken off... usually.  We can
	// re-enable the top if needed this way for certain format configurations.
	b.styleFooter = b.styleSingleCell.Copy().
		Align(lipgloss.Right).
		BorderBottom(true).
		BorderRight(true).
		BorderLeft(true)
}

func (b *Border) styleLeftWithFooter(original lipgloss.Style) lipgloss.Style {
	border := original.GetBorderStyle()

	border.BottomLeft = b.LeftJunction

	return original.Copy().BorderStyle(border)
}

func (b *Border) styleRightWithFooter(original lipgloss.Style) lipgloss.Style {
	border := original.GetBorderStyle()

	border.BottomRight = b.RightJunction

	return original.Copy().BorderStyle(border)
}

func (b *Border) styleBothWithFooter(original lipgloss.Style) lipgloss.Style {
	border := original.GetBorderStyle()

	border.BottomLeft = b.LeftJunction
	border.BottomRight = b.RightJunction

	return original.Copy().BorderStyle(border)
}

// This function is long, but it's just repetitive...
//
//nolint:funlen
func (b *Border) generateMultiStyles() {
	b.styleMultiTopLeft = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			TopLeft:     b.TopLeft,
			Top:         b.Top,
			TopRight:    b.TopJunction,
			Right:       b.InnerDivider,
			BottomRight: b.InnerJunction,
			Bottom:      b.Bottom,
			BottomLeft:  b.LeftJunction,
			Left:        b.Left,
		},
	)

	b.styleMultiTop = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Right:  b.InnerDivider,
			Bottom: b.Bottom,

			TopRight:    b.TopJunction,
			BottomRight: b.InnerJunction,
		},
	).BorderTop(true).BorderBottom(true).BorderRight(true)

	b.styleMultiTopRight = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Right:  b.Right,
			Bottom: b.Bottom,

			TopRight:    b.TopRight,
			BottomRight: b.RightJunction,
		},
	).BorderTop(true).BorderBottom(true).BorderRight(true)

	b.styleMultiLeft = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Left:  b.Left,
			Right: b.InnerDivider,
		},
	).BorderRight(true).BorderLeft(true)

	b.styleMultiRight = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Right: b.Right,
		},
	).BorderRight(true)

	b.styleMultiInner = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Right: b.InnerDivider,
		},
	).BorderRight(true)

	b.styleMultiBottomLeft = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Left:   b.Left,
			Right:  b.InnerDivider,
			Bottom: b.Bottom,

			BottomLeft:  b.BottomLeft,
			BottomRight: b.BottomJunction,
		},
	).BorderLeft(true).BorderBottom(true).BorderRight(true)

	b.styleMultiBottom = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Right:  b.InnerDivider,
			Bottom: b.Bottom,

			BottomRight: b.BottomJunction,
		},
	).BorderBottom(true).BorderRight(true)

	b.styleMultiBottomRight = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Right:  b.Right,
			Bottom: b.Bottom,

			BottomRight: b.BottomRight,
		},
	).BorderBottom(true).BorderRight(true)
}

func (b *Border) generateSingleColumnStyles() {
	b.styleSingleColumnTop = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Left:   b.Left,
			Right:  b.Right,
			Bottom: b.Bottom,

			TopLeft:     b.TopLeft,
			TopRight:    b.TopRight,
			BottomLeft:  b.LeftJunction,
			BottomRight: b.RightJunction,
		},
	)

	b.styleSingleColumnInner = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Left:  b.Left,
			Right: b.Right,
		},
	).BorderRight(true).BorderLeft(true)

	b.styleSingleColumnBottom = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Left:   b.Left,
			Right:  b.Right,
			Bottom: b.Bottom,

			BottomLeft:  b.BottomLeft,
			BottomRight: b.BottomRight,
		},
	).BorderRight(true).BorderLeft(true).BorderBottom(true)
}

func (b *Border) generateSingleRowStyles() {
	b.styleSingleRowLeft = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Left:   b.Left,
			Right:  b.InnerDivider,
			Bottom: b.Bottom,

			BottomLeft:  b.BottomLeft,
			BottomRight: b.BottomJunction,
			TopRight:    b.TopJunction,
			TopLeft:     b.TopLeft,
		},
	)

	b.styleSingleRowInner = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Right:  b.InnerDivider,
			Bottom: b.Bottom,

			BottomRight: b.BottomJunction,
			TopRight:    b.TopJunction,
		},
	).BorderTop(true).BorderBottom(true).BorderRight(true)

	b.styleSingleRowRight = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Right:  b.Right,
			Bottom: b.Bottom,

			BottomRight: b.BottomRight,
			TopRight:    b.TopRight,
		},
	).BorderTop(true).BorderBottom(true).BorderRight(true)
}

func (b *Border) generateSingleCellStyle() {
	b.styleSingleCell = lipgloss.NewStyle().BorderStyle(
		lipgloss.Border{
			Top:    b.Top,
			Left:   b.Left,
			Right:  b.Right,
			Bottom: b.Bottom,

			BottomLeft:  b.BottomLeft,
			BottomRight: b.BottomRight,
			TopRight:    b.TopRight,
			TopLeft:     b.TopLeft,
		},
	)
}

// BorderDefault uses the basic square border, useful to reset the border if
// it was changed somehow.
func (m Model) BorderDefault() Model {
	// Already generated styles
	m.border = borderDefault

	return m
}

// BorderRounded uses a thin, rounded border.
func (m Model) BorderRounded() Model {
	// Already generated styles
	m.border = borderRounded

	return m
}

// Border uses the given border components to render the table.
func (m Model) Border(border Border) Model {
	border.generateStyles()

	m.border = border

	return m
}

type borderStyleRow struct {
	left  lipgloss.Style
	inner lipgloss.Style
	right lipgloss.Style
}

func (b *borderStyleRow) inherit(s lipgloss.Style) {
	b.left = b.left.Copy().Inherit(s)
	b.inner = b.inner.Copy().Inherit(s)
	b.right = b.right.Copy().Inherit(s)
}

// There's a lot of branches here, but splitting it up further would make it
// harder to follow.  So just be careful with comments and make sure it's tested!
//
//nolint:nestif
func (m Model) styleHeaders() borderStyleRow {
	hasRows := len(m.GetVisibleRows()) > 0
	singleColumn := len(m.columns) == 1
	styles := borderStyleRow{}

	// Possible configurations:
	// - Single cell
	// - Single row
	// - Single column
	// - Multi

	if singleColumn {
		if hasRows {
			// Single column
			styles.left = m.border.styleSingleColumnTop
			styles.inner = styles.left
			styles.right = styles.left
		} else {
			// Single cell
			styles.left = m.border.styleSingleCell
			styles.inner = styles.left
			styles.right = styles.left

			if m.hasFooter() {
				styles.left = m.border.styleBothWithFooter(styles.left)
			}
		}
	} else if !hasRows {
		// Single row
		styles.left = m.border.styleSingleRowLeft
		styles.inner = m.border.styleSingleRowInner
		styles.right = m.border.styleSingleRowRight

		if m.hasFooter() {
			styles.left = m.border.styleLeftWithFooter(styles.left)
			styles.right = m.border.styleRightWithFooter(styles.right)
		}
	} else {
		// Multi
		styles.left = m.border.styleMultiTopLeft
		styles.inner = m.border.styleMultiTop
		styles.right = m.border.styleMultiTopRight
	}

	styles.inherit(m.headerStyle)

	return styles
}

func (m Model) styleRows() (inner borderStyleRow, last borderStyleRow) {
	if len(m.columns) == 1 {
		inner.left = m.border.styleSingleColumnInner
		inner.inner = inner.left
		inner.right = inner.left

		last.left = m.border.styleSingleColumnBottom

		if m.hasFooter() {
			last.left = m.border.styleBothWithFooter(last.left)
		}

		last.inner = last.left
		last.right = last.left
	} else {
		inner.left = m.border.styleMultiLeft
		inner.inner = m.border.styleMultiInner
		inner.right = m.border.styleMultiRight

		last.left = m.border.styleMultiBottomLeft
		last.inner = m.border.styleMultiBottom
		last.right = m.border.styleMultiBottomRight

		if m.hasFooter() {
			last.left = m.border.styleLeftWithFooter(last.left)
			last.right = m.border.styleRightWithFooter(last.right)
		}
	}

	return inner, last
}
