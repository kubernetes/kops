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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/install"
	"k8s.io/kops/pkg/apis/kops/v1beta1"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	install.Install(Scheme)
}

func encoder(gv runtime.GroupVersioner, mediaType string) runtime.Encoder {
	e, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		klog.Fatalf("no %s serializer registered", mediaType)
	}
	return Codecs.EncoderForVersion(e.Serializer, gv)
}

func decoder() runtime.Decoder {
	// TODO: Cache?
	// Codecs provides access to encoding and decoding for the scheme
	codec := Codecs.UniversalDecoder(kops.SchemeGroupVersion)
	return codec
}

// ToVersionedYaml encodes the object to YAML
func ToVersionedYaml(obj runtime.Object) ([]byte, error) {
	return ToVersionedYamlWithVersion(obj, v1beta1.SchemeGroupVersion)
}

// ToVersionedYamlWithVersion encodes the object to YAML, in a specified API version
func ToVersionedYamlWithVersion(obj runtime.Object, version runtime.GroupVersioner) ([]byte, error) {
	var w bytes.Buffer
	err := encoder(version, "application/yaml").Encode(obj, &w)
	if err != nil {
		return nil, fmt.Errorf("error encoding %T: %v", obj, err)
	}
	return w.Bytes(), nil
}

// ToVersionedJSON encodes the object to JSON
func ToVersionedJSON(obj runtime.Object) ([]byte, error) {
	return ToVersionedJSONWithVersion(obj, v1beta1.SchemeGroupVersion)
}

// ToVersionedJSONWithVersion encodes the object to JSON, in a specified API version
func ToVersionedJSONWithVersion(obj runtime.Object, version runtime.GroupVersioner) ([]byte, error) {
	var w bytes.Buffer
	err := encoder(version, "application/json").Encode(obj, &w)
	if err != nil {
		return nil, fmt.Errorf("error encoding %T: %v", obj, err)
	}
	return w.Bytes(), nil
}

// Decode decodes the specified data, with the specified default version
func Decode(data []byte, defaultReadVersion *schema.GroupVersionKind) (runtime.Object, *schema.GroupVersionKind, error) {
	data = rewriteAPIGroup(data)

	decoder := decoder()

	object, gvk, err := decoder.Decode(data, defaultReadVersion, nil)
	return object, gvk, err
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
