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
	// apimachinery's fuzzer.FuzzerFor takes a math/rand (v1) Source, so v1 it is.
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/api/apitesting/roundtrip"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kops/pkg/apis/kops"
)

// testScheme is shared: Install registers several hundred conversion functions per
// call, and the resulting scheme is read only once registration completes.
var testScheme = sync.OnceValue(func() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Install(scheme)

	return scheme
})

var internalGV = schema.GroupVersion{Group: kops.GroupName, Version: runtime.APIVersionInternal}

// externalVersion pairs an API version with the fuzzer functions describing what it
// cannot represent.
type externalVersion struct {
	version string
	funcs   fuzzer.FuzzerFuncs
}

// externalVersions lists every external version and the fuzzer functions describing
// what it loses. Adding an API version means adding an entry here.
var externalVersions = []externalVersion{
	{version: "v1alpha2", funcs: v1alpha2FuzzerFuncs},
	{version: "v1alpha3", funcs: v1alpha3FuzzerFuncs},
}

// TestRoundTripTypes fuzzes an internal object, converts it to an external version,
// serializes it, and checks that reading it back produces the same internal object.
//
// Each version is exercised with its own fuzzer functions, because the two versions
// lose different fields: v1alpha3 drops the settings kOps has removed, while v1alpha2
// lacks a home for some newer ones and flattens others into legacy kube-apiserver
// flags. What those functions clear in fuzzer_test.go is the authoritative list of what
// is silently dropped on write, with the reason recorded on each entry.
func TestRoundTripTypes(t *testing.T) {
	scheme := testScheme()
	codecFactory := runtimeserializer.NewCodecFactory(scheme)
	kinds := internalKindNames(t, scheme)

	for _, ver := range externalVersions {
		t.Run(ver.version, func(t *testing.T) {
			seed := fuzzSeed(t)
			t.Logf("fuzzing with seed %d; reproduce with KOPS_FUZZ_SEED=%d", seed, seed)

			filler := fuzzer.FuzzerFor(
				fuzzer.MergeFuzzerFuncs(metafuzzer.Funcs, ver.funcs),
				rand.NewSource(seed),
				codecFactory,
			)

			skip := skipOtherVersions(ver.version, kinds)
			for _, kind := range kinds {
				roundtrip.RoundTripSpecificKindWithoutProtobuf(t, internalGV.WithKind(kind), scheme, codecFactory, filler, skip)
			}
		})
	}
}

// TestFuzzerFuncsCoverAllVersions fails if an external version reaches the scheme
// without fuzzer functions describing what it loses. Such a version would otherwise be
// round tripped under another version's fuzzer, reporting failures against the wrong
// subtest.
func TestFuzzerFuncsCoverAllVersions(t *testing.T) {
	var missing []string
	for _, gv := range testScheme().PrioritizedVersionsForGroup(kops.GroupName) {
		if !slices.ContainsFunc(externalVersions, func(v externalVersion) bool { return v.version == gv.Version }) {
			missing = append(missing, gv.String())
		}
	}
	if len(missing) > 0 {
		t.Errorf("no fuzzer funcs registered for %v; add them to externalVersions", missing)
	}
}

// fuzzSeed returns the seed to fuzz with. A lossy field combination usually only shows
// up on a fraction of seeds, so a CI failure is only actionable if the seed it used can
// be fed back in.
func fuzzSeed(t *testing.T) int64 {
	t.Helper()

	s := os.Getenv("KOPS_FUZZ_SEED")
	if s == "" {
		return rand.Int63()
	}

	seed, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		t.Fatalf("parsing KOPS_FUZZ_SEED %q: %v", s, err)
	}

	return seed
}

// skipOtherVersions returns the non-round-trippable set covering every external version
// but keep. roundTripToAllExternalVersions walks every external version of a kind, and
// each version only round trips under the fuzzer functions written for it.
func skipOtherVersions(keep string, kinds []string) map[schema.GroupVersionKind]bool {
	skip := make(map[schema.GroupVersionKind]bool, (len(externalVersions)-1)*len(kinds))
	for _, ver := range externalVersions {
		if ver.version == keep {
			continue
		}
		for _, kind := range kinds {
			skip[schema.GroupVersionKind{Group: kops.GroupName, Version: ver.version, Kind: kind}] = true
		}
	}

	return skip
}

// internalKindNames returns the internal kinds of the kOps API group, so that a newly
// added type is covered without having to be listed here.
func internalKindNames(t *testing.T, scheme *runtime.Scheme) []string {
	t.Helper()

	var kinds []string
	for kind, goType := range scheme.KnownTypes(internalGV) {
		// metav1.AddToGroupVersion registers its own types into the group; they are
		// not ours to round trip.
		if !strings.HasPrefix(goType.PkgPath(), "k8s.io/kops/") {
			continue
		}
		kinds = append(kinds, kind)
	}
	if len(kinds) == 0 {
		t.Fatalf("no kinds registered for %s", internalGV)
	}
	slices.Sort(kinds)

	return kinds
}
