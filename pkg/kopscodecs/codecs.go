/*
Copyright 2016 The Kubernetes Authors.

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

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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

func encoder(gv runtime.GroupVersioner, mediaType string) runtime.Encoder {
	e, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		glog.Fatalf("no %s serializer registered", mediaType)
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
	return ToVersionedYamlWithVersion(obj, v1alpha2.SchemeGroupVersion)
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
	return ToVersionedJSONWithVersion(obj, v1alpha2.SchemeGroupVersion)
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

func ParseVersionedYaml(data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	return decoder().Decode(data, nil, nil)
}
