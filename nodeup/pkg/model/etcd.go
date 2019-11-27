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
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog"
)

// EtcdBuilder installs etcd
type EtcdBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcdBuilder{}

// Build is responsible for creating the etcd user
func (b *EtcdBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	switch b.Distribution {
	case distros.DistributionCoreOS:
		klog.Infof("Detected CoreOS; skipping etcd user installation")
		return nil

	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; skipping etcd user installation")
		return nil

	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; skipping etcd user installation")
		return nil
	}

	// TODO: Do we actually use the user anywhere?

	c.AddTask(&nodetasks.UserTask{
		// TODO: Should we set a consistent UID in case we remount?
		Name:  "user",
		Shell: "/sbin/nologin",
		Home:  "/var/etcd",
	})

	return nil
}
