package table

import (
	"fmt"
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// RowData is a map of string column keys to interface{} data.  Data with a key
// that matches a column key will be displayed.  Data with a key that does not
// match a column key will not be displayed, but will remain attached to the Row.
// This can be useful for attaching hidden metadata for future reference when
// retrieving rows.
type RowData map[string]interface{}

// Row represents a row in the table with some data keyed to the table columns>
// Can have a style applied to it such as color/bold.  Create using NewRow().
type Row struct {
	Style lipgloss.Style
	Data  RowData

	selected bool

	// id is an internal unique ID to match rows after they're copied
	id uint32
}

var lastRowID uint32 = 1

// NewRow creates a new row and copies the given row data.
func NewRow(data RowData) Row {
	row := Row{
		Data: make(map[string]interface{}),
		id:   lastRowID,
	}

	atomic.AddUint32(&lastRowID, 1)

	for key, val := range data {
		// Doesn't deep copy val, but close enough for now...
		row.Data[key] = val
	}

	return row
}

// WithStyle uses the given style for the text in the row.
func (r Row) WithStyle(style lipgloss.Style) Row {
	r.Style = style.Copy()

	return r
}

//nolint:nestif,cyclop // This has many ifs, but they're short
func (m Model) renderRowColumnData(row Row, column Column, rowStyle lipgloss.Style, borderStyle lipgloss.Style) string {
	cellStyle := rowStyle.Copy().Inherit(column.style).Inherit(m.baseStyle)

	var str string

	if column.key == columnKeySelect {
		if row.selected {
			str = m.selectedText
		} else {
			str = m.unselectedText
		}
	} else if column.key == columnKeyOverflowRight {
		cellStyle = cellStyle.Align(lipgloss.Right)
		str = ">"
	} else if column.key == columnKeyOverflowLeft {
		str = "<"
	} else {
		fmtString := "%v"

		var data interface{}

		if entry, exists := row.Data[column.key]; exists {
			data = entry

			if column.fmtString != "" {
				fmtString = column.fmtString
			}
		} else if m.missingDataIndicator != nil {
			data = m.missingDataIndicator
		} else {
			data = ""
		}

		switch entry := data.(type) {
		case StyledCell:
			str = fmt.Sprintf(fmtString, entry.Data)
			cellStyle = entry.Style.Copy().Inherit(cellStyle)
		default:
			str = fmt.Sprintf(fmtString, entry)
		}
	}

	if m.multiline {
		str = wordwrap.String(str, column.width)
		cellStyle = cellStyle.Align(lipgloss.Top)
	} else {
		str = limitStr(str, column.width)
	}

	cellStyle = cellStyle.Inherit(borderStyle)
	cellStr := cellStyle.Render(str)

	return cellStr
}

func (m Model) renderRow(rowIndex int, last bool) string {
	row := m.GetVisibleRows()[rowIndex]
	highlighted := rowIndex == m.rowCursorIndex

	rowStyle := row.Style.Copy()

	if m.rowStyleFunc != nil {
		styleResult := m.rowStyleFunc(RowStyleFuncInput{
			Index:         rowIndex,
			Row:           row,
			IsHighlighted: m.focused && highlighted,
		})

		rowStyle = rowStyle.Inherit(styleResult)
	} else if m.focused && highlighted {
		rowStyle = rowStyle.Inherit(m.highlightStyle)
	}

	return m.renderRowData(row, rowStyle, last)
}

func (m Model) renderBlankRow(last bool) string {
	return m.renderRowData(NewRow(nil), lipgloss.NewStyle(), last)
}

// This is long and could use some refactoring in the future, but not quite sure
// how to pick it apart yet.
//
//nolint:funlen, cyclop, gocognit
func (m Model) renderRowData(row Row, rowStyle lipgloss.Style, last bool) string {
	numColumns := len(m.columns)

	columnStrings := []string{}
	totalRenderedWidth := 0

	stylesInner, stylesLast := m.styleRows()

	maxCellHeight := 1
	if m.multiline {
		for _, column := range m.columns {
			cellStr := m.renderRowColumnData(row, column, rowStyle, lipgloss.NewStyle())
			maxCellHeight = max(maxCellHeight, lipgloss.Height(cellStr))
		}
	}

	for columnIndex, column := range m.columns {
		var borderStyle lipgloss.Style
		var rowStyles borderStyleRow

		if !last {
			rowStyles = stylesInner
		} else {
			rowStyles = stylesLast
		}
		rowStyle = rowStyle.Copy().Height(maxCellHeight)

		if m.horizontalScrollOffsetCol > 0 && columnIndex == m.horizontalScrollFreezeColumnsCount {
			var borderStyle lipgloss.Style

			if columnIndex == 0 {
				borderStyle = rowStyles.left.Copy()
			} else {
				borderStyle = rowStyles.inner.Copy()
			}

			rendered := m.renderRowColumnData(row, genOverflowColumnLeft(1), rowStyle, borderStyle)

			totalRenderedWidth += lipgloss.Width(rendered)

			columnStrings = append(columnStrings, rendered)
		}

		if columnIndex >= m.horizontalScrollFreezeColumnsCount &&
			columnIndex < m.horizontalScrollOffsetCol+m.horizontalScrollFreezeColumnsCount {
			continue
		}

		if len(columnStrings) == 0 {
			borderStyle = rowStyles.left
		} else if columnIndex < numColumns-1 {
			borderStyle = rowStyles.inner
		} else {
			borderStyle = rowStyles.right
		}

		cellStr := m.renderRowColumnData(row, column, rowStyle, borderStyle)

		if m.maxTotalWidth != 0 {
			renderedWidth := lipgloss.Width(cellStr)

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
				overflowStyle := genOverflowStyle(rowStyles.right, overflowWidth)
				overflowColumn := genOverflowColumnRight(overflowWidth)
				overflowStr := m.renderRowColumnData(row, overflowColumn, rowStyle, overflowStyle)

				columnStrings = append(columnStrings, overflowStr)

				break
			}

			totalRenderedWidth += renderedWidth
		}

		columnStrings = append(columnStrings, cellStr)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, columnStrings...)
}

// Selected returns a copy of the row that's set to be selected or deselected.
// The old row is not changed in-place.
func (r Row) Selected(selected bool) Row {
	r.selected = selected

	return r
}
