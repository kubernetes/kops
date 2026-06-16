/*
Copyright 2026 The Kubernetes Authors.

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

package applyset

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// TestAllAppliedTreatsDeferredAsApplied covers the e2e-kops-ai-conformance failure:
// Cilium's GatewayClass cannot be applied until the Gateway API CRDs are installed,
// but that should not pin the kops-channels readiness probe to NotReady forever.
func TestAllAppliedTreatsDeferredAsApplied(t *testing.T) {
	r := &ApplyResults{total: 2}
	gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GatewayClass"}
	nn := types.NamespacedName{Name: "cilium"}

	r.applySuccess(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, types.NamespacedName{Namespace: "kube-system", Name: "cilium-config"})
	r.reportHealth(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, types.NamespacedName{Namespace: "kube-system", Name: "cilium-config"}, true)
	r.applyDeferred(gvk, nn, &meta.NoKindMatchError{GroupKind: gvk.GroupKind(), SearchedVersions: []string{gvk.Version}})

	if !r.AllApplied() {
		t.Errorf("AllApplied() = false, want true when only failure is a deferred missing CRD; got %#v", r)
	}
	if !r.AllHealthy() {
		t.Errorf("AllHealthy() = false, want true; got %#v", r)
	}
}

func TestApplyDeferredDetectsNoMatchError(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GatewayClass"}
	noMatch := &meta.NoKindMatchError{GroupKind: gvk.GroupKind(), SearchedVersions: []string{gvk.Version}}
	// The applyset.go change wraps the error with fmt.Errorf("error getting rest mapping for %v: %w", ...).
	// Make sure the wrapping does not defeat meta.IsNoMatchError.
	wrapped := errors.Join(errors.New("error getting rest mapping for v1, Kind=GatewayClass: "), noMatch)
	if !meta.IsNoMatchError(wrapped) {
		t.Fatalf("meta.IsNoMatchError(wrapped) = false, want true")
	}
}
