package main

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kops/upup/pkg/fi"
	"reflect"
	"text/tabwriter"
)

// Table renders tables to stdout
type Table struct {
	columns []string
	getters []reflect.Value
}

// AddColumn registers an available column for formatting
func (t *Table) AddColumn(name string, getter interface{}) {
	getterVal := reflect.ValueOf(getter)

	t.columns = append(t.columns, name)
	t.getters = append(t.getters, getterVal)
}

// Render writes the items in a table, to out
func (t *Table) Render(items interface{}, out io.Writer) error {
	itemsValue := reflect.ValueOf(items)
	if itemsValue.Kind() != reflect.Slice {
		glog.Fatal("unexpected kind for items: ", itemsValue.Kind())
	}

	length := itemsValue.Len()

	var b bytes.Buffer
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(out, 0, 8, 0, '\t', tabwriter.StripEscape)

	writeHeader := true
	if writeHeader {
		for i, c := range t.columns {
			if i != 0 {
				b.WriteByte('\t')
			}
			b.WriteByte(tabwriter.Escape)
			b.WriteString(c)
			b.WriteByte(tabwriter.Escape)
		}
		b.WriteByte('\n')

		_, err := w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		b.Reset()
	}

	for i := 0; i < length; i++ {
		item := itemsValue.Index(i)

		for j := range t.columns {
			if j != 0 {
				b.WriteByte('\t')
			}

			getter := t.getters[j]
			var args []reflect.Value
			args = append(args, item)
			fvs := getter.Call(args)
			fv := fvs[0]

			s := fi.ValueAsString(fv)

			b.WriteByte(tabwriter.Escape)
			b.WriteString(s)
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
