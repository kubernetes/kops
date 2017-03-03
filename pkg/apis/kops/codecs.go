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

package kops

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func decoder() runtime.Decoder {
	// TODO: Cache?
	// Codecs provides access to encoding and decoding for the scheme
	codecs := k8sapi.Codecs
	codec := codecs.UniversalDecoder(SchemeGroupVersion)
	return codec
}

func encoder(version string) runtime.Encoder {
	// TODO: Which is better way to build yaml?
	//yaml := json.NewYAMLSerializer(json.DefaultMetaFactory, k8sapi.Scheme, k8sapi.Scheme)

	// TODO: Cache?
	yaml, ok := runtime.SerializerInfoForMediaType(k8sapi.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		glog.Fatalf("no YAML serializer registered")
	}
	gv := schema.GroupVersion{Group: GroupName, Version: version}
	return k8sapi.Codecs.EncoderForVersion(yaml.Serializer, gv)
}

func preferredAPIVersion() string {
	return "v1alpha2"
}

// ToVersionedYaml encodes the object to YAML, in our preferred API version
func ToVersionedYaml(obj runtime.Object) ([]byte, error) {
	return ToVersionedYamlWithVersion(obj, preferredAPIVersion())
}

// ToVersionedYamlWithVersion encodes the object to YAML, in a specified API version
func ToVersionedYamlWithVersion(obj runtime.Object, version string) ([]byte, error) {
	var w bytes.Buffer
	err := encoder(version).Encode(obj, &w)
	if err != nil {
		return nil, fmt.Errorf("error encoding %T: %v", obj, err)
	}
	return w.Bytes(), nil
}

func ParseVersionedYaml(data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	return decoder().Decode(data, nil, nil)
}
