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

package model

import (
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func Test_ProxyFunc(t *testing.T) {
	b := &BootstrapScript{}
	ps := &kops.EgressProxySpec{
		HTTPProxy: kops.HTTPProxy{
			Host: "example.com",
			Port: 80,
		},
	}

	script := b.createProxyEnv(ps)

	if script == "" {
		t.Fatalf("script cannot be empty")
	}

	if !strings.HasPrefix(script, "export http_proxy=http://example.com:80") {
		t.Fatalf("script not setting http_proxy properly")
	}

	ps.ProxyExcludes = "www.google.com,www.kubernetes.io"

	script = b.createProxyEnv(ps)
	t.Logf(script)
	if !strings.Contains(script, "export no_proxy="+ps.ProxyExcludes) {
		t.Fatalf("script not setting no_proxy properly")
	}

}
