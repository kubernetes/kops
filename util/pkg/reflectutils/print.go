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

package reflectutils

import (
	"bytes"
	"fmt"
	"reflect"

	"k8s.io/klog"

	"k8s.io/kops/pkg/values"
)

// Printer is a custom printer function, so we can add special display for objects
// (without introducing a package dependency)
type Printer func(o interface{}) (string, bool)

var printers []Printer

// RegisterPrinter adds a custom printer function
func RegisterPrinter(p Printer) {
	printers = append(printers, p)
}

// ValueAsString returns a human-readable string representation of the passed value
func ValueAsString(value reflect.Value) string {
	b := &bytes.Buffer{}

	walker := func(path string, field *reflect.StructField, v reflect.Value) error {
		if IsPrimitiveValue(v) || v.Kind() == reflect.String {
			fmt.Fprintf(b, "%v", v.Interface())
			return SkipReflection
		}

		switch v.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
			if v.IsNil() {
				fmt.Fprintf(b, "<nil>")
				return SkipReflection
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
			return SkipReflection

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
			return SkipReflection

		case reflect.Struct:
			intf := v.Addr().Interface()

			done := false
			for _, p := range printers {
				s, ok := p(intf)
				if ok {
					fmt.Fprintf(b, "%s", s)
					done = true
					break
				}
			}

			if !done {
				klog.V(4).Infof("Unhandled kind in asString for %q: %T", path, v.Interface())
				fmt.Fprint(b, values.DebugAsJsonString(intf))
			}

			return SkipReflection

		default:
			klog.Infof("Unhandled kind in asString for %q: %T", path, v.Interface())
			return fmt.Errorf("Unhandled kind for %q: %v", path, v.Kind())
		}
	}

	err := ReflectRecursive(value, walker)
	if err != nil {
		klog.Fatalf("unexpected error during reflective walk: %v", err)
	}
	return b.String()
}
