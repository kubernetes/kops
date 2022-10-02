/*
Copyright 2021 The Kubernetes Authors.

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

package dns

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks/dnstasks"
)

// GossipBuilder seeds some hostnames into /etc/hosts, avoiding some circular dependencies.
type GossipBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &GossipBuilder{}

// Build is responsible for configuring the gossip DNS tasks.
func (b *GossipBuilder) Build(c *fi.ModelBuilderContext) error {
	useGossip := dns.IsGossipCluster(b.Cluster)
	if !useGossip {
		return nil
	}

	if b.IsMaster {
		task := &dnstasks.UpdateEtcHostsTask{
			Name: "control-plane-bootstrap",
		}

		if b.Cluster.Spec.MasterInternalName != "" {
			task.Records = append(task.Records, dnstasks.HostRecord{
				Hostname:  b.Cluster.Spec.MasterInternalName,
				Addresses: []string{"127.0.0.1"},
			})
		}
		if b.Cluster.Spec.MasterPublicName != "" {
			task.Records = append(task.Records, dnstasks.HostRecord{
				Hostname:  b.Cluster.Spec.MasterPublicName,
				Addresses: []string{"127.0.0.1"},
			})
		}

		if len(task.Records) != 0 {
			c.AddTask(task)
		}
	}

	return nil
}
