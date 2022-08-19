package table

// PageSize returns the current page size for the table, or 0 if there is no
// pagination enabled.
func (m *Model) PageSize() int {
	return m.pageSize
}

// CurrentPage returns the current page that the table is on, starting from an
// index of 1.
func (m *Model) CurrentPage() int {
	return m.currentPage + 1
}

// MaxPages returns the maximum number of pages that are visible.
func (m *Model) MaxPages() int {
	if m.pageSize == 0 || len(m.GetVisibleRows()) == 0 {
		return 1
	}

	return (len(m.GetVisibleRows())-1)/m.pageSize + 1
}

// TotalRows returns the current total row count of the table.  If the table is
// paginated, this is the total number of rows across all pages.
func (m *Model) TotalRows() int {
	return len(m.GetVisibleRows())
}

// VisibleIndices returns the current visible rows by their 0 based index.
// Useful for custom pagination footers.
func (m *Model) VisibleIndices() (start, end int) {
	totalRows := len(m.GetVisibleRows())

	if m.pageSize == 0 {
		start = 0
		end = totalRows - 1

		return start, end
	}

	start = m.pageSize * m.currentPage
	end = start + m.pageSize - 1

	if end >= totalRows {
		end = totalRows - 1
	}

	return start, end
}

func (m *Model) pageDown() {
	if m.pageSize == 0 || len(m.GetVisibleRows()) <= m.pageSize {
		return
	}

	m.currentPage++

	maxPageIndex := m.MaxPages() - 1

	if m.currentPage > maxPageIndex {
		if m.paginationWrapping {
			m.currentPage = 0
		} else {
			m.currentPage = maxPageIndex
		}
	}

	m.rowCursorIndex = m.currentPage * m.pageSize
}

func (m *Model) pageUp() {
	if m.pageSize == 0 || len(m.GetVisibleRows()) <= m.pageSize {
		return
	}

	m.currentPage--

	maxPageIndex := m.MaxPages() - 1

	if m.currentPage < 0 {
		if m.paginationWrapping {
			m.currentPage = maxPageIndex
		} else {
			m.currentPage = 0
		}
	}

	m.rowCursorIndex = m.currentPage * m.pageSize
}

func (m *Model) pageFirst() {
	m.currentPage = 0
	m.rowCursorIndex = 0
}

func (m *Model) pageLast() {
	m.currentPage = m.MaxPages() - 1
	m.rowCursorIndex = m.currentPage * m.pageSize
}

func (m *Model) expectedPageForRowIndex(rowIndex int) int {
	if m.pageSize == 0 {
		return 0
	}

	expectedPage := rowIndex / m.pageSize

	return expectedPage
}
