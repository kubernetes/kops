package table

import (
	"fmt"
	"strings"
)

func (m Model) hasFooter() bool {
	return m.footerVisible && (m.staticFooter != "" || m.pageSize != 0 || m.filtered)
}

func (m Model) renderFooter(width int, includeTop bool) string {
	if !m.hasFooter() {
		return ""
	}

	const borderAdjustment = 2

	styleFooter := m.baseStyle.Copy().Inherit(m.border.styleFooter).Width(width - borderAdjustment)

	if includeTop {
		styleFooter.BorderTop(true)
	}

	if m.staticFooter != "" {
		return styleFooter.Render(m.staticFooter)
	}

	sections := []string{}

	if m.filtered && (m.filterTextInput.Focused() || m.filterTextInput.Value() != "") {
		sections = append(sections, m.filterTextInput.View())
	}

	// paged feature enabled
	if m.pageSize != 0 {
		sections = append(sections, fmt.Sprintf("%d/%d", m.CurrentPage(), m.MaxPages()))
	}

	footerText := strings.Join(sections, " ")

	return styleFooter.Render(footerText)
}
