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

package scalewaymodel

import (
	"fmt"
	"os"

	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/scalewaytasks"
)

// InstanceModelBuilder configures instances for the cluster
type InstanceModelBuilder struct {
	*ScwModelContext

	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
}

var _ fi.ModelBuilder = &InstanceModelBuilder{}

func (d *InstanceModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range d.InstanceGroups {
		name := d.AutoscalingGroupName(ig)
		zone := os.Getenv("SCW_DEFAULT_ZONE")

		instance := scalewaytasks.Instance{
			Count:          int(fi.Int32Value(ig.Spec.MinSize)),
			Name:           fi.String(name),
			Lifecycle:      d.Lifecycle,
			Zone:           fi.String(zone),
			CommercialType: fi.String(ig.Spec.MachineType),
			Image:          fi.String(ig.Spec.Image),
			Tags: []string{
				scaleway.TagInstanceGroup + "=" + ig.Name,
				scaleway.TagClusterName + "=" + d.Cluster.Name,
			},
		}

		userData, err := d.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return fmt.Errorf("error building instance task: %w", err)
		}
		instance.UserData = &userData

		c.AddTask(&instance)
	}
	return nil
}
