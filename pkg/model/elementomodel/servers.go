/*
Copyright 2025 The Kubernetes Authors.

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

package elementomodel

import (
	"fmt"
	"io"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
	"k8s.io/kops/upup/pkg/fi/cloudup/elementotasks"
)

// ServerGroupModelBuilder configures server objects
type ServerGroupModelBuilder struct {
	*ElementoModelContext
	Lifecycle              fi.Lifecycle
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
}

var _ fi.CloudupModelBuilder = &ServerGroupModelBuilder{}

func (b *ServerGroupModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	var sshkeyTasks []*elementotasks.SSHKey
	for _, sshkey := range b.SSHPublicKeys {
		fingerprint, err := pki.ComputeOpenSSHKeyFingerprint(string(sshkey))
		if err != nil {
			return err
		}
		t := &elementotasks.SSHKey{
			Name:      fi.PtrTo(b.ClusterName() + "-" + fingerprint),
			Lifecycle: b.Lifecycle,
			PublicKey: string(sshkey),
			Labels: map[string]string{
				elemento.TagKubernetesClusterName: b.ClusterName(),
			},
		}
		c.AddTask(t)
		sshkeyTasks = append(sshkeyTasks, t)
	}

	for _, ig := range b.InstanceGroups {
		igSize := fi.ValueOf(ig.Spec.MinSize)
		labels, err := b.CloudTagsForInstanceGroup(ig)
		if err != nil {
			return err
		}
		labels[elemento.TagKubernetesClusterName] = b.ClusterName()
		labels[elemento.TagKubernetesInstanceGroup] = ig.Name
		labels[elemento.TagKubernetesInstanceRole] = string(ig.Spec.Role)

		userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return err
		}

		// For debugging: wrap the userData to print it when it's ready
		if userData != nil {
			userData = &debugUserDataResource{
				Resource:      userData,
				instanceGroup: ig.Name,
			}
		}

		fmt.Printf("CREATING server group for instance group %q with size %d\n", ig.Name, igSize)
		fmt.Printf("--- End of UserData ---\n")

		// Determine root volume size
		var rootVolumeSize *int32
		if ig.Spec.RootVolume != nil && ig.Spec.RootVolume.Size != nil {
			rootVolumeSize = ig.Spec.RootVolume.Size
		} else {
			// Use default volume size based on role
			defaultSize, err := defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
			if err != nil {
				return err
			}
			rootVolumeSize = fi.PtrTo(defaultSize)
		}

		serverGroup := elementotasks.ServerGroup{
			Name:           fi.PtrTo(ig.Name),
			Lifecycle:      b.Lifecycle,
			SSHKeys:        sshkeyTasks,
			Network:        b.LinkToNetwork(),
			Count:          int(igSize),
			Location:       ig.Spec.Subnets[0],
			Size:           ig.Spec.MachineType,
			Image:          ig.Spec.Image,
			Architecture:   determineArchitecture(ig),
			EnableIPv4:     true,
			EnableIPv6:     false,
			UserData:       userData,
			Labels:         labels,
			RootVolumeSize: rootVolumeSize,
		}
		if b.Cluster.PublishesDNSRecords() {
			serverGroup.ClusterName = fi.PtrTo(b.ClusterName())
			serverGroup.DNSZone = fi.PtrTo(b.ClusterName())
			if ig.HasAPIServer() {
				if !b.UseLoadBalancerForAPI() {
					apiPublicName := b.Cluster.Spec.API.PublicName
					if apiPublicName == "" {
						apiPublicName = "api." + b.ClusterName()
					}
					serverGroup.APIPublicName = fi.PtrTo(apiPublicName)
				}
				if !b.UseLoadBalancerForInternalAPI() {
					serverGroup.APIInternalName = fi.PtrTo(b.Cluster.APIInternalName())
				}
				serverGroup.KopsControllerInternalName = fi.PtrTo("kops-controller.internal." + b.ClusterName())
			}
		}

		c.AddTask(&serverGroup)
	}

	return nil
}

// determines the appropriate architecture for an instance group
func determineArchitecture(ig *kops.InstanceGroup) string {
	// Check if architecture is explicitly specified in the instance group
	if ig.Spec.Architecture != "" {
		return ig.Spec.Architecture
	}

	// Default to X86_64 for Elemento cloud provider
	return "X86_64"
}

// debugUserDataResource is a wrapper that prints userData content when it's accessed
type debugUserDataResource struct {
	fi.Resource
	instanceGroup string
	printed       bool
}

func (d *debugUserDataResource) Open() (io.Reader, error) {
	reader, err := d.Resource.Open()
	if err != nil {
		return reader, err
	}

	// Print the content only once when it's first accessed
	if !d.printed {
		d.printed = true
		content, readErr := fi.ResourceAsString(d.Resource)
		if readErr == nil {
			fmt.Printf("=== UserData for instance group %q ===\n", d.instanceGroup)
			fmt.Printf("%s\n", content)
			fmt.Printf("=== End UserData for %q ===\n", d.instanceGroup)
		}
	}

	return reader, err
}
