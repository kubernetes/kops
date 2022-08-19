package table

import "github.com/charmbracelet/lipgloss"

const columnKeyOverflowRight = "___overflow_r___"
const columnKeyOverflowLeft = "___overflow_l__"

func genOverflowStyle(base lipgloss.Style, width int) lipgloss.Style {
	style := lipgloss.NewStyle().Width(width).Align(lipgloss.Right)

	style.Inherit(base)

	return style
}

func genOverflowColumnRight(width int) Column {
	return NewColumn(columnKeyOverflowRight, ">", width)
}

func genOverflowColumnLeft(width int) Column {
	return NewColumn(columnKeyOverflowLeft, "<", width)
}
