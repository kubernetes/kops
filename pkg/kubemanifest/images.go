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

type ImageRemapFunction func(image string) (string, error)

func (m *Manifest) RemapImages(mapper ImageRemapFunction) error {
	visitor := &imageRemapVisitor{
		mapper: mapper,
	}
	err := m.accept(visitor)
	if err != nil {
		return err
	}
	return nil
}

type imageRemapVisitor struct {
	visitorBase
	mapper ImageRemapFunction
}

func (m *imageRemapVisitor) VisitString(path []string, v string, mutator func(string)) error {
	n := len(path)
	if n < 1 || path[n-1] != "image" {
		return nil
	}

	// Deployments/DaemonSets/Jobs/StatefulSets have two image fields
	//	- spec.template.spec.containers.[2].image
	//  - spec.template.spec.initContainers.[2].image
	if n < 3 || (path[n-3] != "containers" && path[n-3] != "initContainers") {
		klog.Warningf("Skipping likely image field: %s", strings.Join(path, "."))
		return nil
	}

	image := v
	klog.V(4).Infof("Consider image for re-mapping: %q", image)
	remapped, err := m.mapper(v)
	if err != nil {
		return fmt.Errorf("error remapping image %q: %v", image, err)
	}
	if remapped != image {
		mutator(remapped)
	}
	return nil
}
