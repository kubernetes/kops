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
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/install"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"os"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)

var Registry = registered.NewOrDie(os.Getenv("KUBE_API_VERSIONS"))
var GroupFactoryRegistry = make(announced.APIGroupFactoryRegistry)

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	install.Install(GroupFactoryRegistry, Registry, Scheme)
}

func encoder(gv runtime.GroupVersioner) runtime.Encoder {
	yaml, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		glog.Fatalf("no YAML serializer registered")
	}
	return Codecs.EncoderForVersion(yaml.Serializer, gv)
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
	err := encoder(version).Encode(obj, &w)
	if err != nil {
		return nil, fmt.Errorf("error encoding %T: %v", obj, err)
	}
	return w.Bytes(), nil
}

func ParseVersionedYaml(data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	return decoder().Decode(data, nil, nil)
}
