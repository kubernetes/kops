package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the table.  It does not end in a newline, so that it can be
// composed with other elements more consistently.
func (m Model) View() string {
	// Safety valve for empty tables
	if len(m.columns) == 0 {
		return ""
	}

	body := strings.Builder{}

	rowStrs := make([]string, 0, 1)

	headers := m.renderHeaders()

	startRowIndex, endRowIndex := m.VisibleIndices()

	if m.headerVisible {
		rowStrs = append(rowStrs, headers)
	} else if endRowIndex-startRowIndex > 0 {
		// nolint: gomnd // This is just getting the first newlined substring
		split := strings.SplitN(headers, "\n", 2)
		rowStrs = append(rowStrs, split[0])
	}

	for i := startRowIndex; i <= endRowIndex; i++ {
		rowStrs = append(rowStrs, m.renderRow(i, i == endRowIndex))
	}

	var footer string

	if len(rowStrs) > 0 {
		footer = m.renderFooter(lipgloss.Width(rowStrs[0]), false)
	} else {
		footer = m.renderFooter(lipgloss.Width(headers), true)
	}

	if footer != "" {
		rowStrs = append(rowStrs, footer)
	}

	if len(rowStrs) == 0 {
		return ""
	}

	body.WriteString(lipgloss.JoinVertical(lipgloss.Left, rowStrs...))

	return body.String()
}
