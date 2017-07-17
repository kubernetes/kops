/*
Copyright 2016 The Kubernetes Authors.

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

package fi

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"

	"text/tabwriter"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// DryRunTarget is a special Target that does not execute anything, but instead tracks all changes.
// By running against a DryRunTarget, a list of changes that would be made can be easily collected,
// without any special support from the Tasks.
type DryRunTarget struct {
	mutex sync.Mutex

	changes   []*render
	deletions []Deletion

	// The destination to which the final report will be printed on Finish()
	out io.Writer

	// inventory of cluster
	inventory *assets.Inventory
}

type render struct {
	a       Task
	aIsNil  bool
	e       Task
	changes Task
}

// ByTaskKey sorts []*render by TaskKey (type/name)
type ByTaskKey []*render

func (a ByTaskKey) Len() int      { return len(a) }
func (a ByTaskKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTaskKey) Less(i, j int) bool {
	return buildTaskKey(a[i].e) < buildTaskKey(a[j].e)
}

// DeletionByTaskName sorts []Deletion by TaskName
type DeletionByTaskName []Deletion

func (a DeletionByTaskName) Len() int      { return len(a) }
func (a DeletionByTaskName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DeletionByTaskName) Less(i, j int) bool {
	return a[i].TaskName() < a[i].TaskName()
}

var _ Target = &DryRunTarget{}

func NewDryRunTarget(out io.Writer, inventory *assets.Inventory) *DryRunTarget {
	t := &DryRunTarget{}
	t.out = out
	t.inventory = inventory
	return t
}

func (t *DryRunTarget) ProcessDeletions() bool {
	// We display deletions
	return true
}

func (t *DryRunTarget) Render(a, e, changes Task) error {
	valA := reflect.ValueOf(a)
	aIsNil := valA.IsNil()

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.changes = append(t.changes, &render{
		a:       a,
		aIsNil:  aIsNil,
		e:       e,
		changes: changes,
	})
	return nil
}

func (t *DryRunTarget) Delete(deletion Deletion) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.deletions = append(t.deletions, deletion)

	return nil
}

func idForTask(taskMap map[string]Task, t Task) string {
	for k, v := range taskMap {
		if v == t {
			// Skip task type, if present (taskType/taskName)
			firstSlash := strings.Index(k, "/")
			if firstSlash != -1 {
				k = k[firstSlash+1:]
			}
			return k
		}
	}
	glog.Fatalf("unknown task: %v", t)
	return "?"
}

func (t *DryRunTarget) PrintReport(taskMap map[string]Task, out io.Writer) error {
	b := &bytes.Buffer{}

	if len(t.changes) != 0 {
		var creates []*render
		var updates []*render

		for _, r := range t.changes {
			if r.aIsNil {
				creates = append(creates, r)
			} else {
				updates = append(updates, r)
			}
		}

		// Give everything a consistent ordering
		sort.Sort(ByTaskKey(creates))
		sort.Sort(ByTaskKey(updates))

		if len(creates) != 0 {
			fmt.Fprintf(b, "Will create resources:\n")
			for _, r := range creates {
				taskName := getTaskName(r.changes)
				fmt.Fprintf(b, "  %s/%s\n", taskName, idForTask(taskMap, r.e))

				changes := reflect.ValueOf(r.changes)
				if changes.Kind() == reflect.Ptr && !changes.IsNil() {
					changes = changes.Elem()
				}

				if changes.Kind() == reflect.Struct {
					for i := 0; i < changes.NumField(); i++ {

						field := changes.Field(i)

						fieldName := changes.Type().Field(i).Name
						if changes.Type().Field(i).PkgPath != "" {
							// Not exported
							continue
						}

						fieldValue := ValueAsString(field)

						shouldPrint := true
						if fieldName == "Name" {
							// The field name is already printed above, no need to repeat it.
							shouldPrint = false
						}
						if fieldValue == "<nil>" || fieldValue == "<resource>" {
							// Uninformative
							shouldPrint = false
						}
						if fieldValue == "id:<nil>" {
							// Uninformative, but we can often print the name instead
							name := ""
							if field.CanInterface() {
								hasName, ok := field.Interface().(HasName)
								if ok {
									name = StringValue(hasName.GetName())
								}
							}
							if name != "" {
								fieldValue = "name:" + name
							} else {
								shouldPrint = false
							}
						}
						if shouldPrint {
							fmt.Fprintf(b, "  \t%-20s\t%s\n", fieldName, fieldValue)
						}
					}
				}

				fmt.Fprintf(b, "\n")
			}
		}

		if len(updates) != 0 {
			fmt.Fprintf(b, "Will modify resources:\n")
			// We can't use our reflection helpers here - we want corresponding values from a,e,c
			for _, r := range updates {
				type change struct {
					FieldName   string
					Description string
				}
				var changeList []change

				valC := reflect.ValueOf(r.changes)
				valA := reflect.ValueOf(r.a)
				valE := reflect.ValueOf(r.e)
				if valC.Kind() == reflect.Ptr && !valC.IsNil() {
					valC = valC.Elem()
				}
				if valA.Kind() == reflect.Ptr && !valA.IsNil() {
					valA = valA.Elem()
				}
				if valE.Kind() == reflect.Ptr && !valE.IsNil() {
					valE = valE.Elem()
				}
				if valC.Kind() == reflect.Struct {
					for i := 0; i < valC.NumField(); i++ {
						if valC.Type().Field(i).PkgPath != "" {
							// Not exported
							continue
						}

						fieldValC := valC.Field(i)

						changed := true
						switch fieldValC.Kind() {
						case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
							changed = !fieldValC.IsNil()

						case reflect.String:
							changed = fieldValC.Interface().(string) != ""
						}
						if !changed {
							continue
						}

						if fieldValC.Kind() == reflect.String && fieldValC.Interface().(string) == "" {
							// No change
							continue
						}

						fieldValE := valE.Field(i)

						description := ""
						ignored := false
						if fieldValE.CanInterface() {
							fieldValA := valA.Field(i)

							switch fieldValE.Interface().(type) {
							//case SimpleUnit:
							//	ignored = true
							case Resource, ResourceHolder:
								resA, okA := tryResourceAsString(fieldValA)
								resE, okE := tryResourceAsString(fieldValE)
								if okA && okE {
									description = diff.FormatDiff(resA, resE)
								}
							}

							if !ignored && description == "" {
								description = fmt.Sprintf(" %v -> %v", ValueAsString(fieldValA), ValueAsString(fieldValE))
							}
						}
						if ignored {
							continue
						}
						changeList = append(changeList, change{FieldName: valC.Type().Field(i).Name, Description: description})
					}
				} else {
					return fmt.Errorf("unhandled change type: %v", valC.Type())
				}

				if len(changeList) == 0 {
					continue
				}

				taskName := getTaskName(r.changes)
				fmt.Fprintf(b, "  %s/%s\n", taskName, idForTask(taskMap, r.e))
				for _, change := range changeList {
					lines := strings.Split(change.Description, "\n")
					if len(lines) == 1 {
						fmt.Fprintf(b, "  \t%-20s\t%s\n", change.FieldName, change.Description)
					} else {
						fmt.Fprintf(b, "  \t%-20s\n", change.FieldName)
						for _, line := range lines {
							fmt.Fprintf(b, "  \t%-20s\t%s\n", "", line)
						}
					}
				}
				fmt.Fprintf(b, "\n")
			}
		}
	}

	if t.inventory != nil {
		fmt.Fprintf(out, "Cluster Inventory\n\n")
		fmt.Fprintf(out, "Files\n\n")

		asset := "Asset"
		name := "Name"
		sha := "SHA"

		table := &Table{}
		table.AddColumn(asset, func(i *assets.ExecutableFileAsset) string {
			return i.Location
		})
		table.AddColumn(name, func(i *assets.ExecutableFileAsset) string {
			return i.Name
		})
		if err := table.Render(t.inventory.ExecutableFileAsset, out, name, asset); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nFile SHAs\n\n")

		table = &Table{}
		table.AddColumn(name, func(i *assets.ExecutableFileAsset) string {
			return i.Name
		})
		table.AddColumn(sha, func(i *assets.ExecutableFileAsset) string {
			return i.SHA
		})
		if err := table.Render(t.inventory.ExecutableFileAsset, out, name, sha); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nCompressed Files\n\n")

		table = &Table{}
		table.AddColumn(asset, func(i *assets.CompressedFileAsset) string {
			return i.Location
		})
		table.AddColumn(name, func(i *assets.CompressedFileAsset) string {
			return i.Name
		})
		if err := table.Render(t.inventory.CompressedFileAssets, out, name, asset); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nCompressed File SHAs\n\n")

		table = &Table{}
		table.AddColumn(name, func(i *assets.CompressedFileAsset) string {
			return i.Name
		})
		table.AddColumn(sha, func(i *assets.CompressedFileAsset) string {
			return i.SHA
		})
		if err := table.Render(t.inventory.CompressedFileAssets, out, name, sha); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nContainers\n\n")
		table = &Table{}
		table.AddColumn(asset, func(i *assets.ContainerAsset) string {
			if i.String != "" {
				return i.String
			} else if i.Location != "" {
				return i.Location
			}

			glog.Errorf("unable to print container asset %q", i)

			return "asset name not set correctly"
		})

		table.AddColumn(name, func(i *assets.ContainerAsset) string {
			return i.Name
		})
		if err := table.Render(t.inventory.ContainerAssets, out, name, asset); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\nHosts\n\n")
		table = &Table{}
		table.AddColumn("Image", func(i *assets.HostAsset) string {
			return i.Name
		})

		table.AddColumn("Instance Group", func(i *assets.HostAsset) string {
			return i.InstanceGroup
		})
		if err := table.Render(t.inventory.HostAssets, out, "Instance Group", "Image"); err != nil {
			return err
		}

		fmt.Fprintf(out, "\n\n")

		if len(t.deletions) != 0 {
			// Give everything a consistent ordering
			sort.Sort(DeletionByTaskName(t.deletions))

			fmt.Fprintf(b, "Will delete items:\n")
			for _, d := range t.deletions {
				fmt.Fprintf(b, "  %-20s %s\n", d.TaskName(), d.Item())
			}
		}
	}

	_, err := out.Write(b.Bytes())
	return err
}

func tryResourceAsString(v reflect.Value) (string, bool) {
	if !v.CanInterface() {
		return "", false
	}

	intf := v.Interface()
	if res, ok := intf.(Resource); ok {
		s, err := ResourceAsString(res)
		if err != nil {
			glog.Warningf("error converting to resource: %v", err)
			return "", false
		}
		return s, true
	}
	if res, ok := intf.(*ResourceHolder); ok {
		s, err := res.AsString()
		if err != nil {
			glog.Warningf("error converting to resource: %v", err)
			return "", false
		}
		return s, true
	}
	return "", false
}

func getTaskName(t Task) string {
	s := fmt.Sprintf("%T", t)
	lastDot := strings.LastIndexByte(s, '.')
	if lastDot != -1 {
		s = s[lastDot+1:]
	}
	return s
}

// asString returns a human-readable string representation of the passed value
func ValueAsString(value reflect.Value) string {
	b := &bytes.Buffer{}

	walker := func(path string, field *reflect.StructField, v reflect.Value) error {
		if utils.IsPrimitiveValue(v) || v.Kind() == reflect.String {
			fmt.Fprintf(b, "%v", v.Interface())
			return utils.SkipReflection
		}

		switch v.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
			if v.IsNil() {
				fmt.Fprintf(b, "<nil>")
				return utils.SkipReflection
			}
		}

		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			return nil // descend into value

		case reflect.Slice:
			len := v.Len()
			fmt.Fprintf(b, "[")
			for i := 0; i < len; i++ {
				av := v.Index(i)

				if i != 0 {
					fmt.Fprintf(b, ", ")
				}
				fmt.Fprintf(b, "%s", ValueAsString(av))
			}
			fmt.Fprintf(b, "]")
			return utils.SkipReflection

		case reflect.Map:
			keys := v.MapKeys()
			fmt.Fprintf(b, "{")
			for i, key := range keys {
				mv := v.MapIndex(key)

				if i != 0 {
					fmt.Fprintf(b, ", ")
				}
				fmt.Fprintf(b, "%s: %s", ValueAsString(key), ValueAsString(mv))
			}
			fmt.Fprintf(b, "}")
			return utils.SkipReflection

		case reflect.Struct:
			intf := v.Addr().Interface()
			if _, ok := intf.(Resource); ok {
				fmt.Fprintf(b, "<resource>")
			} else if _, ok := intf.(*ResourceHolder); ok {
				fmt.Fprintf(b, "<resource>")
			} else if compareWithID, ok := intf.(CompareWithID); ok {
				id := compareWithID.CompareWithID()
				name := ""
				hasName, ok := intf.(HasName)
				if ok {
					name = StringValue(hasName.GetName())
				}
				if id == nil {
					// Uninformative, but we can often print the name instead
					if name != "" {
						fmt.Fprintf(b, "name:%s", name)
					} else {
						fmt.Fprintf(b, "id:<nil>")
					}
				} else {
					// Uninformative, but we can often print the name instead
					if name != "" {
						fmt.Fprintf(b, "name:%s id:%s", name, *id)
					} else {
						fmt.Fprintf(b, "id:%s", *id)
					}

				}
			} else {
				glog.V(4).Infof("Unhandled kind in asString for %q: %T", path, v.Interface())
				fmt.Fprint(b, DebugAsJsonString(intf))
			}
			return utils.SkipReflection

		default:
			glog.Infof("Unhandled kind in asString for %q: %T", path, v.Interface())
			return fmt.Errorf("Unhandled kind for %q: %v", path, v.Kind())
		}
	}

	err := utils.ReflectRecursive(value, walker)
	if err != nil {
		glog.Fatalf("unexpected error during reflective walk: %v", err)
	}
	return b.String()
}

// Finish is called at the end of a run, and prints a list of changes to the configured Writer
func (t *DryRunTarget) Finish(taskMap map[string]Task) error {
	return t.PrintReport(taskMap, t.out)
}

// HasChanges returns true iff any changes would have been made
func (t *DryRunTarget) HasChanges() bool {
	return len(t.changes)+len(t.deletions) != 0
}

// Not sure what to do with this, but I am getting import cycle not allowed
// TODO: where should I refactor this to?

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

	return ValueAsString(fv)
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
		glog.Fatal("unexpected kind for items: ", itemsValue.Kind())
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
