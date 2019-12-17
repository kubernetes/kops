/*
Copyright 2017 The Kubernetes Authors.

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

package vspheremodel

// autoscalinggroup is a model for vSphere cloud. It's responsible for building tasks, necessary for kubernetes cluster deployment.

import (
	"strconv"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vspheretasks"
)

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*VSphereModelContext

	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

// Build creates tasks related to cluster deployment and adds them to ModelBuilderContext.
func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// Note that we are creating a VM per instance group. Instance group represents a group of VMs.
	// The following logic should considerably change once we add support for multiple master/worker nodes,
	// cloud-init etc.
	for _, ig := range b.InstanceGroups {
		instanceCount := int(fi.Int32Value(ig.Spec.MinSize))
		if ig.Spec.Role == kops.InstanceGroupRoleMaster {
			instanceCount = 1
		}
		for i := 1; i <= instanceCount; i++ {
			name := b.InstanceName(ig, strconv.Itoa(i))
			createVmTask := &vspheretasks.VirtualMachine{
				Name:           &name,
				VMTemplateName: fi.String(ig.Spec.Image),
			}

			c.AddTask(createVmTask)

			attachISOTaskName := "AttachISO-" + name
			attachISOTask := &vspheretasks.AttachISO{
				Name:            &attachISOTaskName,
				VM:              createVmTask,
				IG:              ig,
				BootstrapScript: b.BootstrapScript,
				Cluster:         b.Cluster,
			}

			c.AddTask(attachISOTask)

			powerOnTaskName := "PowerON-" + name
			powerOnTask := &vspheretasks.VMPowerOn{
				Name:      &powerOnTaskName,
				AttachISO: attachISOTask,
			}
			c.AddTask(powerOnTask)
		}
	}
	return nil
}
