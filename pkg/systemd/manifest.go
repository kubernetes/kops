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
	"fmt"
)

// Manifest defines a systemd unit
type Manifest struct {
	sections []*section
}

// section defines a section of the unit i.e. Unit, Service etc
type section struct {
	key     string
	content string
	entries []string
}

// Set adds a key/pair to the a section in the systemd manifest
func (m *Manifest) Set(name, key, value string) {
	s := m.getSection(name)
	s.entries = append(s.entries, fmt.Sprintf("%s=%s\n", key, value))
}

// SetSection sets the raw content of a section
func (m *Manifest) SetSection(name, content string) {
	m.getSection(name).content = content
}

// getSection checks if a section already exists
func (m *Manifest) getSection(key string) *section {
	for _, s := range m.sections {
		if s.key == key {
			return s
		}
	}
	// create a new section for this manifest
	s := &section{key: key, entries: make([]string, 0)}
	m.sections = append(m.sections, s)

	return s
}

// Render is responsible for generating the final unit
func (m *Manifest) Render() string {
	var b bytes.Buffer
	size := len(m.sections) - 1

	for i, section := range m.sections {
		b.WriteString(fmt.Sprintf("[%s]\n", section.key))
		if section.content != "" {
			b.WriteString(section.content)
		}
		for _, x := range section.entries {
			b.WriteString(x)
		}
		if i < size {
			b.WriteString("\n")
		}
	}

	return b.String()
}
