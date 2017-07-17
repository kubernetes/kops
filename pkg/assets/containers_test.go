/*
Copyright 2016 The Kubernetes Authors.

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

package assets

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func Test_Build_Google_Containers(t *testing.T) {

	cs := &kops.ClusterSpec{}

	url := "kubedns-amd64:1.3"

	s, err := GetGoogleImageRegistryContainer(cs, url)

	if err != nil {
		t.Fatalf("incorrect container repo expect: %v", err)
	}

	base := GCR_IO + url

	if s != base {
		t.Fatalf("incorrect container repo expect: %q, got %q", base, s)
	}

	repo := "quay.io/chrislovecnm"
	cs = &kops.ClusterSpec{
		Assets: &kops.Assets{
			ContainerRegistry: &repo,
		},
	}

	googleRepository = nil

	s, err = GetGoogleImageRegistryContainer(cs, url)

	if err != nil {
		t.Fatalf("incorrect container repo expect: %v", err)
	}

	base = *cs.Assets.ContainerRegistry + "/" + url

	if s != base {
		t.Fatalf("incorrect container repo expect: %q, got %q", base, s)
	}

	googleRepository = nil

	repo = "chrislovecnm"
	cs = &kops.ClusterSpec{
		Assets: &kops.Assets{
			ContainerRegistry: &repo,
		},
	}

	googleRepository = nil

	s, err = GetGoogleImageRegistryContainer(cs, url)

	if err != nil {
		t.Fatalf("incorrect container repo expect: %v", err)
	}

	base = *cs.Assets.ContainerRegistry + "/" + url

	if s != base {
		t.Fatalf("incorrect container repo expect: %q, got %q", base, s)
	}

	googleRepository = nil

	repo = "https://aa/asdf$$^/aa"
	cs = &kops.ClusterSpec{
		Assets: &kops.Assets{
			ContainerRegistry: &repo,
		},
	}

	googleRepository = nil

	s, err = GetGoogleImageRegistryContainer(cs, url)

	if err == nil {
		t.Fatalf("incorrect container failure expected: %v", err)
	}

	googleRepository = nil
}
