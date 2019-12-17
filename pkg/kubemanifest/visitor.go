/*
Copyright 2017 The Kubernetes Authors.

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

package kubemanifest

import (
	"fmt"
	"strings"

	"k8s.io/klog"
)

type visitorBase struct {
}

func (m *visitorBase) VisitString(path []string, v string, mutator func(string)) error {
	klog.V(10).Infof("string value at %s: %s", strings.Join(path, "."), v)
	return nil
}

func (m *visitorBase) VisitBool(path []string, v bool, mutator func(bool)) error {
	klog.V(10).Infof("string value at %s: %v", strings.Join(path, "."), v)
	return nil
}

func (m *visitorBase) VisitFloat64(path []string, v float64, mutator func(float64)) error {
	klog.V(10).Infof("float64 value at %s: %f", strings.Join(path, "."), v)
	return nil
}

type Visitor interface {
	VisitBool(path []string, v bool, mutator func(bool)) error
	VisitString(path []string, v string, mutator func(string)) error
	VisitFloat64(path []string, v float64, mutator func(float64)) error
}

func visit(visitor Visitor, data interface{}, path []string, mutator func(interface{})) error {
	switch data := data.(type) {
	case string:
		err := visitor.VisitString(path, data, func(v string) {
			mutator(v)
		})
		if err != nil {
			return err
		}

	case bool:
		err := visitor.VisitBool(path, data, func(v bool) {
			mutator(v)
		})
		if err != nil {
			return err
		}

	case float64:
		err := visitor.VisitFloat64(path, data, func(v float64) {
			mutator(v)
		})
		if err != nil {
			return err
		}

	case map[string]interface{}:
		m := data
		for k, v := range m {
			path = append(path, k)

			err := visit(visitor, v, path, func(v interface{}) {
				m[k] = v
			})
			if err != nil {
				return err
			}
			path = path[:len(path)-1]
		}

	case []interface{}:
		s := data
		for i, v := range s {
			path = append(path, fmt.Sprintf("[%d]", i))

			err := visit(visitor, v, path, func(v interface{}) {
				s[i] = v
			})
			if err != nil {
				return err
			}
			path = path[:len(path)-1]
		}

	default:
		return fmt.Errorf("unhandled type in manifest: %T", data)
	}

	return nil
}
