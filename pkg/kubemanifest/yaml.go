/*
Copyright 2021 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// KubeObjectToApplyYAML returns the kubernetes object converted to YAML, with "noisy" fields removed.
//
// We remove:
//   - status (can't be applied, shouldn't be specified)
//   - metadata.creationTimestamp (can't be applied, shouldn't be specified)
func KubeObjectToApplyYAML(data runtime.Object) (string, error) {
	// This logic is inlined sigs.k8s.io/yaml.Marshal, but we delete some fields in the middle.

	// Convert the object to JSON bytes
	j, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling into JSON: %v", err)
	}

	// Convert the JSON to a map.
	jsonObj := make(map[string]interface{})
	if err := yaml.Unmarshal(j, &jsonObj); err != nil {
		return "", err
	}

	// Remove status (can't be applied, shouldn't be specified)
	delete(jsonObj, "status")

	// Remove metadata.creationTimestamp (can't be applied, shouldn't be specified)
	metadataObj, found := jsonObj["metadata"]
	if found {
		if metadata, ok := metadataObj.(map[string]interface{}); ok {
			delete(metadata, "creationTimestamp")
		} else {
			klog.Warningf("unexpected type for object metadata: %T", metadataObj)
		}
	} else {
		klog.Warningf("object did not have metadata: %#v", jsonObj)
	}

	// Marshal the cleaned-up map into YAML.
	y, err := yaml.Marshal(jsonObj)
	if err != nil {
		return "", err
	}
	return string(y), nil
}
