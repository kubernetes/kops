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

package kopscodecs

import (
	"bytes"
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubeyaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/install"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	install.Install(Scheme)
}

// ToVersionedYaml encodes the object to YAML
func ToVersionedYaml(obj runtime.Object) ([]byte, error) {
	return ToVersionedYamlWithVersion(obj, v1alpha2.SchemeGroupVersion)
}

// ToMediaTypeWithVersion encodes the object to the specified mediaType, in a specified API version
func ToMediaTypeWithVersion(obj runtime.Object, mediaType string, gv runtime.GroupVersioner) ([]byte, error) {
	e, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, fmt.Errorf("no serializer for %q", mediaType)
	}

	_, isUnstructured := obj.(*unstructured.Unstructured)
	var w bytes.Buffer
	if isUnstructured {
		err := e.Serializer.Encode(obj, &w)
		if err != nil {
			return nil, fmt.Errorf("error encoding %T with unstructured encoder: %w", obj, err)
		}
	} else {
		encoder := Codecs.EncoderForVersion(e.Serializer, gv)
		if err := encoder.Encode(obj, &w); err != nil {
			return nil, fmt.Errorf("error encoding %T with structured encoder: %w", obj, err)
		}
	}
	return w.Bytes(), nil
}

// ToVersionedYamlWithVersion encodes the object to YAML, in a specified API version
func ToVersionedYamlWithVersion(obj runtime.Object, version runtime.GroupVersioner) ([]byte, error) {
	return ToMediaTypeWithVersion(obj, "application/yaml", version)
}

// ToVersionedJSON encodes the object to JSON
func ToVersionedJSON(obj runtime.Object) ([]byte, error) {
	return ToVersionedJSONWithVersion(obj, v1alpha2.SchemeGroupVersion)
}

// ToVersionedJSONWithVersion encodes the object to JSON, in a specified API version
func ToVersionedJSONWithVersion(obj runtime.Object, version runtime.GroupVersioner) ([]byte, error) {
	return ToMediaTypeWithVersion(obj, "application/json", version)
}

// Decode decodes the specified data, with the specified default version
func Decode(data []byte, defaultReadVersion *schema.GroupVersionKind) (runtime.Object, *schema.GroupVersionKind, error) {
	u := &unstructured.Unstructured{}

	// First decode into unstructured.Unstructured so we get the GVK
	unstructuredDecoder := kubeyaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj, gvk, err := unstructuredDecoder.Decode(data, nil, u)
	if err != nil {
		return obj, gvk, err
	}

	// If this isn't a kOps type, return it as unstructured
	if gvk.Group != "kops.k8s.io" && gvk.Group != "kops" {
		return u, gvk, nil
	}

	// Remap the "kops" group => kops.k8s.io
	if gvk.Group == "kops" {
		data = rewriteAPIGroup(data)
	}

	// Decode into kops types
	// TODO: Cache kopsDecoder?
	kopsDecoder := Codecs.UniversalDecoder(kops.SchemeGroupVersion)
	return kopsDecoder.Decode(data, defaultReadVersion, nil)
}

// rewriteAPIGroup rewrites the apiVersion from kops/v1alphaN -> kops.k8s.io/v1alphaN
// This allows us to register as a normal CRD
func rewriteAPIGroup(y []byte) []byte {
	changed := false

	lines := bytes.Split(y, []byte("\n"))
	for i := range lines {
		if !bytes.Contains(lines[i], []byte("apiVersion:")) {
			continue
		}

		{
			re := regexp.MustCompile("kops/v1alpha2")
			lines[i] = re.ReplaceAllLiteral(lines[i], []byte("kops.k8s.io/v1alpha2"))
			changed = true
		}
	}

	if changed {
		y = bytes.Join(lines, []byte("\n"))
	}

	return y
}
