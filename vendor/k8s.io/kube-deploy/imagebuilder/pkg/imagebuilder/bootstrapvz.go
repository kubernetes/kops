/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package imagebuilder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// NewBootstrapVzTemplate builds a BootstrapVzTemplate from a file
func NewBootstrapVzTemplate(data string) (*BootstrapVzTemplate, error) {
	m := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}
	return &BootstrapVzTemplate{data: m, raw: []byte(data)}, nil
}

// BootstrapVzTemplate represents a bootstrap-vz template file
type BootstrapVzTemplate struct {
	data map[interface{}]interface{}
	raw  []byte
}

// Bytes returns the template contents
func (t *BootstrapVzTemplate) Bytes() []byte {
	return t.raw
}

// BuildImageName computes the name of the image that will be built
func (t *BootstrapVzTemplate) BuildImageName() (string, error) {
	name, err := t.getString("name")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("name not found in template")
	}
	regex := regexp.MustCompile("{([^}]*)}")

	now := time.Now().UTC()

	var replaceErr error

	replacer := func(path string) string {
		// Remove { and }
		path = path[1 : len(path)-1]

		if path == "" {
			return ""
		}

		if path[0] == '%' {
			switch path {
			case "%Y":
				return strconv.Itoa(now.Year())
			case "%m":
				return fmt.Sprintf("%02d", now.Month())
			case "%d":
				return fmt.Sprintf("%02d", now.Day())
			default:
				replaceErr = fmt.Errorf("unknown template specifier: %q", path)
				return ""
			}
		} else {
			v, err := t.getString(path)
			if err != nil {
				replaceErr = fmt.Errorf("error replacing template spec %q: %v", path, err)
				return ""
			}
			return v
		}
	}

	name = regex.ReplaceAllStringFunc(name, replacer)
	if replaceErr != nil {
		return "", replaceErr
	}
	return name, nil
}

func (t *BootstrapVzTemplate) getString(path string) (string, error) {
	tokens := strings.Split(path, ".")
	pos := t.data
	for i, token := range tokens {
		next, found := pos[token]
		if !found {
			return "", nil
		}

		if (i + 1) == len(tokens) {
			s, ok := next.(string)
			if !ok {
				return "", fmt.Errorf("Expected string, found %T at %q", next, path)
			}
			return s, nil
		}

		m, ok := next.(map[interface{}]interface{})
		if !ok {
			return "", fmt.Errorf("Expected map, found %T at %q", next, path)
		}
		pos = m
	}

	panic("unreachable")
}
