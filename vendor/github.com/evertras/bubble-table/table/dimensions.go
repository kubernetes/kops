package table

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) recalculateWidth() {
	if m.targetTotalWidth != 0 {
		m.totalWidth = m.targetTotalWidth
	} else {
		total := 0

		for _, column := range m.columns {
			total += column.width
		}

		m.totalWidth = total + len(m.columns) + 1
	}

	updateColumnWidths(m.columns, m.targetTotalWidth)

	m.recalculateLastHorizontalColumn()
}

// Updates column width in-place.  This could be optimized but should be called
// very rarely so we prioritize simplicity over performance here.
func updateColumnWidths(cols []Column, totalWidth int) {
	totalFlexWidth := totalWidth - len(cols) - 1
	totalFlexFactor := 0
	flexGCD := 0

	for index, col := range cols {
		if !col.isFlex() {
			totalFlexWidth -= col.width
			cols[index].style = col.style.Width(col.width)
		} else {
			totalFlexFactor += col.flexFactor
			flexGCD = gcd(flexGCD, col.flexFactor)
		}
	}

	if totalFlexFactor == 0 {
		return
	}

	// We use the GCD here because otherwise very large values won't divide
	// nicely as ints
	totalFlexFactor /= flexGCD

	flexUnit := totalFlexWidth / totalFlexFactor
	leftoverWidth := totalFlexWidth % totalFlexFactor

	for index := range cols {
		if !cols[index].isFlex() {
			continue
		}

		width := flexUnit * (cols[index].flexFactor / flexGCD)

		if leftoverWidth > 0 {
			width++
			leftoverWidth--
		}

		if index == len(cols)-1 {
			width += leftoverWidth
			leftoverWidth = 0
		}

		width = max(width, 1)

		cols[index].width = width

		// Take borders into account for the actual style
		cols[index].style = cols[index].style.Width(width)
	}
}

func (m *Model) recalculateHeight() {
	header := m.renderHeaders()
	headerHeight := 1 // Header always has the top border
	if m.headerVisible {
		headerHeight = lipgloss.Height(header)
	}

	footer := m.renderFooter(lipgloss.Width(header), false)
	var footerHeight int
	if footer != "" {
		footerHeight = lipgloss.Height(footer)
	}

	m.metaHeight = headerHeight + footerHeight
}

func (m *Model) calculatePadding(numRows int) int {
	if m.minimumHeight == 0 {
		return 0
	}

	padding := m.minimumHeight - m.metaHeight - numRows - 1 // additional 1 for bottom border

	if padding == 0 && numRows == 0 {
		// This is an edge case where we want to add 1 additional line of height, i.e.
		// add a border without an empty row. However, this is not possible, so we need
		// to add an extra row which will result in the table being 1 row taller than
		// the requested minimum height.
		return 1
	}

	if padding < 0 {
		// Table is already larger than minimum height, do nothing.
		return 0
	}

	return padding
}
