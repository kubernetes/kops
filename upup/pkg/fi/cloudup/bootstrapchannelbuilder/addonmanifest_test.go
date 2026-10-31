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
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
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

// testManifest joins objects as kubemanifest.ObjectList.ToYAML does.
func testManifest(objects ...string) string {
	return strings.Join(objects, "\n---\n\n")
}

// testInstanceGroups returns the instance groups with the given names: "control-plane" is a control
// plane instance group, the names in cloudGroups are cloud group nodes, and the others are Karpenter
// nodes.
func testInstanceGroups(cloudGroups []string, names ...string) []*kops.InstanceGroup {
	var instanceGroups []*kops.InstanceGroup
	for _, name := range names {
		ig := &kops.InstanceGroup{ObjectMeta: metav1.ObjectMeta{Name: name}}
		switch {
		case name == "control-plane":
			ig.Spec.Role = kops.InstanceGroupRoleControlPlane
		case slices.Contains(cloudGroups, name):
			ig.Spec.Role = kops.InstanceGroupRoleNode
			ig.Spec.Manager = kops.InstanceManagerCloudGroup
		default:
			ig.Spec.Role = kops.InstanceGroupRoleNode
			ig.Spec.Manager = kops.InstanceManagerKarpenter
		}
		instanceGroups = append(instanceGroups, ig)
	}
	return instanceGroups
}

func TestAddonManifestNormalizeKeepsInstanceGroupObjects(t *testing.T) {
	settings := func(version string) string {
		return fmt.Sprintf("apiVersion: v1\ndata:\n  version: %s\nkind: ConfigMap\nmetadata:\n  name: settings\n", version)
	}
	// objects returns the manifest of the EC2NodeClass and NodePool of an instance group.
	objects := func(name, version string) string {
		return testManifest(
			fmt.Sprintf("apiVersion: karpenter.k8s.aws/v1\nkind: EC2NodeClass\nmetadata:\n  name: %s\nspec:\n  version: %s\n", name, version),
			fmt.Sprintf("apiVersion: karpenter.sh/v1\nkind: NodePool\nmetadata:\n  name: %s\nspec:\n  version: %s\n", name, version),
		)
	}
	// The previous manifest has the objects of the Karpenter instance groups a and b, and of the
	// instance group "deleted", which no longer exists.
	previousManifest := testManifest(settings("old"), objects("a", "old"), objects("b", "old"), objects("deleted", "old"))

	grid := []struct {
		desc              string
		allInstanceGroups []string
		instanceGroups    []string
		cloudGroups       []string
		previousManifest  *string
		rendered          string
		expected          string
		expectedError     string
	}{
		{
			desc:              "full update",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"control-plane", "a", "b"},
			// A full update does not read the previous manifest.
			previousManifest: new("not: valid: yaml"),
			rendered:         testManifest(settings("new"), objects("a", "new")),
			expected:         testManifest(settings("new"), objects("a", "new")),
		},
		{
			desc:              "control plane",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"control-plane"},
			previousManifest:  &previousManifest,
			rendered:          settings("new"),
			expected:          testManifest(settings("new"), objects("a", "old"), objects("b", "old")),
		},
		{
			// The objects follow the order of the instance groups, not of the manifests.
			desc:              "one Karpenter instance group",
			allInstanceGroups: []string{"control-plane", "b", "a"},
			instanceGroups:    []string{"a"},
			previousManifest:  &previousManifest,
			rendered:          testManifest(settings("new"), objects("a", "new")),
			expected:          testManifest(settings("new"), objects("b", "old"), objects("a", "new")),
		},
		{
			// For example, an instance group deleted and created again as a cloud group.
			desc:              "instance group being updated without objects",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"a"},
			previousManifest:  &previousManifest,
			rendered:          settings("new"),
			expected:          testManifest(settings("new"), objects("b", "old")),
		},
		{
			// For example, an instance group deleted and created again as a cloud group, which is not
			// being updated.
			desc:              "instance group not being updated that is a cloud group",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"control-plane"},
			cloudGroups:       []string{"b"},
			previousManifest:  &previousManifest,
			rendered:          settings("new"),
			expected:          testManifest(settings("new"), objects("a", "old")),
		},
		{
			desc:              "no objects to keep",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"control-plane"},
			previousManifest:  new(testManifest(settings("old"), objects("deleted", "old"))),
			rendered:          settings("new"),
			expected:          settings("new"),
		},
		{
			desc:              "no previous manifest",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"control-plane"},
			rendered:          settings("new"),
			expected:          settings("new"),
		},
		{
			desc:              "invalid previous manifest",
			allInstanceGroups: []string{"control-plane", "a", "b"},
			instanceGroups:    []string{"control-plane"},
			previousManifest:  new("not: valid: yaml"),
			rendered:          settings("new"),
			expectedError:     "parsing the previous manifest",
		},
	}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ctx := context.Background()
			configBase := vfs.NewMemFSPath(vfs.NewMemFSContext(), "clusters.example.com/minimal.example.com")
			location := "addons/karpenter.sh/k8s-1.19.yaml"
			if g.previousManifest != nil {
				if err := configBase.Join(location).WriteFile(ctx, bytes.NewReader([]byte(*g.previousManifest)), nil); err != nil {
					t.Fatalf("writing the previous manifest: %v", err)
				}
			}
			cloudupContext, err := fi.NewCloudupContext(ctx, fi.DeletionProcessingModeDeleteIncludingDeferred, nil, nil, nil, nil, nil, configBase, map[string]fi.CloudupTask{})
			if err != nil {
				t.Fatalf("building cloudup context: %v", err)
			}

			addon := &AddonManifest{
				Name:          new("karpenter.sh"),
				Location:      new(location),
				source:        fi.NewBytesResource([]byte("{{ template }}")),
				addonRenderer: &recordingAddonRenderer{rendered: []byte(g.rendered)},
				addonSpec:     testAddonSpec("karpenter.sh"),
				skipRemap:     true,
				modelContext: &model.KopsModelContext{
					AllInstanceGroups: testInstanceGroups(g.cloudGroups, g.allInstanceGroups...),
					InstanceGroups:    testInstanceGroups(g.cloudGroups, g.instanceGroups...),
				},
				instanceGroupKinds: karpenterInstanceGroupKinds,
			}

			err = addon.Normalize(cloudupContext)
			if g.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), g.expectedError) {
					t.Fatalf("Normalize() error = %v, want an error containing %q", err, g.expectedError)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizing addon: %v", err)
			}
			actual, err := fi.ResourceAsString(addon.Contents)
			if err != nil {
				t.Fatalf("reading addon contents: %v", err)
			}
			if expected := strings.TrimSpace(g.expected); actual != expected {
				t.Errorf("addon contents:\n%s\nwant:\n%s", actual, expected)
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
