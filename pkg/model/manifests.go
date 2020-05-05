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

package model

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
)

// ParseManifest parses a set of objects from a []byte
func ParseManifest(data []byte) ([]runtime.Object, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	deser := scheme.Codecs.UniversalDeserializer()

	var objects []runtime.Object

	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s", string(data))
			klog.Infof("manifest: %s", string(data))
			return nil, fmt.Errorf("error parsing manifest: %v", err)
		}

		obj, _, err := deser.Decode([]byte(ext.Raw), nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error parsing object in manifest: %v", err)
		}

		objects = append(objects, obj)
	}

	return objects, nil
}
