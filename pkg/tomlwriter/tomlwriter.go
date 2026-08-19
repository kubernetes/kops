/*
Copyright 2026 The Kubernetes Authors.

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

// Package tomlwriter provides a minimal, write-only TOML document builder.
//
// It replaces the frozen github.com/pelletier/go-toml v1 subset used for containerd while matching
// its output byte for byte. Scalars precede subtables, each sorted lexicographically; subtable
// headers have a leading blank line and two-space-per-level indentation. The byte-stable output
// makes the migration off go-toml provably a no-op (the containerdbuilder golden fixtures assert
// exact file contents) and keeps files like /etc/containerd/config.toml identical across upgrades.
package tomlwriter

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// Tree represents a TOML document root or table. Its zero value is unusable; use NewTree.
type Tree struct {
	values map[string]any
}

// NewTree returns an empty document.
func NewTree() *Tree {
	return &Tree{values: map[string]any{}}
}

// Table returns or creates the table at path. To match go-toml v1, a scalar path element leaves
// traversal in the current table.
func (t *Tree) Table(path ...string) *Tree {
	subtree := t
	for _, key := range path {
		next, exists := subtree.values[key]
		if !exists {
			next = NewTree()
			subtree.values[key] = next
		}
		if tree, ok := next.(*Tree); ok {
			subtree = tree
		}
	}
	return subtree
}

// SetPath assigns value at path, creating tables with Table's traversal semantics. Path elements
// are unescaped keys, with quoting applied during serialization. Only string, int64, and bool
// values are accepted; other types panic, and a leaf value replaces an existing table.
func (t *Tree) SetPath(path []string, value any) {
	switch value.(type) {
	case string, int64, bool:
	default:
		panic(fmt.Sprintf("tomlwriter: unsupported value type %T at %q", value, strings.Join(path, ".")))
	}

	t.Table(path[:len(path)-1]...).values[path[len(path)-1]] = value
}

// String serializes the document.
func (t *Tree) String() string {
	var sb strings.Builder
	t.write(&sb, "", "")
	return sb.String()
}

// write emits scalars before subtables; keyspace is the dotted, quoted path or empty at root.
func (t *Tree) write(sb *strings.Builder, indent, keyspace string) {
	keys := slices.Sorted(maps.Keys(t.values))

	for _, k := range keys {
		if _, ok := t.values[k].(*Tree); ok {
			continue
		}
		sb.WriteString(indent + quoteKeyIfNeeded(k) + " = " + scalarString(t.values[k]) + "\n")
	}
	for _, k := range keys {
		subtree, ok := t.values[k].(*Tree)
		if !ok {
			continue
		}
		combinedKey := quoteKeyIfNeeded(k)
		if keyspace != "" {
			combinedKey = keyspace + "." + combinedKey
		}
		sb.WriteString("\n" + indent + "[" + combinedKey + "]\n")
		subtree.write(sb, indent+"  ", combinedKey)
	}
}

func scalarString(v any) string {
	switch value := v.(type) {
	case string:
		return `"` + escapeString(value) + `"`
	case int64:
		return strconv.FormatInt(value, 10)
	case bool:
		return strconv.FormatBool(value)
	}
	panic(fmt.Sprintf("tomlwriter: unsupported value type %T", v))
}

// quoteKeyIfNeeded leaves A-Za-z0-9_- keys bare and quotes all others as TOML basic strings.
func quoteKeyIfNeeded(k string) string {
	// Match go-toml v1 by passing through keys already enclosed in double quotes without escaping.
	if len(k) >= 2 && k[0] == '"' && k[len(k)-1] == '"' {
		return k
	}
	// Unlike go-toml v1, quote an empty key to produce valid TOML.
	if k == "" {
		return `""`
	}
	for _, r := range k {
		if !isValidBareChar(r) {
			return `"` + escapeString(k) + `"`
		}
	}
	return k
}

func isValidBareChar(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '_' || r == '-'
}

// escapeString matches the go-toml v1 escaping of TOML basic strings.
func escapeString(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch r {
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			if r < 0x1F {
				fmt.Fprintf(&b, "\\u%0.4X", uint16(r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
