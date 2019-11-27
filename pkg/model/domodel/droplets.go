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

package domodel

import (
	"strconv"
	"strings"

	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/dotasks"
)

// DropletBuilder configures droplets for the cluster
type DropletBuilder struct {
	*DOModelContext

	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &DropletBuilder{}

func (d *DropletBuilder) Build(c *fi.ModelBuilderContext) error {
	sshKeyName, err := d.SSHKeyName()
	if err != nil {
		return err
	}

	splitSSHKeyName := strings.Split(sshKeyName, "-")
	sshKeyFingerPrint := splitSSHKeyName[len(splitSSHKeyName)-1]

	// replace "." with "-" since DO API does not accept "."
	clusterTag := do.TagKubernetesClusterNamePrefix + ":" + strings.Replace(d.ClusterName(), ".", "-", -1)
	clusterMasterTag := do.TagKubernetesClusterMasterPrefix + ":" + strings.Replace(d.ClusterName(), ".", "-", -1)

	masterIndexCount := 0
	// In the future, DigitalOcean will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of droplets
	for _, ig := range d.InstanceGroups {
		name := d.AutoscalingGroupName(ig)

		var droplet dotasks.Droplet
		droplet.Count = int(fi.Int32Value(ig.Spec.MinSize))
		droplet.Name = fi.String(name)

		// during alpha support we only allow 1 region
		// validation for only 1 region is done at this point
		droplet.Region = fi.String(d.Cluster.Spec.Subnets[0].Region)
		droplet.Size = fi.String(ig.Spec.MachineType)
		droplet.Image = fi.String(ig.Spec.Image)
		droplet.SSHKey = fi.String(sshKeyFingerPrint)

		droplet.Tags = []string{clusterTag}

		if ig.IsMaster() {
			masterIndexCount++
			clusterTagIndex := do.TagKubernetesClusterIndex + ":" + strconv.Itoa(masterIndexCount)
			droplet.Tags = append(droplet.Tags, clusterTagIndex)
			droplet.Tags = append(droplet.Tags, clusterMasterTag)
			droplet.Tags = append(droplet.Tags, do.TagKubernetesClusterInstanceGroupPrefix+":"+"master-"+d.Cluster.Spec.Subnets[0].Region)
		} else {
			droplet.Tags = append(droplet.Tags, do.TagKubernetesClusterInstanceGroupPrefix+":"+"nodes")
		}

		userData, err := d.BootstrapScript.ResourceNodeUp(ig, d.Cluster)
		if err != nil {
			return err
		}
		droplet.UserData = userData

		c.AddTask(&droplet)
	}
	return nil
}
