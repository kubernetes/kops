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

// DockerBuilder install docker (just the packages at the moment)
type EC2InstanceConnectBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EC2InstanceConnectBuilder{}

// Build is responsible for installing packages
func (b *EC2InstanceConnectBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.InstanceGroup.Spec.EC2InstanceConnect {
		return nil
	}

	if b.Distribution.IsUbuntu() || b.Distribution == distros.DistributionAmazonLinux2 {
		c.AddTask(&nodetasks.Package{Name: "ec2-instance-connect"})
	} else if b.Distribution.IsDebianFamily() {
		c.AddTask(&nodetasks.Package{
			Name:   "ec2-instance-connect",
			Source: s("http://archive.ubuntu.com/ubuntu/pool/universe/e/ec2-instance-connect/ec2-instance-connect_1.1.12+dfsg1-0ubuntu3~19.10.0_all.deb"),
			Hash:   s("52768695e6b9bac9e55bf6324026c240"),
		})
	} else {
		klog.Warningf("unsupported distribution, skipping installation on: %v", b.Distribution)
	}

	return nil

}
