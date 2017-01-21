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

package systemd

import "bytes"

type Manifest struct {
	Sections []*ManifestSection
}

type ManifestSection struct {
	Key     string
	Entries []*ManifestEntry
}

type ManifestEntry struct {
	Key   string
	Value string
}

func (s *Manifest) Set(sectionKey string, key string, value string) {
	section := s.getOrCreateSection(sectionKey)
	section.Set(key, value)
}

func (s *Manifest) getOrCreateSection(key string) *ManifestSection {
	for _, section := range s.Sections {
		if section.Key == key {
			return section
		}
	}
	section := &ManifestSection{
		Key: key,
	}
	s.Sections = append(s.Sections, section)
	return section
}

func (s *Manifest) Render() string {
	var b bytes.Buffer

	for i, section := range s.Sections {
		if i != 0 {
			b.WriteString("\n")
		}
		b.WriteString(section.Render())
	}

	return b.String()
}

func (s *ManifestSection) Set(key string, value string) {
	for _, entry := range s.Entries {
		if entry.Key == key {
			entry.Value = value
			return
		}
	}

	entry := &ManifestEntry{
		Key:   key,
		Value: value,
	}
	s.Entries = append(s.Entries, entry)
}

func (s *ManifestSection) Render() string {
	var b bytes.Buffer

	b.WriteString("[" + s.Key + "]\n")
	for _, entry := range s.Entries {
		b.WriteString(entry.Key + "=" + entry.Value + "\n")
	}

	return b.String()
}
