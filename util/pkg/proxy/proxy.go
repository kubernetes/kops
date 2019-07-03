/*
Copyright 2019 The Kubernetes Authors.

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
	"strconv"

	"k8s.io/kops/pkg/apis/kops"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func GetProxyEnvVars(proxies *kops.EgressProxySpec) []v1.EnvVar {
	if proxies == nil {
		klog.V(8).Info("proxies is == nil, returning empty list")
		return []v1.EnvVar{}
	}

	if proxies.HTTPProxy.Host == "" {
		klog.Warning("EgressProxy set but no proxy host provided")
	}

	var httpProxyURL string
	if proxies.HTTPProxy.Port == 0 {
		httpProxyURL = "http://" + proxies.HTTPProxy.Host
	} else {
		httpProxyURL = "http://" + proxies.HTTPProxy.Host + ":" + strconv.Itoa(proxies.HTTPProxy.Port)
	}

	noProxy := proxies.ProxyExcludes

	return []v1.EnvVar{
		{Name: "http_proxy", Value: httpProxyURL},
		{Name: "https_proxy", Value: httpProxyURL},
		{Name: "NO_PROXY", Value: noProxy},
		{Name: "no_proxy", Value: noProxy},
	}
}
