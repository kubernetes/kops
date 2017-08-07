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
	"strconv"

	"github.com/golang/glog"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// s is a helper that builds a *string from a string value
func s(v string) *string {
	return fi.String(v)
}

// i64 is a helper that builds a *int64 from an int64 value
func i64(v int64) *int64 {
	return fi.Int64(v)
}

func getProxyEnvVars(proxies *kops.EgressProxySpec) []v1.EnvVar {
	if proxies == nil {
		glog.V(8).Info("proxies is == nil, returning empty list")
		return []v1.EnvVar{}
	}

	if proxies.HTTPProxy.Host == "" {
		glog.Warning("EgressProxy set but no proxy host provided")
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

// b returns a pointer to a boolean
func b(v bool) *bool {
	return fi.Bool(v)
}
