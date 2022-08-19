package table

import (
	"fmt"
	"strings"
)

func (m Model) getFilteredRows(rows []Row) []Row {
	if !m.filtered || m.filterTextInput.Value() == "" {
		return rows
	}

	filteredRows := make([]Row, 0)

	for _, row := range rows {
		if isRowMatched(m.columns, row, m.filterTextInput.Value()) {
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

	for _, column := range columns {
		if !column.filterable {
			continue
		}

		checkedAny = true

		data, ok := row.Data[column.key]

		if !ok {
			continue
		}

		switch dataV := data.(type) {
		case string:
			if strings.Contains(strings.ToLower(dataV), strings.ToLower(filter)) {
				return true
			}

		case fmt.Stringer:
			if strings.Contains(strings.ToLower(dataV.String()), strings.ToLower(filter)) {
				return true
			}
		}
	}

	return !checkedAny
}
