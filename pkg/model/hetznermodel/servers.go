/*
Copyright 2022 The Kubernetes Authors.

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

package hetznermodel

import (
	"strconv"

	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

// ServerModelBuilder configures network objects
type ServerModelBuilder struct {
	*HetznerModelContext
	Lifecycle              fi.Lifecycle
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
}

var _ fi.ModelBuilder = &ServerModelBuilder{}

func (b *ServerModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		igSize := fi.Int32Value(ig.Spec.MinSize)

		labels := make(map[string]string)
		labels[hetzner.TagKubernetesClusterName] = b.ClusterName()
		labels[hetzner.TagKubernetesInstanceGroup] = ig.Name
		labels[hetzner.TagKubernetesInstanceRole] = string(ig.Spec.Role)

		userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return err
		}

		for i := 1; i <= int(igSize); i++ {
			// hcloud-cloud-controller-manager requires hostname to be same as server name.
			// This means server names should not contain the cluster name (which contains "." chars"
			// https://github.com/hetznercloud/hcloud-cloud-controller-manager/blob/f7d624e83c2c3475c5606306214814250922cb8a/hcloud/util.go#L39
			name := ig.Name + "-" + strconv.Itoa(i)
			server := hetznertasks.Server{
				Name:       fi.String(name),
				Lifecycle:  b.Lifecycle,
				SSHKey:     b.LinkToSSHKey(),
				Network:    b.LinkToNetwork(),
				Location:   ig.Spec.Subnets[0],
				Size:       ig.Spec.MachineType,
				Image:      ig.Spec.Image,
				EnableIPv4: true,
				EnableIPv6: false,
				UserData:   userData,
				Labels:     labels,
			}

			c.AddTask(&server)
		}
	}

	return nil
}
