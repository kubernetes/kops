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

	"github.com/scaleway/scaleway-sdk-go/scw"
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

var _ fi.CloudupModelBuilder = &InstanceModelBuilder{}

func (d *InstanceModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	for _, ig := range d.InstanceGroups {
		name := ig.Name
		zone, err := scw.ParseZone(ig.Spec.Subnets[0])
		if err != nil {
			return fmt.Errorf("error building instance task for %q: %w", name, err)
		}

		userData, err := d.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return fmt.Errorf("error building bootstrap script for %q: %w", name, err)
		}

		instance := scalewaytasks.Instance{
			Count:          int(fi.ValueOf(ig.Spec.MinSize)),
			Name:           fi.PtrTo(name),
			Lifecycle:      d.Lifecycle,
			Zone:           fi.PtrTo(string(zone)),
			CommercialType: fi.PtrTo(ig.Spec.MachineType),
			Image:          fi.PtrTo(ig.Spec.Image),
			UserData:       &userData,
			Tags: []string{
				scaleway.TagInstanceGroup + "=" + ig.Name,
				scaleway.TagClusterName + "=" + d.Cluster.Name,
			},
		}

		if ig.IsControlPlane() {
			instance.Tags = append(instance.Tags, scaleway.TagNameRolePrefix+"="+scaleway.TagRoleControlPlane)
			instance.Role = fi.PtrTo(scaleway.TagRoleControlPlane)
		} else {
			instance.Role = fi.PtrTo(scaleway.TagRoleNode)
		}

		c.AddTask(&instance)
	}
	return nil
}
