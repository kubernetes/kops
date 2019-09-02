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

package env

import (
	"os"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/proxy"
)

type EnvVars map[string]string

func (m EnvVars) addEnvVariableIfExist(name string) {
	v := os.Getenv(name)
	if v != "" {
		m[name] = v
	}
}

func BuildSystemComponentEnvVars(spec *kops.ClusterSpec) EnvVars {
	vars := make(EnvVars)

	for _, v := range proxy.GetProxyEnvVars(spec.EgressProxy) {
		vars[v.Name] = v.Value
	}

	// Custom S3 endpoint
	vars.addEnvVariableIfExist("S3_ENDPOINT")
	vars.addEnvVariableIfExist("S3_ACCESS_KEY_ID")
	vars.addEnvVariableIfExist("S3_SECRET_ACCESS_KEY")

	// Openstack related values
	vars.addEnvVariableIfExist("OS_TENANT_ID")
	vars.addEnvVariableIfExist("OS_TENANT_NAME")
	vars.addEnvVariableIfExist("OS_PROJECT_ID")
	vars.addEnvVariableIfExist("OS_PROJECT_NAME")
	vars.addEnvVariableIfExist("OS_PROJECT_DOMAIN_NAME")
	vars.addEnvVariableIfExist("OS_PROJECT_DOMAIN_ID")
	vars.addEnvVariableIfExist("OS_DOMAIN_NAME")
	vars.addEnvVariableIfExist("OS_DOMAIN_ID")
	vars.addEnvVariableIfExist("OS_USERNAME")
	vars.addEnvVariableIfExist("OS_PASSWORD")
	vars.addEnvVariableIfExist("OS_AUTH_URL")
	vars.addEnvVariableIfExist("OS_REGION_NAME")

	// Digital Ocean related values.
	vars.addEnvVariableIfExist("DIGITALOCEAN_ACCESS_TOKEN")

	return vars
}

func (m EnvVars) ToEnvVars() []corev1.EnvVar {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var l []corev1.EnvVar
	for _, k := range keys {
		l = append(l, corev1.EnvVar{Name: k, Value: m[k]})
	}

	return l
}
