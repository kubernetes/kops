package table

import "github.com/charmbracelet/lipgloss"

const columnKeyOverflowRight = "___overflow_r___"
const columnKeyOverflowLeft = "___overflow_l__"

func genOverflowStyle(base lipgloss.Style, width int) lipgloss.Style {
	return base.Width(width).Align(lipgloss.Right)
}

func genOverflowColumnRight(width int) Column {
	return NewColumn(columnKeyOverflowRight, ">", width)
}

func genOverflowColumnLeft(width int) Column {
	return NewColumn(columnKeyOverflowLeft, "<", width)
}
