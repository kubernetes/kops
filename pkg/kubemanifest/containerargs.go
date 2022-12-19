/*
Copyright 2023 The Kubernetes Authors.

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
)

type ContainerVisitorFunction func(container map[string]interface{}) error

func (m *Object) VisitContainers(visitorFn ContainerVisitorFunction) error {
	visitorObj := &containerVisitor{
		visitor: visitorFn,
	}
	err := m.accept(visitorObj)
	if err != nil {
		return err
	}
	return nil
}

type containerVisitor struct {
	visitorBase
	visitor ContainerVisitorFunction
}

func (m *containerVisitor) VisitMap(path []string, v map[string]interface{}) error {
	n := len(path)
	if n < 2 || path[n-2] != "containers" || !strings.HasPrefix(path[n-1], "[") {
		return nil
	}

	if err := m.visitor(v); err != nil {
		return fmt.Errorf("error visiting container %v: %w", v, err)
	}

	return nil
}
