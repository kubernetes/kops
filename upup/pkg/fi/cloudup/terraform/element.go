/*
Copyright 2022 The Kubernetes Authors.

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

package terraform

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// element is some sort of value that can be written in HCL
type element interface {
	IsSingleValue() bool
	Write(buffer *bytes.Buffer, indent int, key string)
}

type object struct {
	field map[string]element
}

var _ element = &object{}

func (o *object) IsSingleValue() bool {
	return false
}

func (o *object) Write(buffer *bytes.Buffer, indent int, key string) {
	writeIndent(buffer, indent)
	buffer.WriteString(key)
	buffer.WriteString(" {\n")
	keys := make([]string, 0, len(o.field))
	for key := range o.field {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	maxKeyLen := 0
	for i, key := range keys {
		if o.field[key].IsSingleValue() {
			if maxKeyLen == 0 {
				for j := i; j < len(keys) && o.field[keys[j]].IsSingleValue(); j++ {
					if len(keys[j]) > maxKeyLen {
						maxKeyLen = len(keys[j])
					}
				}
				maxKeyLen++
			}
			writeIndent(buffer, indent+2)
			buffer.WriteString(key)
			writeIndent(buffer, maxKeyLen-len(key))
		} else {
			maxKeyLen = 0
		}
		o.field[key].Write(buffer, indent+2, key)
	}
	writeIndent(buffer, indent)
	buffer.WriteString("}\n")
}

type sliceObject struct {
	members []element
}

func (s *sliceObject) IsSingleValue() bool {
	return false
}

func (s *sliceObject) Write(buffer *bytes.Buffer, indent int, key string) {
	for _, member := range s.members {
		member.Write(buffer, indent, key)
	}
}

var literalType = reflect.TypeOf(terraformWriter.Literal{})

func toElement(item interface{}) element {
	if literal, ok := item.(*terraformWriter.Literal); ok {
		if literal == nil {
			return nil
		}
		return literal
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		return terraformWriter.LiteralTokens(strconv.FormatBool(v.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return terraformWriter.LiteralTokens(strconv.FormatInt(v.Int(), 10))
	case reflect.Map:
		return mapToElement(v.Interface())
	case reflect.String:
		return terraformWriter.LiteralFromStringValue(v.String())
	case reflect.Struct:
		o := &object{
			field: map[string]element{},
		}
		for _, field := range reflect.VisibleFields(v.Type()) {
			element := toElement(v.FieldByIndex(field.Index).Interface())
			if element != nil {
				o.field[fieldKey(field)] = element
			}
		}
		return o
	case reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		elemType := v.Type().Elem()
		if elemType.Kind() == reflect.Pointer {
			elemType = elemType.Elem()
			if elemType == literalType {
				elements := make([]*terraformWriter.Literal, v.Len())
				for i := range elements {
					elem := v.Index(i)
					elements[i] = elem.Interface().(*terraformWriter.Literal)
				}
				return terraformWriter.LiteralListExpression(elements...)
			}
		}
		switch elemType.Kind() {
		case reflect.String:
			elements := make([]*terraformWriter.Literal, v.Len())
			for i := range elements {
				elem := v.Index(i)
				if elem.Kind() == reflect.Pointer {
					// TODO can these ever be nil?
					elem = elem.Elem()
				}
				elements[i] = terraformWriter.LiteralFromStringValue(elem.String())
			}
			return terraformWriter.LiteralListExpression(elements...)
		case reflect.Struct:
			o := &sliceObject{members: make([]element, v.Len())}
			for i := range o.members {
				elem := toElement(v.Index(i).Interface())
				o.members[i] = elem
			}
			return o
		default:
			panic(fmt.Sprintf("unhandled slice member kind %s", elemType.Kind()))
		}
	default:
		panic(fmt.Sprintf("unhandled kind %s", v.Kind()))
	}
}

func fieldKey(field reflect.StructField) string {
	key := field.Tag.Get("cty")
	if key != "" {
		return key
	}

	var b strings.Builder
	prev_lower := false
	for _, r := range field.Name {
		if unicode.IsUpper(r) {
			if prev_lower {
				b.WriteRune('_')
				prev_lower = false
			}
		} else {
			prev_lower = true
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
