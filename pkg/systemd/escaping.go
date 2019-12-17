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

package systemd

import (
	"bytes"
	"encoding/hex"
	"strings"

	"k8s.io/klog"
)

// EscapeCommand is used to escape a command
func EscapeCommand(argv []string) string {
	var escaped []string
	for _, arg := range argv {
		escaped = append(escaped, escapeArg(arg))
	}
	return strings.Join(escaped, " ")
}

func escapeArg(s string) string {
	var b bytes.Buffer

	needQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
			b.WriteByte(c)
			continue
		}

		switch c {
		case '!', '#', '$', '%', '&', '(', ')', '*', '+', ',', '-', '.', '/', ':', ';',
			'<', '>', '=', '?', '@', '[', ']', '^', '_', '`', '{', '|', '}', '~':
			b.WriteByte(c)

		case ' ':
			needQuotes = true
			b.WriteByte(c)

		case '"':
			b.WriteString("\\\"")
		case '\'':
			b.WriteString("\\'")
		case '\\':
			b.WriteString("\\\\")

		default:
			klog.Warningf("Unusual character in systemd command: %v", s)
			b.WriteString("\\x")
			b.WriteString(hex.EncodeToString([]byte{c}))
		}
	}

	if needQuotes {
		return "\"" + b.String() + "\""
	}

	return b.String()
}
