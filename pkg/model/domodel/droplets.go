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

	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
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
	clusterName := do.SafeClusterName(d.ClusterName())
	clusterTag := do.TagKubernetesClusterNamePrefix + ":" + clusterName
	clusterMasterTag := do.TagKubernetesClusterMasterPrefix + ":" + clusterName

	masterIndexCount := 0
	// In the future, DigitalOcean will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of droplets
	for _, ig := range d.InstanceGroups {
		name := d.AutoscalingGroupName(ig)

		droplet := dotasks.Droplet{
			Count:     int(fi.Int32Value(ig.Spec.MinSize)),
			Name:      fi.String(name),
			Lifecycle: d.Lifecycle,

			// kops do supports allow only 1 region
			Region: fi.String(d.Cluster.Spec.Subnets[0].Region),
			Size:   fi.String(ig.Spec.MachineType),
			Image:  fi.String(ig.Spec.Image),
			SSHKey: fi.String(sshKeyFingerPrint),
			Tags:   []string{clusterTag},
		}

		if ig.IsMaster() {
			masterIndexCount++
			// create tag based on etcd name. etcd name is now prefixed with etcd-
			// Ref: https://github.com/kubernetes/kops/commit/31f8cbd571964f19d3c31024ddba918998d29929
			clusterTagIndex := do.TagKubernetesClusterIndex + ":" + "etcd-" + strconv.Itoa(masterIndexCount)
			droplet.Tags = append(droplet.Tags, clusterTagIndex)
			droplet.Tags = append(droplet.Tags, clusterMasterTag)
			droplet.Tags = append(droplet.Tags, do.TagKubernetesInstanceGroup+":"+ig.Name)
		} else {
			droplet.Tags = append(droplet.Tags, do.TagKubernetesInstanceGroup+":"+ig.Name)
		}

		if d.Cluster.Spec.NetworkID != "" {
			droplet.VPCUUID = fi.String(d.Cluster.Spec.NetworkID)
		} else if d.Cluster.Spec.NetworkCIDR != "" {
			// since networkCIDR specified as part of the request, it is made sure that vpc with this cidr exist before
			// creating the droplet, so you can associate with vpc uuid for this droplet.
			vpcName := "vpc-" + clusterName
			droplet.VPCName = fi.String(vpcName)
			droplet.NetworkCIDR = fi.String(d.Cluster.Spec.NetworkCIDR)
		}

		userData, err := d.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return err
		}
		droplet.UserData = userData

		c.AddTask(&droplet)
	}
	return nil
}
