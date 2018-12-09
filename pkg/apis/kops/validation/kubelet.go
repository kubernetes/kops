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

package validation

import (
	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/validation/field"
	api "k8s.io/kops/pkg/apis/kops"
)

func validateKubeletConfig(kubeletPath *field.Path, kubernetesRelease semver.Version, kubeletConfig *api.KubeletConfigSpec, strict bool) *field.Error {
	// The ExperimentalAllowedUnsafeSysctls flag was renamed in k/k #63717
	// and moved to AllowedUnsafeSysctls.
	if kubeletConfig.ExperimentalAllowedUnsafeSysctls != nil {
		if kubernetesRelease.GTE(semver.MustParse("1.11.0")) {
			glog.V(1).Info("ExperimentalAllowedUnsafeSysctls was renamed in Kubernetes 1.11+, please use AllowedUnsafeSysctls instead.")
			kubeletConfig.AllowedUnsafeSysctls = append(kubeletConfig.ExperimentalAllowedUnsafeSysctls, kubeletConfig.AllowedUnsafeSysctls...)
			kubeletConfig.ExperimentalAllowedUnsafeSysctls = nil
		}
	}

	if kubernetesRelease.GTE(semver.MustParse("1.6.0")) {
		// Flag removed in 1.6
		if kubeletConfig.APIServers != "" {
			return field.Invalid(
				kubeletPath.Child("APIServers"),
				kubeletConfig.APIServers,
				"api-servers flag was removed in 1.6")
		}
	} else {
		if strict && kubeletConfig.APIServers == "" {
			return field.Required(kubeletPath.Child("APIServers"), "")
		}
	}

	if kubernetesRelease.GTE(semver.MustParse("1.10.0")) {
		// Flag removed in 1.10
		if kubeletConfig.RequireKubeconfig != nil {
			return field.Invalid(
				kubeletPath.Child("requireKubeconfig"),
				*kubeletConfig.RequireKubeconfig,
				"require-kubeconfig flag was removed in 1.10.  (Please be sure you are not using a cluster config from `kops get cluster --full`)")
		}
	}

	if kubeletConfig.APIServers != "" && !isValidAPIServersURL(kubeletConfig.APIServers) {
		return field.Invalid(kubeletPath.Child("APIServers"), kubeletConfig.APIServers, "Not a valid APIServer URL")
	}

	return nil
}
