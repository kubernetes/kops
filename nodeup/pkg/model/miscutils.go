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

package model

import (
	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// MiscUtilsBuilder ensures that some system packages that are
// required for kubernetes are installed (e.g. socat)
type MiscUtilsBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &MiscUtilsBuilder{}

// Build is responsible for configuring the miscellaneous packages we want installed
func (b *MiscUtilsBuilder) Build(c *fi.ModelBuilderContext) error {
	switch b.Distribution {
	case distros.DistributionContainerOS:
		klog.V(2).Infof("Detected ContainerOS; won't install misc. utils")
		return nil
	case distros.DistributionCoreOS:
		klog.V(2).Infof("Detected CoreOS; won't install misc. utils")
		return nil
	case distros.DistributionFlatcar:
		klog.V(2).Infof("Detected Flatcar; won't install misc. utils")
		return nil
	}

	// TODO: These packages have been auto-installed for a long time, and likely we don't need all of them any longer
	// We could prune from auto-install at a particular k8s release (e.g. 1.13?)

	var packages []string
	if b.Distribution.IsDebianFamily() {
		packages = append(packages, "curl")
		packages = append(packages, "wget")
		packages = append(packages, "nfs-common")
		packages = append(packages, "perl")
		packages = append(packages, "python-apt")
		packages = append(packages, "apt-transport-https")
	} else if b.Distribution.IsRHELFamily() {
		packages = append(packages, "curl")
		packages = append(packages, "wget")
		packages = append(packages, "nfs-utils")
		packages = append(packages, "python2")
		packages = append(packages, "git")
	} else {
		klog.Warningf("unknown distribution, skipping misc utils install: %v", b.Distribution)
		return nil
	}

	if b.Distribution.IsUbuntu() {
		packages = append(packages, "netcat-traditional")
		packages = append(packages, "git")
	}

	for _, p := range packages {
		c.AddTask(&nodetasks.Package{Name: p})
	}

	return nil
}
