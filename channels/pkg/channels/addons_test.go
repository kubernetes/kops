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

package channels

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	fakecertmanager "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubernetes "k8s.io/client-go/kubernetes/fake"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func TestParseAddons(t *testing.T) {
	location := mustParseURL(t, "https://example.com/addons/addon.yaml")

	data := []byte(`
kind: Addons
metadata:
  name: example
spec:
  addons:
  - name: example.addons.k8s.io
    manifest: example.yaml
    manifestHash: abc123
`)

	addons, err := ParseAddons("channel", location, data)
	if err != nil {
		t.Fatalf("ParseAddons returned error: %v", err)
	}

	if addons.APIObject.ObjectMeta.Name != "example" {
		t.Fatalf("expected Addons object name %q, got %q", "example", addons.APIObject.ObjectMeta.Name)
	}
	if len(addons.APIObject.Spec.Addons) != 1 {
		t.Fatalf("expected one addon, got %d", len(addons.APIObject.Spec.Addons))
	}
	addon := addons.APIObject.Spec.Addons[0]
	if addon.Name == nil || *addon.Name != "example.addons.k8s.io" {
		t.Fatalf("expected addon name %q, got %v", "example.addons.k8s.io", addon.Name)
	}
}

// A kOps Addons index may contain more than one YAML document; it must still be
// parsed as an index (from the first document) rather than wrapped as a raw manifest.
func TestParseAddonsMultiDocIndex(t *testing.T) {
	location := mustParseURL(t, "https://example.com/addons/addon.yaml")

	data := []byte(`
kind: Addons
metadata:
  name: example
spec:
  addons:
  - name: example.addons.k8s.io
    manifest: example.yaml
    manifestHash: abc123
---
kind: Addons
metadata:
  name: other
spec:
  addons:
  - name: other.addons.k8s.io
    manifest: other.yaml
    manifestHash: def456
`)

	addons, err := ParseAddons("channel", location, data)
	if err != nil {
		t.Fatalf("ParseAddons returned error: %v", err)
	}

	if addons.APIObject.ObjectMeta.Name != "example" {
		t.Fatalf("expected Addons index to be parsed (name %q), got %q", "example", addons.APIObject.ObjectMeta.Name)
	}
	if len(addons.APIObject.Spec.Addons) != 1 {
		t.Fatalf("expected one addon, got %d", len(addons.APIObject.Spec.Addons))
	}
	addon := addons.APIObject.Spec.Addons[0]
	if addon.Name == nil || *addon.Name != "example.addons.k8s.io" {
		t.Fatalf("expected addon name %q, got %v", "example.addons.k8s.io", addon.Name)
	}
}

func TestParseAddonsWrapsDirectManifest(t *testing.T) {
	location := mustParseURL(t, "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml")
	data := []byte(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: gateways.gateway.networking.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller
  namespace: gateway-system
`)

	addons, err := ParseAddons("channel", location, data)
	if err != nil {
		t.Fatalf("ParseAddons returned error: %v", err)
	}

	expectedHash, err := utils.HashString(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatalf("HashString returned error: %v", err)
	}
	expectedLocationHash, err := utils.HashString(location.String())
	if err != nil {
		t.Fatalf("HashString returned error: %v", err)
	}
	expectedName := "manifest-" + expectedLocationHash[:12]

	if addons.APIObject.ObjectMeta.Name != expectedName {
		t.Fatalf("expected Addons object name %q, got %q", expectedName, addons.APIObject.ObjectMeta.Name)
	}
	if len(addons.APIObject.Spec.Addons) != 1 {
		t.Fatalf("expected one addon, got %d", len(addons.APIObject.Spec.Addons))
	}
	addon := addons.APIObject.Spec.Addons[0]
	if addon.Name == nil || *addon.Name != expectedName {
		t.Fatalf("expected addon name %q, got %v", expectedName, addon.Name)
	}
	if addon.Manifest == nil || *addon.Manifest != location.String() {
		t.Fatalf("expected manifest %q, got %v", location.String(), addon.Manifest)
	}
	if addon.ManifestHash != expectedHash {
		t.Fatalf("expected manifest hash %q, got %q", expectedHash, addon.ManifestHash)
	}
}

func TestParseAddonsDirectManifestNameUsesLocationHash(t *testing.T) {
	location := mustParseURL(t, "https://example.com/addons/direct.yaml")

	data1 := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: version-1
`)
	data2 := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: version-2
`)

	addons1, err := ParseAddons("channel", location, data1)
	if err != nil {
		t.Fatalf("ParseAddons returned error: %v", err)
	}
	addons2, err := ParseAddons("channel", location, data2)
	if err != nil {
		t.Fatalf("ParseAddons returned error: %v", err)
	}

	addon1 := addons1.APIObject.Spec.Addons[0]
	addon2 := addons2.APIObject.Spec.Addons[0]
	if addon1.Name == nil || addon2.Name == nil {
		t.Fatalf("expected addon names, got %v and %v", addon1.Name, addon2.Name)
	}
	if *addon1.Name != *addon2.Name {
		t.Fatalf("expected stable addon name, got %q and %q", *addon1.Name, *addon2.Name)
	}
	if addon1.ManifestHash == addon2.ManifestHash {
		t.Fatalf("expected content hash to change, got %q", addon1.ManifestHash)
	}
}

func TestParseAddonsRejectsInvalidYAML(t *testing.T) {
	location := mustParseURL(t, "https://example.com/addons/not-yaml.yaml")

	_, err := ParseAddons("channel", location, []byte("not: [valid"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "error parsing addons or manifest") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

// Content with no objects (empty, whitespace, or comment-only) is a no-op, not an error.
func TestParseAddonsEmpty(t *testing.T) {
	location := mustParseURL(t, "https://example.com/addons/addon.yaml")

	for name, data := range map[string][]byte{
		"empty":        []byte(""),
		"whitespace":   []byte("\n  \n"),
		"comment only": []byte("# nothing to see here\n"),
	} {
		t.Run(name, func(t *testing.T) {
			addons, err := ParseAddons("channel", location, data)
			if err != nil {
				t.Fatalf("ParseAddons returned error: %v", err)
			}
			if len(addons.APIObject.Spec.Addons) != 0 {
				t.Fatalf("expected no addons, got %d", len(addons.APIObject.Spec.Addons))
			}
		})
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", rawURL, err)
	}
	return u
}

func Test_Filtering(t *testing.T) {
	grid := []struct {
		Input             api.AddonSpec
		KubernetesVersion string
		Expected          bool
	}{
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.6.0",
			},
			KubernetesVersion: "1.6.0",
			Expected:          true,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: "<1.6.0",
			},
			KubernetesVersion: "1.6.0",
			Expected:          false,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.6.0",
			},
			KubernetesVersion: "1.5.9",
			Expected:          false,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.4.0 <1.6.0",
			},
			KubernetesVersion: "1.5.9",
			Expected:          true,
		},
		{
			Input: api.AddonSpec{
				KubernetesVersion: ">=1.4.0 <1.6.0",
			},
			KubernetesVersion: "1.6.0",
			Expected:          false,
		},
	}
	for _, g := range grid {
		k8sVersion := semver.MustParse(g.KubernetesVersion)
		addon := &Addon{
			Spec: &g.Input,
		}
		actual := addon.matches(k8sVersion)
		if actual != g.Expected {
			t.Errorf("unexpected result from %v, %s.  got %v", g.Input.KubernetesVersion, g.KubernetesVersion, actual)
		}
	}
}

func Test_Replacement(t *testing.T) {
	hash1 := "3544de6578b2b582c0323b15b7b05a28c60b9430"
	hash2 := "ea9e79bf29adda450446487d65a8fc6b3fdf8c2b"

	grid := []struct {
		Old      *ChannelVersion
		New      *ChannelVersion
		Replaces bool
	}{
		// Test ManifestHash Changes
		{
			Old:      &ChannelVersion{Id: "a", ManifestHash: hash1},
			New:      &ChannelVersion{Id: "a", ManifestHash: hash1},
			Replaces: false,
		},
		{
			Old:      &ChannelVersion{Id: "a", ManifestHash: ""},
			New:      &ChannelVersion{Id: "a", ManifestHash: hash1},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Id: "a", ManifestHash: hash1},
			New:      &ChannelVersion{Id: "a", ManifestHash: hash2},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Id: "a", ManifestHash: hash1},
			New:      &ChannelVersion{Id: "a", ManifestHash: hash1, SystemGeneration: 1},
			Replaces: true,
		},
		{
			Old:      &ChannelVersion{Id: "a", ManifestHash: hash1, SystemGeneration: 1},
			New:      &ChannelVersion{Id: "a", ManifestHash: hash1},
			Replaces: false,
		},
		{
			Old:      &ChannelVersion{Id: "a", ManifestHash: hash1, SystemGeneration: 1},
			New:      &ChannelVersion{Id: "a", ManifestHash: hash1, SystemGeneration: 1},
			Replaces: false,
		},
	}
	for _, g := range grid {
		actual := g.New.replaces(t.Name(), g.Old)
		if actual != g.Replaces {
			t.Errorf("unexpected result from %v -> %v, expect %t.  actual %v", g.Old, g.New, g.Replaces, actual)
		}
	}
}

func Test_GetRequiredUpdates(t *testing.T) {
	ctx := context.Background()
	kubeSystem := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}
	fakek8s := fakekubernetes.NewClientset(kubeSystem)
	fakecm := fakecertmanager.NewSimpleClientset()
	addon := &Addon{
		Name: "test",
		Spec: &api.AddonSpec{
			Name:     new("test"),
			NeedsPKI: true,
		},
	}
	addonUpdate, err := addon.GetRequiredUpdates(ctx, fakek8s, fakecm, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if addonUpdate == nil {
		t.Fatal("expected addon update, got nil")
	}
	if !addonUpdate.InstallPKI {
		t.Errorf("expected addon to require install")
	}
}

func Test_NeedsRollingUpdate(t *testing.T) {
	grid := []struct {
		newAddon            *Addon
		originalAnnotations map[string]string
		updateRequired      bool
		installRequired     bool
		expectedNodeUpdates int
	}{
		{
			newAddon: &Addon{
				Name: "test",
				Spec: &api.AddonSpec{
					Name:               new("test"),
					ManifestHash:       "originalHash",
					NeedsRollingUpdate: api.NeedsRollingUpdateAll,
				},
			},
		},
		{
			newAddon: &Addon{
				Name: "test",
				Spec: &api.AddonSpec{
					Name:               new("test"),
					ManifestHash:       "newHash",
					NeedsRollingUpdate: api.NeedsRollingUpdateAll,
				},
			},
			updateRequired:      true,
			expectedNodeUpdates: 2,
		},
		{
			newAddon: &Addon{
				Name: "test",
				Spec: &api.AddonSpec{
					Name:               new("test"),
					ManifestHash:       "newHash",
					NeedsRollingUpdate: api.NeedsRollingUpdateWorkers,
				},
			},
			updateRequired:      true,
			expectedNodeUpdates: 1,
		},
		{
			newAddon: &Addon{
				Name: "test",
				Spec: &api.AddonSpec{
					Name:               new("test"),
					ManifestHash:       "newHash",
					NeedsRollingUpdate: api.NeedsRollingUpdateControlPlane,
				},
			},
			updateRequired:      true,
			expectedNodeUpdates: 1,
		},
		{
			newAddon: &Addon{
				Name: "test",
				Spec: &api.AddonSpec{
					Name:               new("test"),
					ManifestHash:       "newHash",
					NeedsRollingUpdate: api.NeedsRollingUpdateAll,
				},
			},
			originalAnnotations: map[string]string{
				"addons.k8s.io/placeholder": "{\"manifestHash\":\"originalHash\"}",
			},
			installRequired:     true,
			expectedNodeUpdates: 0,
		},
	}

	for _, g := range grid {
		ctx := context.Background()

		annotations := map[string]string{
			"addons.k8s.io/test": "{\"manifestHash\":\"originalHash\",\"systemGeneration\": 1}",
		}
		if len(g.originalAnnotations) > 0 {
			annotations = g.originalAnnotations
		}

		kubeSystem := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "kube-system",
				Annotations: annotations,
			},
		}

		objects := []runtime.Object{
			kubeSystem,
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cp",
					Labels: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
					Labels: map[string]string{
						"node-role.kubernetes.io/node": "",
					},
				},
			},
		}

		existingChannels := FindChannelVersions(kubeSystem)

		fakek8s := fakekubernetes.NewClientset(objects...)
		fakecm := fakecertmanager.NewSimpleClientset()

		addon := g.newAddon
		required, err := addon.GetRequiredUpdates(ctx, fakek8s, fakecm, existingChannels[addon.Name])
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !g.updateRequired && !g.installRequired {
			if required == nil {
				continue
			} else {
				t.Fatalf("did not expect update, but required was not nil")
			}
		}

		if required == nil {
			t.Fatalf("expected required update, got nil")
		}

		if required.NewVersion == nil {
			t.Errorf("updating or installing addon, but NewVersion was nil")
		}

		if required.ExistingVersion != nil {
			if g.installRequired {
				t.Errorf("new install of addon, but ExistingVersion was not nil")
			}
		} else {
			if g.updateRequired {
				t.Errorf("update of addon, but ExistingVersion was nil")
			}
		}

		if err := addon.AddNeedsUpdateLabel(ctx, fakek8s, required); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		nodes, _ := fakek8s.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		nodeUpdates := 0

		for _, node := range nodes.Items {
			if _, exists := node.Annotations["kops.k8s.io/needs-update"]; exists {
				nodeUpdates++
			}
		}

		if nodeUpdates != g.expectedNodeUpdates {
			t.Errorf("expected %d node updates, but got %d", g.expectedNodeUpdates, nodeUpdates)
		}

	}
}

func Test_InstallPKI(t *testing.T) {
	ctx := context.Background()
	kubeSystem := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}
	fakek8s := fakekubernetes.NewClientset(kubeSystem)
	fakecm := fakecertmanager.NewSimpleClientset()
	addon := &Addon{
		Name: "test",
		Spec: &api.AddonSpec{
			Name:     new("test"),
			NeedsPKI: true,
		},
	}
	err := addon.installPKI(ctx, fakek8s, fakecm)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = fakek8s.CoreV1().Secrets("kube-system").Get(ctx, "test-ca", metav1.GetOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Two consecutive calls should work since multiple CP nodes can update at the same time
	err = addon.installPKI(ctx, fakek8s, fakecm)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = fakecm.CertmanagerV1().Issuers("kube-system").Get(ctx, "test", metav1.GetOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
