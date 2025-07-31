/*
Copyright 2017 The Kubernetes Authors.

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

package install

import (
	"math/rand"
	"testing"

	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/api/apitesting/roundtrip"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
)

func TestRoundTripTypes(t *testing.T) {
	var fuzzingFuncs fuzzer.FuzzerFuncs

	nonRoundTrippableTypes := make(map[schema.GroupVersionKind]bool)

	// TODO: Eliminate this once all types round-trip
	for _, version := range []string{"v1alpha2", "v1alpha3"} {
		nonRoundTrippableTypes[schema.GroupVersionKind{Group: "kops.k8s.io", Version: version, Kind: "Cluster"}] = true
		nonRoundTrippableTypes[schema.GroupVersionKind{Group: "kops.k8s.io", Version: version, Kind: "ClusterList"}] = true
		nonRoundTrippableTypes[schema.GroupVersionKind{Group: "kops.k8s.io", Version: version, Kind: "InstanceGroup"}] = true
		nonRoundTrippableTypes[schema.GroupVersionKind{Group: "kops.k8s.io", Version: version, Kind: "InstanceGroupList"}] = true
	}

	if len(nonRoundTrippableTypes) == 0 {
		roundtrip.RoundTripTestForAPIGroup(t, Install, fuzzingFuncs)
	} else {
		scheme := runtime.NewScheme()
		Install(scheme)

		codecFactory := runtimeserializer.NewCodecFactory(scheme)
		f := fuzzer.FuzzerFor(
			fuzzer.MergeFuzzerFuncs(metafuzzer.Funcs, fuzzingFuncs),
			rand.NewSource(rand.Int63()),
			codecFactory,
		)
		roundtrip.RoundTripTypesWithoutProtobuf(t, scheme, codecFactory, f, nonRoundTrippableTypes)
	}
}
