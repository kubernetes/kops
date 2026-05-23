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

package templates

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kops/util/pkg/vfs"
)

func TestLoadTemplatesTracksTemplateSources(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "addons", "example"), 0755); err != nil {
		t.Fatalf("creating addon dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "addons", "example", "plain.yaml"), []byte("plain"), 0644); err != nil {
		t.Fatalf("writing plain resource: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "addons", "example", "rendered.yaml.template"), []byte("template"), 0644); err != nil {
		t.Fatalf("writing template resource: %v", err)
	}

	templates, err := LoadTemplates(context.Background(), vfs.NewFSPath(dir))
	if err != nil {
		t.Fatalf("loading templates: %v", err)
	}

	if templates.Find("addons/example/plain.yaml") == nil {
		t.Fatalf("plain resource was not loaded")
	}
	if templates.IsTemplate("addons/example/plain.yaml") {
		t.Fatalf("plain resource reported as template")
	}

	if templates.Find("addons/example/rendered.yaml") == nil {
		t.Fatalf("template resource was not loaded under trimmed key")
	}
	if !templates.IsTemplate("addons/example/rendered.yaml") {
		t.Fatalf("template resource did not report template source")
	}
}
