package table

import (
	"fmt"
	"sort"
)

// SortDirection indicates whether a column should sort by ascending or descending.
type SortDirection int

const (
	// SortDirectionAsc indicates the column should be in ascending order.
	SortDirectionAsc SortDirection = iota

	// SortDirectionDesc indicates the column should be in descending order.
	SortDirectionDesc
)

// SortColumn describes which column should be sorted and how.
type SortColumn struct {
	ColumnKey string
	Direction SortDirection
}

// SortByAsc sets the main sorting column to the given key, in ascending order.
// If a previous sort was used, it is replaced by the given column each time
// this function is called.  Values are sorted as numbers if possible, or just
// as simple string comparisons if not numbers.
func (m Model) SortByAsc(columnKey string) Model {
	m.sortOrder = []SortColumn{
		{
			ColumnKey: columnKey,
			Direction: SortDirectionAsc,
		},
	}

	m.visibleRowCacheUpdated = false

	return m
}

// SortByDesc sets the main sorting column to the given key, in descending order.
// If a previous sort was used, it is replaced by the given column each time
// this function is called.  Values are sorted as numbers if possible, or just
// as simple string comparisons if not numbers.
func (m Model) SortByDesc(columnKey string) Model {
	m.sortOrder = []SortColumn{
		{
			ColumnKey: columnKey,
			Direction: SortDirectionDesc,
		},
	}

	m.visibleRowCacheUpdated = false

	return m
}

// ThenSortByAsc provides a secondary sort after the first, in ascending order.
// Can be chained multiple times, applying to smaller subgroups each time.
func (m Model) ThenSortByAsc(columnKey string) Model {
	m.sortOrder = append([]SortColumn{
		{
			ColumnKey: columnKey,
			Direction: SortDirectionAsc,
		},
	}, m.sortOrder...)

	m.visibleRowCacheUpdated = false

	return m
}

// ThenSortByDesc provides a secondary sort after the first, in descending order.
// Can be chained multiple times, applying to smaller subgroups each time.
func (m Model) ThenSortByDesc(columnKey string) Model {
	m.sortOrder = append([]SortColumn{
		{
			ColumnKey: columnKey,
			Direction: SortDirectionDesc,
		},
	}, m.sortOrder...)

	m.visibleRowCacheUpdated = false

	return m
}

type sortableTable struct {
	rows     []Row
	byColumn SortColumn
}

func (s *sortableTable) Len() int {
	return len(s.rows)
}

func (s *sortableTable) Swap(i, j int) {
	old := s.rows[i]
	s.rows[i] = s.rows[j]
	s.rows[j] = old
}

func (s *sortableTable) extractString(i int, column string) string {
	iData, exists := s.rows[i].Data[column]

	if !exists {
		return ""
	}

	switch iData := iData.(type) {
	case StyledCell:
		return fmt.Sprintf("%v", iData.Data)

	case string:
		return iData

	default:
		return fmt.Sprintf("%v", iData)
	}
}

func (s *sortableTable) extractNumber(i int, column string) (float64, bool) {
	iData, exists := s.rows[i].Data[column]

	if !exists {
		return 0, false
	}

	return asNumber(iData)
}

func (s *sortableTable) Less(first, second int) bool {
	firstNum, firstNumIsValid := s.extractNumber(first, s.byColumn.ColumnKey)
	secondNum, secondNumIsValid := s.extractNumber(second, s.byColumn.ColumnKey)

	if firstNumIsValid && secondNumIsValid {
		if s.byColumn.Direction == SortDirectionAsc {
			return firstNum < secondNum
		}

		return firstNum > secondNum
	}

	firstVal := s.extractString(first, s.byColumn.ColumnKey)
	secondVal := s.extractString(second, s.byColumn.ColumnKey)

	if s.byColumn.Direction == SortDirectionAsc {
		return firstVal < secondVal
	}

	return firstVal > secondVal
}

func getSortedRows(sortOrder []SortColumn, rows []Row) []Row {
	var sortedRows []Row
	if len(sortOrder) == 0 {
		sortedRows = rows

		return sortedRows
	}

	sortedRows = make([]Row, len(rows))
	copy(sortedRows, rows)

	for _, byColumn := range sortOrder {
		sorted := &sortableTable{
			rows:     sortedRows,
			byColumn: byColumn,
		}

		sort.Stable(sorted)

		sortedRows = sorted.rows
	}

	return sortedRows
}
