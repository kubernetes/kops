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

package domodel

import (
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/dotasks"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*DOModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {

	ipRange := b.Cluster.Spec.NetworkCIDR
	if ipRange == "" {
		// no cidr specified, use the default vpc in DO that's always available
		return nil
	}

	clusterName := strings.Replace(b.ClusterName(), ".", "-", -1)
	vpcName := "vpc-" + clusterName

	// Create a separate vpc for this cluster.
	vpc := &dotasks.VPC{
		Name:      fi.String(vpcName),
		Region:    fi.String(b.Cluster.Spec.Subnets[0].Region),
		Lifecycle: b.Lifecycle,
		IPRange:   fi.String(ipRange),
	}
	c.AddTask(vpc)

	return nil
}
