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

package builder

import (
	"context"
	"testing"
)

func TestAzureBlobStoreURIs(t *testing.T) {
	store, err := (&BuildOptions{StageLocation: "azureblob://stkopsstatestore/kops-e2e-staging/pulls/job/"}).azureBlobStore()
	if err != nil {
		t.Fatalf("azureBlobStore() = %v", err)
	}
	if store.stagePrefix() != "pulls/job" {
		t.Errorf("stagePrefix() = %q, expected %q", store.stagePrefix(), "pulls/job")
	}

	urlPath := "pulls/job/1.37.0-alpha.1+v1.37.0-alpha.0-12-gabcdef0"
	if actual, expected := store.uri(urlPath), "azureblob://stkopsstatestore/kops-e2e-staging/"+urlPath; actual != expected {
		t.Errorf("uri(%q) = %q, expected %q", urlPath, actual, expected)
	}
	actual, err := store.objectURL(context.Background(), urlPath)
	if err != nil {
		t.Fatalf("objectURL(%q) = %v", urlPath, err)
	}
	if expected := "https://stkopsstatestore.blob.core.windows.net/kops-e2e-staging/" + urlPath; actual != expected {
		t.Errorf("objectURL(%q) = %q, expected %q", urlPath, actual, expected)
	}
}

// krel cannot push to Azure, which must be detected before anything is built.
func TestBuildKubernetesToAzureBlobFails(t *testing.T) {
	b := &BuildOptions{StageLocation: "azureblob://account/staging", BuildKubernetes: true}
	if _, err := b.Build(); err == nil {
		t.Error("Build() succeeded, expected an error for building Kubernetes to an azureblob:// stage location")
	}
}
