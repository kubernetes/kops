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

package components

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func Test_Build_Google_File_URL(t *testing.T) {

	cs := &kops.ClusterSpec{}

	url := "foo"

	s, err := GetGoogleFileRepositoryURL(cs, url)

	if err != nil {
		t.Fatalf("unable to parse url %q", url)
	}

	if s != GCR_STORAGE+"/"+url {
		t.Fatalf("incorret url %q", s)
	}
}
