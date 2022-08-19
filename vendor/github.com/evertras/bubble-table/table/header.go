package table

import "github.com/charmbracelet/lipgloss"

// This is long and could use some refactoring in the future, but unsure of how
// to pick it apart right now.
// nolint: funlen, cyclop
func (m Model) renderHeaders() string {
	headerStrings := []string{}

	totalRenderedWidth := 0

	headerStyles := m.styleHeaders()

	renderHeader := func(column Column, borderStyle lipgloss.Style) string {
		borderStyle = borderStyle.Inherit(column.style).Inherit(m.baseStyle)

		headerSection := limitStr(column.title, column.width)

		return borderStyle.Render(headerSection)
	}

	for columnIndex, column := range m.columns {
		var borderStyle lipgloss.Style

		if m.horizontalScrollOffsetCol > 0 && columnIndex == m.horizontalScrollFreezeColumnsCount {
			if columnIndex == 0 {
				borderStyle = headerStyles.left.Copy()
			} else {
				borderStyle = headerStyles.inner.Copy()
			}

			rendered := renderHeader(genOverflowColumnLeft(1), borderStyle)

			totalRenderedWidth += lipgloss.Width(rendered)

			headerStrings = append(headerStrings, rendered)
		}

		if columnIndex >= m.horizontalScrollFreezeColumnsCount &&
			columnIndex < m.horizontalScrollOffsetCol+m.horizontalScrollFreezeColumnsCount {
			continue
		}

		if len(headerStrings) == 0 {
			borderStyle = headerStyles.left.Copy()
		} else if columnIndex < len(m.columns)-1 {
			borderStyle = headerStyles.inner.Copy()
		} else {
			borderStyle = headerStyles.right.Copy()
		}

		rendered := renderHeader(column, borderStyle)

		if m.maxTotalWidth != 0 {
			renderedWidth := lipgloss.Width(rendered)

			const (
				borderAdjustment = 1
				overflowColWidth = 2
			)

			targetWidth := m.maxTotalWidth - overflowColWidth

			if columnIndex == len(m.columns)-1 {
				// If this is the last header, we don't need to account for the
				// overflow arrow column
				targetWidth = m.maxTotalWidth
			}

			if totalRenderedWidth+renderedWidth > targetWidth {
				overflowWidth := m.maxTotalWidth - totalRenderedWidth - borderAdjustment
				overflowStyle := genOverflowStyle(headerStyles.right, overflowWidth)
				overflowColumn := genOverflowColumnRight(overflowWidth)

				overflowStr := renderHeader(overflowColumn, overflowStyle)

				headerStrings = append(headerStrings, overflowStr)

				break
			}

			totalRenderedWidth += renderedWidth
		}

		headerStrings = append(headerStrings, rendered)
	}

	headerBlock := lipgloss.JoinHorizontal(lipgloss.Bottom, headerStrings...)

	return headerBlock
}
