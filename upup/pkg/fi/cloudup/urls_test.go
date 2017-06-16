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

package cloudup

import (
	"fmt"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"testing"
)

// TODO test env var

func Test_BaseUrl(t *testing.T) {
	baseUrl = ""
	c := buildMinimalCluster()

	b := BaseUrl(&c.Spec)

	base := fmt.Sprintf("https://kubeupv2.s3.amazonaws.com/kops/%s/", kops.Version)

	baseUrl = ""
	if b != base {
		t.Fatalf("wrong url %q, expected %q", b, base)
	}

	b = "https://example.com/"
	f := fmt.Sprintf("%skops/%s/", b, kops.Version)
	c.Spec.Assets = &api.Assets{
		FileRepository: &b,
	}

	baseUrl = ""
	b = BaseUrl(&c.Spec)

	if b != f {
		t.Fatalf("wrong url %q, expected %q", b, f)
	}

}

// TODO test ENV var
func Test_ProtokubeUrl(t *testing.T) {
	baseUrl = ""
	c := buildMinimalCluster()

	b, err := ProtokubeImageSource(&c.Spec)

	if err != nil {
		t.Fatalf("error building protokube url: %v", err)
	}

	f := "https://kubeupv2.s3.amazonaws.com/kops/1.5.0/images/protokube.tar.gz"

	if b.Source != f {
		t.Fatalf("wrong protokube image %q, expected %q", b.Source, f)
	}

	repo := "quay.io/foo"
	c.Spec.Assets = &api.Assets{
		ContainerRegistry: &repo,
	}

	protokubeImageSource = nil

	b, err = ProtokubeImageSource(&c.Spec)

	if err != nil {
		t.Fatalf("error building protokube url: %v", err)
	}

	p := repo + "/protokube:" + kops.Version
	if b.Source != p {
		t.Fatalf("wrong url %q, expected %q", b, p)
	}

}
