/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tables

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"text/tabwriter"

	"k8s.io/klog"

	"k8s.io/kops/util/pkg/reflectutils"
)

// Table renders tables to stdout
type Table struct {
	columns map[string]*TableColumn
}

type TableColumn struct {
	Name   string
	Getter reflect.Value
}

func (c *TableColumn) getFromValue(v reflect.Value) string {
	var args []reflect.Value
	args = append(args, v)
	fvs := c.Getter.Call(args)
	fv := fvs[0]

	return reflectutils.ValueAsString(fv)
}

type getterFunction func(interface{}) string

// AddColumn registers an available column for formatting
func (t *Table) AddColumn(name string, getter interface{}) {
	getterVal := reflect.ValueOf(getter)

	column := &TableColumn{
		Name:   name,
		Getter: getterVal,
	}
	if t.columns == nil {
		t.columns = make(map[string]*TableColumn)
	}
	t.columns[name] = column
}

type funcSorter struct {
	len  int
	less func(int, int) bool
	swap func(int, int)
}

func (f *funcSorter) Len() int {
	return f.len
}
func (f *funcSorter) Less(i, j int) bool {
	return f.less(i, j)
}
func (f *funcSorter) Swap(i, j int) {
	f.swap(i, j)
}

func SortByFunction(len int, swap func(int, int), less func(int, int) bool) {
	sort.Sort(&funcSorter{len, less, swap})
}

func (t *Table) findColumns(columnNames ...string) ([]*TableColumn, error) {
	columns := make([]*TableColumn, len(columnNames))
	for i, columnName := range columnNames {
		c := t.columns[columnName]
		if c == nil {
			return nil, fmt.Errorf("column not found: %v", columnName)
		}
		columns[i] = c
	}
	return columns, nil
}

// Render writes the items in a table, to out
func (t *Table) Render(items interface{}, out io.Writer, columnNames ...string) error {
	itemsValue := reflect.ValueOf(items)
	if itemsValue.Kind() != reflect.Slice {
		klog.Fatal("unexpected kind for items: ", itemsValue.Kind())
	}

	columns, err := t.findColumns(columnNames...)
	if err != nil {
		return err
	}

	n := itemsValue.Len()

	rows := make([][]string, n)
	for i := 0; i < n; i++ {
		row := make([]string, len(columns))
		item := itemsValue.Index(i)
		for j, column := range columns {
			row[j] = column.getFromValue(item)
		}
		rows[i] = row
	}

	SortByFunction(n, func(i, j int) {
		row := rows[i]
		rows[i] = rows[j]
		rows[j] = row
	}, func(i, j int) bool {
		l := rows[i]
		r := rows[j]

		for k := 0; k < len(columns); k++ {
			lV := l[k]
			rV := r[k]

			if lV != rV {
				return lV < rV
			}
		}
		return false
	})

	var b bytes.Buffer
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(out, 0, 8, 1, '\t', tabwriter.StripEscape)

	writeHeader := true
	if writeHeader {
		for i, c := range columns {
			if i != 0 {
				b.WriteByte('\t')
			}
			b.WriteByte(tabwriter.Escape)
			b.WriteString(c.Name)
			b.WriteByte(tabwriter.Escape)
		}
		b.WriteByte('\n')

		_, err := w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		b.Reset()
	}

	for _, row := range rows {
		for i, col := range row {
			if i != 0 {
				b.WriteByte('\t')
			}

			b.WriteByte(tabwriter.Escape)
			b.WriteString(col)
			b.WriteByte(tabwriter.Escape)
		}
		b.WriteByte('\n')

		_, err := w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		b.Reset()
	}
	w.Flush()

	return nil
}
