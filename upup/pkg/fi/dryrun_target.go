package fi

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"io"
	"reflect"
)

// DryRunTarget is a special Target that does not execute anything, but instead tracks all changes.
// By running against a DryRunTarget, a list of changes that would be made can be easily collected,
// without any special support from the Tasks.
type DryRunTarget struct {
	changes []*render

	// The destination to which the final report will be printed on Finish()
	out io.Writer
}

type render struct {
	a       Task
	aIsNil  bool
	e       Task
	changes Task
}

var _ Target = &DryRunTarget{}

func NewDryRunTarget(out io.Writer) *DryRunTarget {
	t := &DryRunTarget{}
	t.out = out
	return t
}

func (t *DryRunTarget) Render(a, e, changes Task) error {
	valA := reflect.ValueOf(a)
	aIsNil := valA.IsNil()

	t.changes = append(t.changes, &render{
		a:       a,
		aIsNil:  aIsNil,
		e:       e,
		changes: changes,
	})
	return nil
}

func IdForTask(taskMap map[string]Task, t Task) string {
	for k, v := range taskMap {
		if v == t {
			return k
		}
	}
	glog.Fatalf("unknown task: %v", t)
	return "?"
}

func (t *DryRunTarget) PrintReport(taskMap map[string]Task, out io.Writer) error {
	b := &bytes.Buffer{}

	if len(t.changes) != 0 {
		fmt.Fprintf(b, "Created resources:\n")
		for _, r := range t.changes {
			if !r.aIsNil {
				continue
			}

			fmt.Fprintf(b, "  %T\t%s\n", r.changes, IdForTask(taskMap, r.e))
		}

		fmt.Fprintf(b, "Changed resources:\n")
		for _, r := range t.changes {
			if r.aIsNil {
				continue
			}
			var changeList []string

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
					fieldValC := valC.Field(i)

					if (fieldValC.Kind() == reflect.Ptr || fieldValC.Kind() == reflect.Slice || fieldValC.Kind() == reflect.Map) && fieldValC.IsNil() {
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
						default:
							description = fmt.Sprintf(" %v -> %v", asString(fieldValA), asString(fieldValE))
						}
					}
					if ignored {
						continue
					}
					changeList = append(changeList, valC.Type().Field(i).Name+description)
				}
			} else {
				return fmt.Errorf("unhandled change type: %v", valC.Type())
			}

			if len(changeList) == 0 {
				continue
			}

			fmt.Fprintf(b, "  %T\t%s\n", r.changes, IdForTask(taskMap, r.e))
			for _, f := range changeList {
				fmt.Fprintf(b, "    %s\n", f)
			}
			fmt.Fprintf(b, "\n")
		}
	}

	_, err := out.Write(b.Bytes())
	return err
}

// asString returns a human-readable string representation of the passed value
func asString(v reflect.Value) string {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "<nil>"
		}
	}
	if v.CanInterface() {
		iv := v.Interface()
		_, isResource := iv.(Resource)
		if isResource {
			return "<resource>"
		}
		_, isHasID := iv.(CompareWithID)
		if isHasID {
			id := iv.(CompareWithID).CompareWithID()
			if id == nil {
				return "id:<nil>"
			} else {
				return "id:" + *id
			}
		}
		switch typed := iv.(type) {
		case *string:
			return *typed
		case *bool:
			return fmt.Sprintf("%v", *typed)
		default:
			return fmt.Sprintf("%T (%v)", iv, iv)
		}

	} else {
		return fmt.Sprintf("Unhandled: %T", v.Type())

	}
}

// Finish is called at the end of a run, and prints a list of changes to the configured Writer
func (t *DryRunTarget) Finish(taskMap map[string]Task) error {
	return t.PrintReport(taskMap, t.out)
}
