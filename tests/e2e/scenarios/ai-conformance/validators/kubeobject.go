/*
Copyright The Kubernetes Authors.

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

package validators

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// KubeObjectID represents a Kubernetes object parsed from a manifest.
type KubeObjectID struct {
	h *ValidatorHarness

	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

// GVK returns the GroupVersionKind of the object.
func (o *KubeObjectID) GVK() schema.GroupVersionKind {
	return o.gvk
}

// Kind returns the Kind of the object.
func (o *KubeObjectID) Kind() string {
	return o.gvk.Kind
}

// Name returns the name of the object.
func (o *KubeObjectID) Name() string {
	return o.name
}

// Namespace returns the namespace of the object.
func (o *KubeObjectID) Namespace() string {
	return o.namespace
}

// KubectlWaitOption configures the behavior of KubectlWait.
type KubectlWaitOption func(*kubectlWaitOptions)

type kubectlWaitOptions struct {
	timeout string
}

// WithTimeout sets the timeout for kubectl wait.
func WithTimeout(timeout string) KubectlWaitOption {
	return func(o *kubectlWaitOptions) {
		o.timeout = timeout
	}
}

// KubectlWait waits for the object to become healthy using kubectl wait.
// The wait condition is determined by the object's kind.
// Objects that don't have a meaningful wait condition are skipped.
func (o *KubeObjectID) KubectlWait(opts ...KubectlWaitOption) {
	condition := waitConditionForKind(o.gvk.Kind)
	if condition == "" {
		o.h.Errorf("No wait condition for %s/%s, cannot wait", o.gvk.Kind, o.name)
		return
	}

	options := &kubectlWaitOptions{timeout: "300s"}
	for _, opt := range opts {
		opt(options)
	}

	resourceType := kubectlResourceType(o.gvk)
	o.h.ShellExec(fmt.Sprintf("kubectl wait -n %s %s %s/%s --timeout=%s",
		o.namespace, condition, resourceType, o.name, options.timeout))
}

// waitConditionForKind returns the kubectl wait --for condition appropriate for the given kind.
// Returns empty string for kinds that don't have a meaningful wait condition.
func waitConditionForKind(kind string) string {
	switch kind {
	case "HTTPRoute":
		return "--for=jsonpath='{.status.parents[0].conditions[?(@.type==\"Accepted\")].status}'=True"
	case "Gateway":
		return "--for=condition=Programmed"
	case "Deployment":
		return "--for=condition=Available"
	case "Pod":
		return "--for=condition=Ready"
	case "Job":
		return "--for=condition=Complete"
	default:
		return ""
	}
}

// kubectlResourceType returns the kubectl resource type string for a GVK.
// For core API resources, this is the lowercase kind.
// For other groups, this is "kind.group" (lowercased).
func kubectlResourceType(gvk schema.GroupVersionKind) string {
	kind := strings.ToLower(gvk.Kind)
	if gvk.Group == "" {
		return kind
	}
	return kind + "." + gvk.Group
}

// parseManifestObjects parses a multi-document YAML manifest file and returns the objects found.
func (h *ValidatorHarness) parseManifestObjects(manifestPath string, defaultNamespace string) ([]*KubeObjectID, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest %s: %w", manifestPath, err)
	}

	var objects []*KubeObjectID
	reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	for {
		doc, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading YAML document from %s: %w", manifestPath, err)
		}

		// Skip empty documents
		doc = bytes.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}

		obj, err := h.parseMinimalObject(doc, defaultNamespace)
		if err != nil {
			return nil, fmt.Errorf("parsing object from %s: %w", manifestPath, err)
		}
		if obj != nil {
			objects = append(objects, obj)
		}
	}

	return objects, nil
}

// parseMinimalObject extracts GVK and name from a YAML document without full deserialization.
func (h *ValidatorHarness) parseMinimalObject(doc []byte, defaultNamespace string) (*KubeObjectID, error) {
	// Use the YAML-to-JSON utility to decode into a map
	jsonData, err := yaml.ToJSON(doc)
	if err != nil {
		return nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	// Quick parse using the unstructured decoder
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	apiVersion, _ := raw["apiVersion"].(string)
	kind, _ := raw["kind"].(string)
	if apiVersion == "" || kind == "" {
		return nil, nil
	}

	metadata, _ := raw["metadata"].(map[string]interface{})
	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)

	if namespace == "" {
		namespace = defaultNamespace
	}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, fmt.Errorf("parsing apiVersion %q: %w", apiVersion, err)
	}

	return &KubeObjectID{
		h:         h,
		gvk:       gv.WithKind(kind),
		name:      name,
		namespace: namespace,
	}, nil
}
