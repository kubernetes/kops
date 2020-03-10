/*
Copyright 2020 The Kubernetes Authors.

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

package proxy

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
)

func TestGetProxyEnvVars(t *testing.T) {
	tests := []struct {
		inProxies *kops.EgressProxySpec
		expected  []v1.EnvVar
	}{
		{
			inProxies: nil,
			expected:  []v1.EnvVar{},
		},
		{
			inProxies: &kops.EgressProxySpec{
				HTTPProxy: kops.HTTPProxy{
					Host: "",
					Port: 1234,
				},
			},
			// TODO: The output seems a little bit weird for this case
			// Should we return empty array instead of just warning?
			expected: []v1.EnvVar{
				{Name: "http_proxy", Value: "http://:1234"},
				{Name: "https_proxy", Value: "http://:1234"},
				{Name: "NO_PROXY", Value: ""},
				{Name: "no_proxy", Value: ""},
			},
		},
		{
			inProxies: &kops.EgressProxySpec{
				HTTPProxy: kops.HTTPProxy{
					Host: "a.b.c.d",
					Port: 1234,
				},
			},
			expected: []v1.EnvVar{
				{Name: "http_proxy", Value: "http://a.b.c.d:1234"},
				{Name: "https_proxy", Value: "http://a.b.c.d:1234"},
				{Name: "NO_PROXY", Value: ""},
				{Name: "no_proxy", Value: ""},
			},
		},
		{
			inProxies: &kops.EgressProxySpec{
				HTTPProxy: kops.HTTPProxy{
					Host: "a.b.c.d",
					Port: 0,
				},
				ProxyExcludes: "1.1.1.1,2.2.2.2",
			},
			expected: []v1.EnvVar{
				{Name: "http_proxy", Value: "http://a.b.c.d"},
				{Name: "https_proxy", Value: "http://a.b.c.d"},
				{Name: "NO_PROXY", Value: "1.1.1.1,2.2.2.2"},
				{Name: "no_proxy", Value: "1.1.1.1,2.2.2.2"},
			},
		},
	}

	for _, test := range tests {
		result := GetProxyEnvVars(test.inProxies)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Actual result:\n %v \nExpect:\n %v", result, test.expected)
		}
	}
}
