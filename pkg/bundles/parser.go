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

package bundles

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
)

func ParseToTypedObjects(component *Component, scheme *runtime.Scheme) ([]runtime.Object, error) {
	pretty := false
	jsonSerializer := json.NewSerializer(json.DefaultMetaFactory, scheme, scheme, pretty)

	var objects []runtime.Object
	for _, obj := range component.Spec.Objects {
		j, err := obj.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("error marshalling unstructured to JSON: %v", err)
		}

		obj, _, err := jsonSerializer.Decode(j, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error parsing json: %v", err)
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

// ParseBytes parses a set of objects from a []byte
func ParseBytes(data []byte, yamlDecoder runtime.Decoder) ([]runtime.Object, error) {
	//yamlDecoder := yaml.NewDecodingSerializer(serializer)

	reader := json.YAMLFramer.NewFrameReader(ioutil.NopCloser(bytes.NewReader([]byte(data))))
	d := streaming.NewDecoder(reader, yamlDecoder)

	var objects []runtime.Object
	for {
		obj, _, err := d.Decode(nil, nil)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error during parse: %v", err)
		}
		objects = append(objects, obj)
	}

	return objects, nil
}
