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

package model

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"

	_ "k8s.io/kubernetes/pkg/api/install"
)

func encoder() runtime.Encoder {
	// TODO: Which is better way to build yaml?
	//yaml := json.NewYAMLSerializer(json.DefaultMetaFactory, k8sapi.Scheme, k8sapi.Scheme)

	// TODO: Cache?
	yaml, ok := runtime.SerializerInfoForMediaType(api.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		glog.Fatalf("no YAML serializer registered")
	}
	gv := v1.SchemeGroupVersion
	return api.Codecs.EncoderForVersion(yaml.Serializer, gv)
}

// ToVersionedYamlWithVersion encodes the object to YAML
func ToVersionedYaml(obj runtime.Object) ([]byte, error) {
	var w bytes.Buffer
	err := encoder().Encode(obj, &w)
	if err != nil {
		return nil, fmt.Errorf("error encoding %T: %v", obj, err)
	}
	return w.Bytes(), nil
}
