/*
Copyright 2022 The Kubernetes Authors.

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

package channels

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/applylib/applyset"
)

type Harness struct {
	*testing.T
	Ctx           context.Context
	dynamicClient dynamic.Interface
	kubeClient    kubernetes.Interface
	restMapper    *restmapper.DeferredDiscoveryRESTMapper
}

func NewHarness(t *testing.T) *Harness {
	ctx := context.TODO()

	target := os.Getenv("E2E_KUBE_TARGET")
	if target == "" {
		t.Skip("E2E_KUBE_TARGET not set, skipping")
	}

	h := &Harness{T: t, Ctx: ctx}

	var restConfig *rest.Config
	if target == "kubeconfig" {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		rc, err := kubeconfig.ClientConfig()
		if err != nil {
			t.Fatalf("failed to get kubeconfig: %v", err)
		}
		restConfig = rc
	} else {
		// TODO: Support other E2E_KUBE_TARGET values, like kubeenv?
		t.Fatalf("E2E_KUBE_TARGET=%q not recognized", target)
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		t.Fatalf("error from kubernetes.NewForConfig: %v", err)
	}
	h.kubeClient = kubeClient

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		t.Fatalf("failed to build dynamic client: %v", err)
	}
	h.dynamicClient = dynamicClient
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		t.Fatalf("failed to build discovery client: %v", err)
	}
	h.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	return h
}

func (h *Harness) RESTMapper() *restmapper.DeferredDiscoveryRESTMapper {
	return h.restMapper
}

func (h *Harness) DynamicClient() dynamic.Interface {
	return h.dynamicClient
}

func (h *Harness) EphemeralNamespace() *corev1.Namespace {
	baseName := h.T.Name()

	nsName := baseName + "-" + strconv.FormatInt(time.Now().UnixNano(), 16)
	nsName = strings.ToLower(nsName) // We might need more sanitization in future

	ns := &corev1.Namespace{}
	ns.Name = nsName
	ns.Labels = map[string]string{
		"e2e-test": baseName,
	}

	if _, err := h.kubeClient.CoreV1().Namespaces().Create(h.Ctx, ns, metav1.CreateOptions{}); err != nil {
		h.Fatalf("failed to create namespace: %v", err)
	}

	h.T.Cleanup(func() {
		if err := h.kubeClient.CoreV1().Namespaces().Delete(h.Ctx, ns.Name, metav1.DeleteOptions{}); err != nil {
			h.Errorf("failed to delete namespace: %v", err)
		}
	})

	return ns
}

func (h *Harness) MustGetPDB(id types.NamespacedName) *unstructured.Unstructured {
	gvr := schema.GroupVersionResource{
		Group:    "policy",
		Version:  "v1beta1",
		Resource: "poddisruptionbudgets",
	}
	pdb, err := h.dynamicClient.Resource(gvr).Namespace(id.Namespace).Get(h.Ctx, id.Name, metav1.GetOptions{})
	if err != nil {
		h.Fatalf("failed to get object: %v", err)
	}

	pdbJSON, err := pdb.MarshalJSON()
	if err != nil {
		h.Fatalf("failed to get JSON for pdb: %v", err)
	}
	h.Logf("PDB was %v", string(pdbJSON))

	return pdb
}

const oldPDB = `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    addon.kops.k8s.io/name: coredns.addons.k8s.io
    app.kubernetes.io/managed-by: kops
    k8s-addon: coredns.addons.k8s.io
  name: kube-dns
  namespace: kube-system
spec:
  minAvailable: 1
  selector:
    matchLabels:
      k8s-app: kube-dns
`

const newPDB = `
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  labels:
    addon.kops.k8s.io/name: coredns.addons.k8s.io
    app.kubernetes.io/managed-by: kops
    k8s-addon: coredns.addons.k8s.io
  name: kube-dns
  namespace: kube-system
spec:
  maxUnavailable: 50%
  selector:
    matchLabels:
      k8s-app: kube-dns
`

// TestFieldManagerChange verifies that we can move correctly between our apply-strategies
func TestFieldManagerChange(t *testing.T) {
	h := NewHarness(t)

	legacyKubectlApplier := &legacyClientSideKubectlApplier{}
	kubectlApplier := &KubectlApplier{}
	ssaApplier := &ClientApplier{
		Client:       h.DynamicClient(),
		RESTMapper:   h.RESTMapper(),
		IgnoreHealth: true,
	}

	grid := []struct {
		Name       string
		OldApplier Applier
		NewApplier Applier
	}{
		{
			Name:       "legacy-kubectl-to-kubectl",
			OldApplier: legacyKubectlApplier,
			NewApplier: kubectlApplier,
		},
		{
			Name:       "legacy-kubectl-to-ssa",
			OldApplier: legacyKubectlApplier,
			NewApplier: ssaApplier,
		},
		{
			Name:       "kubectl-to-ssa",
			OldApplier: kubectlApplier,
			NewApplier: ssaApplier,
		},
		{
			Name:       "ssa-to-ssa",
			OldApplier: ssaApplier,
			NewApplier: ssaApplier,
		},
	}

	for _, g := range grid {
		g := g
		t.Run(g.Name, func(t *testing.T) {
			ns := h.EphemeralNamespace()

			oldYAML := strings.ReplaceAll(oldPDB, "namespace: kube-system", "namespace: "+ns.Name)
			newYAML := strings.ReplaceAll(newPDB, "namespace: kube-system", "namespace: "+ns.Name)

			t.Logf("applying 'old' configuration")
			if err := g.OldApplier.Apply(h.Ctx, []byte(oldYAML)); err != nil {
				t.Fatalf("failed to apply 'old' configuration: %v", err)
			}
			t.Logf("applying 'new' configuration")
			if err := g.NewApplier.Apply(h.Ctx, []byte(newYAML)); err != nil {
				t.Fatalf("failed to apply 'new' configuration: %v", err)
			}

			pdb := h.MustGetPDB(types.NamespacedName{Namespace: ns.Name, Name: "kube-dns"})

			maxUnavailable, found, err := unstructured.NestedString(pdb.Object, "spec", "maxUnavailable")
			if err != nil {
				t.Fatalf("error getting spec.maxUnavailable: %v", err)
			}
			if !found {
				t.Fatalf("did not find spec.maxUnavailable")
			}
			if got, want := maxUnavailable, "50%"; got != want {
				t.Errorf("spec.maxUnavailable was not as expected, got %q, want %q", got, want)
			}

			minAvailable, found, err := unstructured.NestedInt64(pdb.Object, "spec", "minAvailable")
			if err != nil {
				t.Fatalf("error getting spec.minAvailable: %v", err)
			}
			if found {
				t.Fatalf("got spec.minAvailable, was expecting it to be cleared: %v", minAvailable)
			}

			if _, isSSA := g.NewApplier.(*ClientApplier); isSSA {
				fieldManagers := gatherFieldManagers(pdb)
				wantFieldManagers := []string{
					"kops:Apply",
					"kube-controller-manager:Update",
				}
				if diff := cmp.Diff(fieldManagers, wantFieldManagers); diff != "" {
					t.Errorf("unexpected field managers after server-side apply; got %v, want %v", fieldManagers, wantFieldManagers)
				}
			}
		})
	}
}

// TestMergeMultipleFieldManagers verifies that we combine field managers with identical keys
func TestMergeMultipleFieldManagers(t *testing.T) {
	h := NewHarness(t)

	legacyKubectlApplierApply := &legacyClientSideKubectlApplier{}
	ssaApplier := &ClientApplier{
		Client:       h.DynamicClient(),
		RESTMapper:   h.RESTMapper(),
		IgnoreHealth: true,
	}

	ns := h.EphemeralNamespace()

	nn := types.NamespacedName{Namespace: ns.Name, Name: "kube-dns"}

	client := applyset.NewUnstructuredClient(applyset.Options{
		RESTMapper: h.RESTMapper(),
		Client:     h.DynamicClient(),
	})

	oldYAML := strings.ReplaceAll(oldPDB, "namespace: kube-system", "namespace: "+ns.Name)
	newYAML := strings.ReplaceAll(newPDB, "namespace: kube-system", "namespace: "+ns.Name)

	t.Logf("applying 'old' configuration")
	if err := legacyKubectlApplierApply.Apply(h.Ctx, []byte(oldYAML)); err != nil {
		t.Fatalf("failed to apply 'old' configuration: %v", err)
	}

	{
		pdb := h.MustGetPDB(nn)
		pdb.SetAnnotations(map[string]string{"foo": "bar"})
		if _, err := client.Update(h.Ctx, pdb, metav1.UpdateOptions{FieldManager: "kubectl-edit"}); err != nil {
			t.Fatalf("failed to update annotations: %v", err)
		}
	}

	{
		got := h.MustGetPDB(nn)

		fieldManagers := gatherFieldManagers(got)
		wantFieldManagers := []string{
			"kube-controller-manager:Update",
			"kubectl-client-side-apply:Update",
			"kubectl-edit:Update",
		}
		if diff := cmp.Diff(fieldManagers, wantFieldManagers); diff != "" {
			t.Errorf("unexpected field managers after two client-side applies; got %v, want %v", fieldManagers, wantFieldManagers)
		}
	}

	t.Logf("applying 'new' configuration with SSA applier")
	if err := ssaApplier.Apply(h.Ctx, []byte(newYAML)); err != nil {
		t.Fatalf("failed to apply 'new' configuration with SSA: %v", err)
	}

	{
		got := h.MustGetPDB(nn)
		fieldManagers := gatherFieldManagers(got)
		wantFieldManagers := []string{
			"kops:Apply",
			"kube-controller-manager:Update",
		}
		if diff := cmp.Diff(fieldManagers, wantFieldManagers); diff != "" {
			t.Errorf("unexpected field managers after server-side apply; got %v, want %v", fieldManagers, wantFieldManagers)
		}
	}

}

// TestManagedFieldsMigrator verifies the ManagedFieldsMigrator directly (without an SSA)
func TestManagedFieldsMigrator(t *testing.T) {
	h := NewHarness(t)

	legacyKubectlApplierApply := &legacyClientSideKubectlApplier{}

	ns := h.EphemeralNamespace()

	nn := types.NamespacedName{Namespace: ns.Name, Name: "kube-dns"}

	client := applyset.NewUnstructuredClient(applyset.Options{
		RESTMapper: h.RESTMapper(),
		Client:     h.DynamicClient(),
	})

	oldYAML := strings.ReplaceAll(oldPDB, "namespace: kube-system", "namespace: "+ns.Name)

	t.Logf("applying 'old' configuration")
	if err := legacyKubectlApplierApply.Apply(h.Ctx, []byte(oldYAML)); err != nil {
		t.Fatalf("failed to apply 'old' configuration: %v", err)
	}

	{
		pdb := h.MustGetPDB(nn)
		pdb.SetAnnotations(map[string]string{"foo": "bar"})
		if _, err := client.Update(h.Ctx, pdb, metav1.UpdateOptions{FieldManager: "kubectl-edit"}); err != nil {
			t.Fatalf("failed to update annotations: %v", err)
		}
	}

	{
		got := h.MustGetPDB(nn)

		fieldManagers := gatherFieldManagers(got)
		wantFieldManagers := []string{
			"kube-controller-manager:Update",
			"kubectl-client-side-apply:Update",
			"kubectl-edit:Update",
		}
		if diff := cmp.Diff(fieldManagers, wantFieldManagers); diff != "" {
			t.Errorf("unexpected field managers after two client-side applies; got %v, want %v", fieldManagers, wantFieldManagers)
		}
	}

	t.Logf("updating field manager directly")
	m := &applyset.ManagedFieldsMigrator{
		NewManager: "kops",
		Client:     client,
	}
	pdb := h.MustGetPDB(nn)
	if err := m.Migrate(h.Ctx, pdb); err != nil {
		t.Fatalf("failed to apply 'new' configuration with SSA: %v", err)
	}

	{
		got := h.MustGetPDB(nn)
		fieldManagers := gatherFieldManagers(got)
		wantFieldManagers := []string{
			"kops:Apply",
			"kube-controller-manager:Update",
		}
		if diff := cmp.Diff(fieldManagers, wantFieldManagers); diff != "" {
			t.Errorf("unexpected field managers after migration; got %v, want %v", fieldManagers, wantFieldManagers)
		}
	}

	{
		pdb := h.MustGetPDB(nn)
		entry := getManagedFieldEntry(pdb, "kops", metav1.ManagedFieldsOperationApply, "")
		if entry == nil {
			t.Fatalf("could not find managed field set")
		}
		got := string(entry.FieldsV1.Raw)
		// Owns annotations and labels
		want := `{"f:metadata":{"f:annotations":{".":{},"f:foo":{}},"f:labels":{".":{},"f:addon.kops.k8s.io/name":{},"f:app.kubernetes.io/managed-by":{},"f:k8s-addon":{}}},"f:spec":{"f:minAvailable":{},"f:selector":{}}}`
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("unexpected field managers after migration; got %v, want %v.  diff=%v", got, want, diff)
		}
	}
}

// gatherFieldManagers returns a sorted slice of Manager:Operation for each field manager
func gatherFieldManagers(u *unstructured.Unstructured) []string {
	var result []string
	for _, manager := range u.GetManagedFields() {
		s := manager.Manager + ":" + string(manager.Operation)
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

// getManagedFieldEntry returns the matching ManagedFields, or nil if no match found.
func getManagedFieldEntry(obj *unstructured.Unstructured, manager string, operation metav1.ManagedFieldsOperationType, subresource string) *metav1.ManagedFieldsEntry {
	managedFields := obj.GetManagedFields()
	for i := range managedFields {
		f := &managedFields[i]
		if f.Manager != manager || f.Subresource != subresource || f.Operation != operation {
			continue
		}
		return f
	}
	return nil
}

// legacyClientSideKubectlApplier is a copy of the original kubectl-backed applier,
// used here to check we can update from older versions.
type legacyClientSideKubectlApplier struct {
}

// Apply calls kubectl apply to apply the manifest.
func (a *legacyClientSideKubectlApplier) Apply(ctx context.Context, data []byte) error {
	// We copy the manifest to a temp file because it is likely e.g. an s3 URL, which kubectl can't read
	tmpDir, err := os.MkdirTemp("", "channel")
	if err != nil {
		return fmt.Errorf("error creating temp dir: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			klog.Warningf("error deleting temp dir %q: %v", tmpDir, err)
		}
	}()

	localManifestFile := path.Join(tmpDir, "manifest.yaml")
	if err := os.WriteFile(localManifestFile, data, 0o600); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	{
		_, err := execKubectl(ctx, "apply", "-f", localManifestFile)
		if err != nil {
			klog.Errorf("failed to apply the manifest: %v", err)
		}
	}

	return nil
}
