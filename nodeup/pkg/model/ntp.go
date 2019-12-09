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

// NTPBuilder installs and starts NTP, to ensure accurate clock times.
// As well as general log confusion, clock-skew of more than 5 minutes
// causes AWS API calls to fail
type NTPBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &NTPBuilder{}

// Build is responsible for configuring NTP
func (b *NTPBuilder) Build(c *fi.ModelBuilderContext) error {
	switch b.Distribution {
	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install ntp")
		return nil
	case distros.DistributionCoreOS:
		klog.Infof("Detected CoreOS; won't install ntp")
		return nil
	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install ntp")
		return nil
	}

	if b.Distribution.IsDebianFamily() {
		c.AddTask(&nodetasks.Package{Name: "ntp"})
		c.AddTask((&nodetasks.Service{Name: "ntp"}).InitDefaults())
	} else if b.Distribution.IsRHELFamily() {
		c.AddTask(&nodetasks.Package{Name: "ntp"})
		c.AddTask((&nodetasks.Service{Name: "ntpd"}).InitDefaults())
	} else {
		klog.Warningf("unknown distribution, skipping ntp install: %v", b.Distribution)
		return nil
	}

	return nil
}
