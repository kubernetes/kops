/*
Package table contains a Bubble Tea component for an interactive and customizable
table.

The simplest useful table can be created with table.New(...).WithRows(...).  Row
data should map to the column keys, as shown below.  Note that extra data will
simply not be shown, while missing data will be safely blank in the row's cell.

	const (
		// This is not necessary, but recommended to avoid typos
		columnKeyName  = "name"
		columnKeyCount = "count"
	)

	// Define the columns and how they appear
	columns := []table.Column{
		table.NewColumn(columnKeyName, "Name", 10),
		table.NewColumn(columnKeyCount, "Count", 6),
	}

	// Define the data that will be in the table, mapping to the column keys
	rows := []table.Row{
		table.NewRow(table.RowData{
			columnKeyName:  "Cheeseburger",
			columnKeyCount: 3,
		}),
		table.NewRow(table.RowData{
			columnKeyName:  "Fries",
			columnKeyCount: 2,
		}),
	}

	// Create the table
	tbl := table.New(columns).WithRows(rows)

	// Use it like any Bubble Tea component in your view
	tbl.View()
*/
package table
