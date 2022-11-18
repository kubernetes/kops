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
	"fmt"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

// LoadBalancerModelBuilder configures Firewall objects
type LoadBalancerModelBuilder struct {
	*HetznerModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &LoadBalancerModelBuilder{}

func (b *LoadBalancerModelBuilder) Build(c *fi.ModelBuilderContext) error {
	controlPlaneLabelSelector := []string{
		fmt.Sprintf("%s=%s", hetzner.TagKubernetesClusterName, b.ClusterName()),
		fmt.Sprintf("%s=%s", hetzner.TagKubernetesInstanceRole, string(kops.InstanceGroupRoleMaster)),
	}
	loadbalancer := hetznertasks.LoadBalancer{
		Name:      fi.PtrTo("api." + b.ClusterName()),
		Lifecycle: b.Lifecycle,
		Network:   b.LinkToNetwork(),
		Location:  b.InstanceGroups[0].Spec.Subnets[0],
		Type:      "lb11",
		Services: []*hetznertasks.LoadBalancerService{
			{
				Protocol:        string(hcloud.LoadBalancerServiceProtocolTCP),
				ListenerPort:    fi.PtrTo(wellknownports.KubeAPIServer),
				DestinationPort: fi.PtrTo(wellknownports.KubeAPIServer),
			},
		},
		Target: strings.Join(controlPlaneLabelSelector, ","),
		Labels: map[string]string{
			hetzner.TagKubernetesClusterName: b.ClusterName(),
		},
	}

	if b.Cluster.UsesNoneDNS() {
		loadbalancer.Services = append(loadbalancer.Services, &hetznertasks.LoadBalancerService{
			Protocol:        string(hcloud.LoadBalancerServiceProtocolTCP),
			ListenerPort:    fi.PtrTo(wellknownports.KopsControllerPort),
			DestinationPort: fi.PtrTo(wellknownports.KopsControllerPort),
		})
	}

	c.AddTask(&loadbalancer)

	return nil
}
