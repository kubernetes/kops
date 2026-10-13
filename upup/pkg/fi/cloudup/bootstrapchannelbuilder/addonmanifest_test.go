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

package bootstrapchannelbuilder

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
)

type recordingAddonRenderer struct {
	calls    int
	rendered []byte
}

func (r *recordingAddonRenderer) RenderTemplate(name string, source []byte, tasks map[string]fi.CloudupTask) ([]byte, error) {
	r.calls++
	if r.rendered != nil {
		return r.rendered, nil
	}
	return source, nil
}

func (r *recordingAddonRenderer) CloudControllerConfigArgv() ([]string, error) {
	return nil, nil
}

func TestAddonManifestNormalizeSkipsRenderForRawSources(t *testing.T) {
	ctx, err := fi.NewCloudupContext(context.Background(), fi.DeletionProcessingModeDeleteIncludingDeferred, nil, nil, nil, nil, nil, nil, map[string]fi.CloudupTask{})
	if err != nil {
		t.Fatalf("building cloudup context: %v", err)
	}

	rawManifest := []byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  literal: \"{{ .Values.image\"\n")
	renderer := &recordingAddonRenderer{
		rendered: []byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  literal: rendered\n"),
	}
	addon := &AddonManifest{
		Name:          new("raw-addon"),
		Location:      new("addons/raw-addon.yaml"),
		source:        fi.NewBytesResource(rawManifest),
		skipRender:    true,
		addonRenderer: renderer,
		addonSpec:     testAddonSpec("raw-addon"),
		skipRemap:     true,
	}

	if err := addon.Normalize(ctx); err != nil {
		t.Fatalf("normalizing raw addon: %v", err)
	}
	if renderer.calls != 0 {
		t.Fatalf("renderer calls = %d, want 0", renderer.calls)
	}

	actual, err := fi.ResourceAsString(addon.Contents)
	if err != nil {
		t.Fatalf("reading addon contents: %v", err)
	}
	if actual != strings.TrimSpace(string(rawManifest)) {
		t.Fatalf("addon contents = %q, want %q", actual, strings.TrimSpace(string(rawManifest)))
	}
}

func TestAddonManifestNormalizeRendersTemplateSources(t *testing.T) {
	ctx, err := fi.NewCloudupContext(context.Background(), fi.DeletionProcessingModeDeleteIncludingDeferred, nil, nil, nil, nil, nil, nil, map[string]fi.CloudupTask{})
	if err != nil {
		t.Fatalf("building cloudup context: %v", err)
	}

	renderer := &recordingAddonRenderer{
		rendered: []byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  literal: rendered\n"),
	}
	addon := &AddonManifest{
		Name:          new("template-addon"),
		Location:      new("addons/template-addon.yaml"),
		source:        fi.NewBytesResource([]byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  literal: {{ .Value }}\n")),
		addonRenderer: renderer,
		addonSpec:     testAddonSpec("template-addon"),
		skipRemap:     true,
	}

	if err := addon.Normalize(ctx); err != nil {
		t.Fatalf("normalizing template addon: %v", err)
	}
	if renderer.calls != 1 {
		t.Fatalf("renderer calls = %d, want 1", renderer.calls)
	}

	actual, err := fi.ResourceAsString(addon.Contents)
	if err != nil {
		t.Fatalf("reading addon contents: %v", err)
	}
	if actual != "apiVersion: v1\nkind: ConfigMap\ndata:\n  literal: rendered" {
		t.Fatalf("addon contents = %q", actual)
	}
}

func TestAddonManifestNormalizeProtectedInstanceGroups(t *testing.T) {
	allInstanceGroups := []*kops.InstanceGroup{
		testInstanceGroup("control-plane", kops.InstanceGroupRoleControlPlane, kops.InstanceManagerCloudGroup),
		testInstanceGroup("karpenter-a", kops.InstanceGroupRoleNode, kops.InstanceManagerKarpenter),
		testInstanceGroup("karpenter-b", kops.InstanceGroupRoleNode, kops.InstanceManagerKarpenter),
		// Not managed by Karpenter, for example after being moved to CloudGroup.
		testInstanceGroup("nodes", kops.InstanceGroupRoleNode, kops.InstanceManagerCloudGroup),
	}
	// The objects rendered when karpenter-a is updated; kOps names them after the InstanceGroup.
	manifest := "apiVersion: karpenter.k8s.aws/v1\nkind: EC2NodeClass\nmetadata:\n  name: karpenter-a\n---\napiVersion: karpenter.sh/v1\nkind: NodePool\nmetadata:\n  name: karpenter-a\n"

	tests := []struct {
		name           string
		instanceGroups []*kops.InstanceGroup
		// wantFieldSelector is expected on the EC2NodeClass and NodePool prune directives only.
		wantFieldSelector string
	}{
		{
			name:              "all instance groups",
			instanceGroups:    allInstanceGroups,
			wantFieldSelector: "metadata.name!=karpenter-a,metadata.name!=karpenter-b",
		},
		{
			name:              "control plane only",
			instanceGroups:    allInstanceGroups[:1],
			wantFieldSelector: "metadata.name!=karpenter-a,metadata.name!=karpenter-b,metadata.name!=nodes",
		},
		{
			name:              "one Karpenter instance group",
			instanceGroups:    allInstanceGroups[1:2],
			wantFieldSelector: "metadata.name!=control-plane,metadata.name!=karpenter-a,metadata.name!=karpenter-b,metadata.name!=nodes",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, err := fi.NewCloudupContext(context.Background(), fi.DeletionProcessingModeDeleteIncludingDeferred, nil, nil, nil, nil, nil, nil, map[string]fi.CloudupTask{})
			if err != nil {
				t.Fatalf("building cloudup context: %v", err)
			}

			addon := &AddonManifest{
				Name:       new("karpenter.sh"),
				Location:   new("addons/karpenter.sh/k8s-1.19.yaml"),
				source:     fi.NewBytesResource([]byte(manifest)),
				skipRender: true,
				skipRemap:  true,
				addonSpec:  testAddonSpec("karpenter.sh"),
				buildPrune: true,
				modelContext: &model.KopsModelContext{
					AllInstanceGroups: allInstanceGroups,
					InstanceGroups:    tc.instanceGroups,
				},
			}
			if err := addon.Normalize(ctx); err != nil {
				t.Fatalf("normalizing addon: %v", err)
			}

			karpenterKinds := 0
			for _, kind := range addon.addonSpec.Prune.Kinds {
				want := ""
				if kind.Kind == "EC2NodeClass" || kind.Kind == "NodePool" {
					want = tc.wantFieldSelector
					karpenterKinds++
				}
				if kind.FieldSelector != want {
					t.Errorf("field selector for %s = %q, want %q", kind.Kind, kind.FieldSelector, want)
				}
			}
			if karpenterKinds != 2 {
				t.Errorf("found %d Karpenter kinds in the prune directives, want 2", karpenterKinds)
			}
		})
	}
}

func TestAddonCollectImagesSkipsRenderForRawSources(t *testing.T) {
	assetBuilder := assets.NewAssetBuilder(nil, nil, false)
	rawManifest := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: raw\ndata:\n  helm: \"{{ .Values.image\"\n")
	renderer := &recordingAddonRenderer{
		rendered: []byte("not: valid: yaml\n"),
	}
	addon := &Addon{
		Spec:       testAddonSpec("raw-addon"),
		Source:     fi.NewBytesResource(rawManifest),
		SkipRender: true,
	}

	if err := addon.CollectImages(assetBuilder, renderer); err != nil {
		t.Fatalf("collecting raw addon images: %v", err)
	}
	if renderer.calls != 0 {
		t.Fatalf("renderer calls = %d, want 0", renderer.calls)
	}
}

func TestAddonCollectImagesRendersTemplateSources(t *testing.T) {
	assetBuilder := assets.NewAssetBuilder(nil, nil, false)
	renderer := &recordingAddonRenderer{
		rendered: []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: rendered\nspec:\n  containers:\n  - name: container\n    image: registry.k8s.io/pause:3.9\n"),
	}
	addon := &Addon{
		Spec:   testAddonSpec("template-addon"),
		Source: fi.NewBytesResource([]byte("{{ template }}")),
	}

	if err := addon.CollectImages(assetBuilder, renderer); err != nil {
		t.Fatalf("collecting template addon images: %v", err)
	}
	if renderer.calls != 1 {
		t.Fatalf("renderer calls = %d, want 1", renderer.calls)
	}
}

// kops-channels applies and prunes an addon only when its manifest hash changes, so the hash must
// change when objects that were kept from pruning are pruned again, even if the manifest does not.
func TestAddonManifestNormalizeHashesKeptObjects(t *testing.T) {
	ctx, err := fi.NewCloudupContext(context.Background(), fi.DeletionProcessingModeDeleteIncludingDeferred, nil, nil, nil, nil, nil, nil, map[string]fi.CloudupTask{})
	if err != nil {
		t.Fatalf("building cloudup context: %v", err)
	}

	// Only the control plane is updated, as in the first step of "kops reconcile cluster", so the
	// manifest has no objects of the Karpenter instance groups, which are kept from pruning.
	controlPlane := testInstanceGroup("control-plane", kops.InstanceGroupRoleControlPlane, kops.InstanceManagerCloudGroup)
	build := func(kept []string) *channels.Addon {
		allInstanceGroups := []*kops.InstanceGroup{controlPlane}
		for _, name := range kept {
			allInstanceGroups = append(allInstanceGroups, testInstanceGroup(name, kops.InstanceGroupRoleNode, kops.InstanceManagerKarpenter))
		}
		spec := testAddonSpec("karpenter.sh")
		addon := &AddonManifest{
			Name:       new("karpenter.sh"),
			Location:   new("addons/karpenter.sh.yaml"),
			source:     fi.NewStringResource("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: karpenter\n  namespace: kube-system\n"),
			skipRender: true,
			skipRemap:  true,
			addonSpec:  spec,
			buildPrune: true,
			modelContext: &model.KopsModelContext{
				AllInstanceGroups: allInstanceGroups,
				InstanceGroups:    []*kops.InstanceGroup{controlPlane},
			},
		}
		if err := addon.Normalize(ctx); err != nil {
			t.Fatalf("normalizing addon: %v", err)
		}
		return &channels.Addon{Name: "karpenter.sh", ChannelName: "bootstrap", Spec: spec}
	}

	grid := []struct {
		desc           string
		previous, next []string
		expectUpdate   bool
	}{
		{
			desc:         "the only kept instance group is deleted",
			previous:     []string{"nodes-a"},
			expectUpdate: true,
		},
		{
			desc:         "one of the kept instance groups is deleted",
			previous:     []string{"nodes-a", "nodes-b"},
			next:         []string{"nodes-b"},
			expectUpdate: true,
		},
		{
			desc:     "the same instance groups are kept",
			previous: []string{"nodes-b"},
			next:     []string{"nodes-b"},
		},
		{
			desc: "no instance groups are kept",
		},
	}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			update, err := build(g.next).GetRequiredUpdates(context.Background(), nil, nil, build(g.previous).ChannelVersion())
			if err != nil {
				t.Fatalf("getting required updates: %v", err)
			}
			if actual := update != nil; actual != g.expectUpdate {
				t.Errorf("addon update required = %v, want %v", actual, g.expectUpdate)
			}
		})
	}
}

func testAddonSpec(name string) *channelsapi.AddonSpec {
	return &channelsapi.AddonSpec{
		Name:     new(name),
		Selector: map[string]string{"k8s-addon": name},
		Manifest: new(name + ".yaml"),
	}
}

func testInstanceGroup(name string, role kops.InstanceGroupRole, manager kops.InstanceManager) *kops.InstanceGroup {
	return &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       kops.InstanceGroupSpec{Role: role, Manager: manager},
	}
}
