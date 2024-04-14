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
		str := fmt.Sprintf("%d/%d", m.CurrentPage(), m.MaxPages())
		if m.filtered && m.filterTextInput.Focused() {
			// Need to apply inline style here in case of filter input cursor, because
			// the input cursor resets the style after rendering.  Note that Inline(true)
			// creates a copy, so it's safe to use here without mutating the underlying
			// base style.
			str = m.baseStyle.Inline(true).Render(str)
		}
		sections = append(sections, str)
	}

	footerText := strings.Join(sections, " ")

	return styleFooter.Render(footerText)
}
