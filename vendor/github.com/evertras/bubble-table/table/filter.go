package table

import (
	"fmt"
	"strings"
)

func (m Model) getFilteredRows(rows []Row) []Row {
	filterInputValue := m.filterTextInput.Value()
	if !m.filtered || filterInputValue == "" {
		return rows
	}

	filteredRows := make([]Row, 0)

	for _, row := range rows {
		if isRowMatched(m.columns, row, filterInputValue) {
			filteredRows = append(filteredRows, row)
		}
	}

	return filteredRows
}

func isRowMatched(columns []Column, row Row, filter string) bool {
	if filter == "" {
		return true
	}

	checkedAny := false

	filterLower := strings.ToLower(filter)

	for _, column := range columns {
		if !column.filterable {
			continue
		}

		checkedAny = true

		data, ok := row.Data[column.key]

		if !ok {
			continue
		}

		// Extract internal StyledCell data
		switch dataV := data.(type) {
		case StyledCell:
			data = dataV.Data
		}

		var target string
		switch dataV := data.(type) {
		case string:
			target = dataV

		case fmt.Stringer:
			target = dataV.String()

		default:
			target = fmt.Sprintf("%v", data)
		}

		if strings.Contains(strings.ToLower(target), filterLower) {
			return true
		}
	}

	return !checkedAny
}
