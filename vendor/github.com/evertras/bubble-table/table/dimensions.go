package table

func (m *Model) recalculateWidth() {
	if m.targetTotalWidth != 0 {
		m.totalWidth = m.targetTotalWidth
	} else {
		total := 0

		for _, column := range m.columns {
			total += column.width
		}

		m.totalWidth = total + len(m.columns) + 1
	}

	updateColumnWidths(m.columns, m.targetTotalWidth)

	m.recalculateLastHorizontalColumn()
}

// Updates column width in-place.  This could be optimized but should be called
// very rarely so we prioritize simplicity over performance here.
func updateColumnWidths(cols []Column, totalWidth int) {
	totalFlexWidth := totalWidth - len(cols) - 1
	totalFlexFactor := 0
	flexGCD := 0

	for index, col := range cols {
		if !col.isFlex() {
			totalFlexWidth -= col.width
			cols[index].style = col.style.Width(col.width)
		} else {
			totalFlexFactor += col.flexFactor
			flexGCD = gcd(flexGCD, col.flexFactor)
		}
	}

	if totalFlexFactor == 0 {
		return
	}

	// We use the GCD here because otherwise very large values won't divide
	// nicely as ints
	totalFlexFactor /= flexGCD

	flexUnit := totalFlexWidth / totalFlexFactor
	leftoverWidth := totalFlexWidth % totalFlexFactor

	for index := range cols {
		if !cols[index].isFlex() {
			continue
		}

		width := flexUnit * (cols[index].flexFactor / flexGCD)

		if leftoverWidth > 0 {
			width++
			leftoverWidth--
		}

		if index == len(cols)-1 {
			width += leftoverWidth
			leftoverWidth = 0
		}

		width = max(width, 1)

		cols[index].width = width

		// Take borders into account for the actual style
		cols[index].style = cols[index].style.Width(width)
	}
}
