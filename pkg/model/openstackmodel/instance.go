/*
Copyright 2018 The Kubernetes Authors.

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

package openstackmodel

import (
	"strings"

	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

// InstanceModelBuilder configures instance objects
type InstanceModelBuilder struct {
	*OpenstackModelContext
	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &InstanceModelBuilder{}

func (b *InstanceModelBuilder) Build(c *fi.ModelBuilderContext) error {
	sshKeyNameFull, err := b.SSHKeyName()
	if err != nil {
		return err
	}

	splitSSHKeyNameFull := strings.Split(sshKeyNameFull, "-")
	sshKeyName := splitSSHKeyNameFull[0]

	clusterTag := "KubernetesCluster:" + strings.Replace(b.ClusterName(), ".", "-", -1)

	// In the future, OpenStack will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of servers
	for _, ig := range b.InstanceGroups {
		clusterName := b.AutoscalingGroupName(ig)

		{
			t := &openstacktasks.Port{
				Name:      s(clusterName),
				Network:   b.LinkToNetwork(),
				Lifecycle: b.Lifecycle,
			}
			c.AddTask(t)
		}

		{
			var t openstacktasks.Instance
			t.Count = int(fi.Int32Value(ig.Spec.MinSize))
			t.Name = fi.String(clusterName)
			t.Region = fi.String(b.Cluster.Spec.Subnets[0].Region)
			t.Flavor = fi.String(ig.Spec.MachineType)
			t.Image = fi.String(ig.Spec.Image)
			t.SSHKey = fi.String(sshKeyName)
			t.Tags = []string{clusterTag}
			t.Role = fi.String(string(ig.Spec.Role))
			t.Port = b.LinkToPort(fi.String(clusterName))
			c.AddTask(&t)
		}
	}
	return nil
}
