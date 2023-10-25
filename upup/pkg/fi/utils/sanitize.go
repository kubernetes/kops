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

package utils

import (
	"bytes"
	"strings"

	"k8s.io/client-go/util/homedir"
)

// SanitizeString iterated a strings, removes any characters not in the allow list and returns at most 200 characters
func SanitizeString(s string) string {
	var buf bytes.Buffer
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	for _, c := range s {
		if strings.ContainsRune(allowed, c) {
			buf.WriteRune(c)
		} else {
			buf.WriteRune('_')
		}
	}

	out := buf.String()
	if len(out) > 200 {
		out = out[len(out)-200:]
	}

	return out
}

// ExpandPath replaces common path aliases: ~ -> $HOME
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		p = homedir.HomeDir() + p[1:]
	}

	return p
}
