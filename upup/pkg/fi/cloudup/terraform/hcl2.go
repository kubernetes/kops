/*
Copyright 2020 The Kubernetes Authors.

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
	"strings"

	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type mapStringLiteral struct {
	members map[string]*terraformWriter.Literal
}

func (m *mapStringLiteral) IsSingleValue() bool {
	return false
}

func (m *mapStringLiteral) ToObject() element {
	o := &object{field: make(map[string]element, len(m.members))}
	for k, v := range m.members {
		o.field[k] = v
	}
	return o
}

// write writes a map's key-value pairs to a body spread across multiple lines.
// Example:
//
//	key = {
//	  "key1" = "value1"
//	  "key2" = "value2"
//	}
func (m *mapStringLiteral) Write(buffer *bytes.Buffer, indent int, key string) {
	if len(m.members) == 0 {
		return
	}
	writeIndent(buffer, indent)
	buffer.WriteString(key)
	buffer.WriteString(" = {\n")
	keys := make([]string, 0, len(m.members))
	maxKeyLen := 0
	for k := range m.members {
		kLen := len(quote(k))
		if kLen > maxKeyLen {
			maxKeyLen = kLen
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		writeIndent(buffer, indent+2)
		quoted := quote(k)
		buffer.WriteString(quoted)
		writeIndent(buffer, maxKeyLen-len(quoted))
		buffer.WriteString(" = ")
		buffer.WriteString(m.members[k].String)
		buffer.WriteRune('\n')
	}
	writeIndent(buffer, indent)
	buffer.WriteString("}\n")
}

func mapToElement(item interface{}) *mapStringLiteral {
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Map {
		panic(fmt.Sprintf("not a map type %s", v.Kind()))
	}
	if v.Type().Key().Kind() != reflect.String {
		panic(fmt.Sprintf("unhandled map key type %s", v.Type().Key().Kind()))
	}
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Pointer && elemType.Elem() == literalType {
		o := &mapStringLiteral{members: make(map[string]*terraformWriter.Literal, v.Len())}
		for _, key := range v.MapKeys() {
			o.members[key.String()] = v.MapIndex(key).Interface().(*terraformWriter.Literal)
		}
		return o
	}
	if elemType.Kind() != reflect.String {
		panic(fmt.Sprintf("unhandled map value type %s", elemType.Kind()))
	}
	o := &mapStringLiteral{members: make(map[string]*terraformWriter.Literal, v.Len())}
	for _, key := range v.MapKeys() {
		o.members[key.String()] = terraformWriter.LiteralFromStringValue(v.MapIndex(key).String())
	}
	return o
}

func writeIndent(buf *bytes.Buffer, indent int) {
	for i := 0; i < indent; i++ {
		buf.WriteString(" ")
	}
}

func quote(s string) string {
	var b strings.Builder
	b.WriteRune('"')
	for _, r := range s {
		if r == '\\' || r == '"' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	b.WriteRune('"')
	return b.String()
}
