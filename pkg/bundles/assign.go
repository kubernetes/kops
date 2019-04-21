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

package bundles

import (
	"fmt"

	"github.com/blang/semver"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
)

// AssignBundle will assign a bundle for the current cluster; upgrade will later advise on recommended upgrades to the bundle version
func AssignBundle(c *kops.Cluster) error {
	k8sVersion, err := KubernetesVersion(c)
	if err != nil {
		return err
	}

	if c.Spec.Bundle == "" {
		channel, err := kops.LoadChannel(c.Spec.Channel)
		if err != nil {
			return err
		}
		bundleSpec := channel.FindBundleVersion(k8sVersion)
		if bundleSpec == nil {
			return fmt.Errorf("cannot find current bundle in channel %s", c.Spec.Channel)
		}
		c.Spec.Bundle = bundleSpec.Bundle
	}

	return nil
}

func KubernetesVersion(c *kops.Cluster) (semver.Version, error) {
	k8sVersion, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return semver.Version{}, fmt.Errorf("unable to parse KubernetesVersion %q", c.Spec.KubernetesVersion)
	}
	return *k8sVersion, nil
}
