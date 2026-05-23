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

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/assets"
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
		Name:          fi.PtrTo("raw-addon"),
		Location:      fi.PtrTo("addons/raw-addon.yaml"),
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
		Name:          fi.PtrTo("template-addon"),
		Location:      fi.PtrTo("addons/template-addon.yaml"),
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

func testAddonSpec(name string) *channelsapi.AddonSpec {
	return &channelsapi.AddonSpec{
		Name:     fi.PtrTo(name),
		Selector: map[string]string{"k8s-addon": name},
		Manifest: fi.PtrTo(name + ".yaml"),
	}
}
