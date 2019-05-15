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

	"github.com/ghodss/yaml"
	"k8s.io/klog"
	"k8s.io/kops/util/pkg/text"
)

type Manifest struct {
	data map[string]interface{}
}

func LoadManifestsFrom(contents []byte) ([]*Manifest, error) {
	var manifests []*Manifest

	// TODO: Support more separators?
	sections := text.SplitContentToSections(contents)

	for _, section := range sections {
		data := make(map[string]interface{})
		err := yaml.Unmarshal(section, &data)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml: %v", err)
		}

		manifest := &Manifest{
			//bytes: section,
			data: data,
		}
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func (m *Manifest) ToYAML() ([]byte, error) {
	b, err := yaml.Marshal(m.data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling manifest to yaml: %v", err)
	}
	return b, nil
}

func (m *Manifest) accept(visitor Visitor) error {
	err := visit(visitor, m.data, []string{}, func(v interface{}) {
		klog.Fatal("cannot mutate top-level data")
	})
	return err
}
