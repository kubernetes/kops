package kops

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

func decoder() runtime.Decoder {
	// TODO: Cache?
	// Codecs provides access to encoding and decoding for the scheme
	codecs := k8sapi.Codecs
	codec := codecs.UniversalDecoder(SchemeGroupVersion)
	return codec
}

func encoder() runtime.Encoder {
	// TODO: Which is better way to build yaml?
	//yaml := json.NewYAMLSerializer(json.DefaultMetaFactory, k8sapi.Scheme, k8sapi.Scheme)

	// TODO: Cache?
	yaml, ok := runtime.SerializerInfoForMediaType(k8sapi.Codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		glog.Fatalf("no YAML serializer registered")
	}
	return k8sapi.Codecs.EncoderForVersion(yaml.Serializer, preferredAPIVersion())
}

func preferredAPIVersion() unversioned.GroupVersion {
	// Avoid circular dependency
	// return v1alpha1.SchemeGroupVersion
	return unversioned.GroupVersion{Group: GroupName, Version: "v1alpha1"}
}

// ToVersionedYaml encodes the object to YAML, in our preferred API version
func ToVersionedYaml(obj runtime.Object) ([]byte, error) {
	var w bytes.Buffer
	err := encoder().Encode(obj, &w)
	if err != nil {
		return nil, fmt.Errorf("error encoding &T: %v", obj, err)
	}
	return w.Bytes(), nil
}

func ParseVersionedYaml(data []byte) (runtime.Object, *unversioned.GroupVersionKind, error) {
	return decoder().Decode(data, nil, nil)
}
