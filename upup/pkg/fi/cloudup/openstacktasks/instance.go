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

package openstacktasks

import (
	"fmt"
	"strconv"

	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/schedulerhints"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type Instance struct {
	ID               *string
	Name             *string
	Port             *Port
	Region           *string
	Flavor           *string
	Image            *string
	SSHKey           *string
	ServerGroup      *ServerGroup
	Role             *string
	UserData         *fi.ResourceHolder
	Metadata         map[string]string
	AvailabilityZone *string
	SecurityGroups   []string
	FloatingIP       *FloatingIP

	Lifecycle    *fi.Lifecycle
	ForAPIServer bool
}

var _ fi.Task = &Instance{}
var _ fi.HasAddress = &Instance{}
var _ fi.HasDependencies = &Instance{}

// GetDependencies returns the dependencies of the Instance task
func (e *Instance) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*ServerGroup); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*Port); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*FloatingIP); ok {
			deps = append(deps, task)
		}
	}

	if e.UserData != nil {
		deps = append(deps, e.UserData.GetDependencies(tasks)...)
	}

	return deps
}

var _ fi.CompareWithID = &Instance{}

func (e *Instance) WaitForStatusActive(t *openstack.OpenstackAPITarget) error {
	return servers.WaitForStatus(t.Cloud.ComputeClient(), *e.ID, "ACTIVE", 120)
}

func (e *Instance) CompareWithID() *string {
	return e.ID
}

func (e *Instance) IsForAPIServer() bool {
	return e.ForAPIServer
}

func (e *Instance) FindIPAddress(context *fi.Context) (*string, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	if e.Port == nil {
		return nil, nil
	}

	ports, err := cloud.GetPort(fi.StringValue(e.Port.ID))
	if err != nil {
		return nil, err
	}

	for _, port := range ports.FixedIPs {
		return fi.String(port.IPAddress), nil
	}

	return nil, nil
}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	if e == nil || e.Name == nil {
		return nil, nil
	}
	client := c.Cloud.(openstack.OpenstackCloud).ComputeClient()
	serverPage, err := servers.List(client, servers.ListOpts{
		Name: fmt.Sprintf("^%s$", fi.StringValue(e.Name)),
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("error finding server with name %s: %v", fi.StringValue(e.Name), err)
	}
	serverList, err := servers.ExtractServers(serverPage)
	if err != nil {
		return nil, fmt.Errorf("error extracting server page: %v", err)
	}
	if len(serverList) == 0 {
		return nil, nil
	}
	if len(serverList) > 1 {
		return nil, fmt.Errorf("Multiple servers found with name %s", fi.StringValue(e.Name))
	}

	server := serverList[0]
	actual := &Instance{
		ID:               fi.String(server.ID),
		Name:             fi.String(server.Name),
		SSHKey:           fi.String(server.KeyName),
		Lifecycle:        e.Lifecycle,
		AvailabilityZone: e.AvailabilityZone,
	}
	e.ID = actual.ID

	return actual, nil
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *Instance) ShouldCreate(a, e, changes *Instance) (bool, error) {
	if a == nil {
		return true, nil
	}

	return false, nil
}

func (_ *Instance) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Instance) error {
	if a == nil {
		klog.V(2).Infof("Creating Instance with name: %q", fi.StringValue(e.Name))
		cloud := t.Cloud.(openstack.OpenstackCloud)

		imageName := fi.StringValue(e.Image)
		image, err := cloud.GetImage(imageName)
		if err != nil {
			return fmt.Errorf("failed to find image %v: %v", imageName, err)
		}

		flavorName := fi.StringValue(e.Flavor)
		flavor, err := cloud.GetFlavor(flavorName)
		if err != nil {
			return fmt.Errorf("failed to find flavor %v: %v", flavorName, err)
		}

		opt := servers.CreateOpts{
			Name:      fi.StringValue(e.Name),
			ImageRef:  image.ID,
			FlavorRef: flavor.ID,
			Networks: []servers.Network{
				{
					Port: fi.StringValue(e.Port.ID),
				},
			},
			Metadata:       e.Metadata,
			ServiceClient:  t.Cloud.ComputeClient(),
			SecurityGroups: e.SecurityGroups,
		}
		if e.UserData != nil {
			bytes, err := e.UserData.AsBytes()
			if err != nil {
				return err
			}
			opt.UserData = bytes
		}
		if e.AvailabilityZone != nil {
			opt.AvailabilityZone = fi.StringValue(e.AvailabilityZone)
		}
		keyext := keypairs.CreateOptsExt{
			CreateOptsBuilder: opt,
			KeyName:           openstackKeyPairName(fi.StringValue(e.SSHKey)),
		}

		sgext := schedulerhints.CreateOptsExt{
			CreateOptsBuilder: keyext,
			SchedulerHints: &schedulerhints.SchedulerHints{
				Group: *e.ServerGroup.ID,
			},
		}

		opts, err := includeBootVolumeOptions(t, e, sgext)
		if err != nil {
			return err
		}

		v, err := t.Cloud.CreateInstance(opts)
		if err != nil {
			return fmt.Errorf("Error creating instance: %v", err)
		}
		e.ID = fi.String(v.ID)
		e.ServerGroup.Members = append(e.ServerGroup.Members, fi.StringValue(e.ID))

		if e.FloatingIP != nil {
			err = associateFloatingIP(t, e)
		}
		if err != nil {
			return err
		}

		klog.V(2).Infof("Creating a new Openstack instance, id=%s", v.ID)

	}

	return nil
}

func associateFloatingIP(t *openstack.OpenstackAPITarget, e *Instance) error {
	cloud := t.Cloud.(openstack.OpenstackCloud)
	client := cloud.NetworkingClient()

	_, err := l3floatingip.Update(client, fi.StringValue(e.FloatingIP.ID), l3floatingip.UpdateOpts{
		PortID: e.Port.ID,
	}).Extract()

	if err != nil {
		return fmt.Errorf("failed to associated floating IP to instance %s: %v", *e.Name, err)
	}
	return nil
}

func includeBootVolumeOptions(t *openstack.OpenstackAPITarget, e *Instance, opts servers.CreateOptsBuilder) (servers.CreateOptsBuilder, error) {
	if !bootFromVolume(e.Metadata) {
		return opts, nil
	}

	i, err := t.Cloud.GetImage(fi.StringValue(e.Image))
	if err != nil {
		return nil, fmt.Errorf("Error getting image information: %v", err)
	}

	bfv := bootfromvolume.CreateOptsExt{
		CreateOptsBuilder: opts,
		BlockDevice: []bootfromvolume.BlockDevice{{
			BootIndex:           0,
			DeleteOnTermination: true,
			DestinationType:     "volume",
			SourceType:          "image",
			UUID:                i.ID,
			VolumeSize:          i.MinDiskGigabytes,
		}},
	}

	if s, ok := e.Metadata[openstack.BOOT_VOLUME_SIZE]; ok {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid value for %v: %v", openstack.BOOT_VOLUME_SIZE, err)
		}

		bfv.BlockDevice[0].VolumeSize = int(i)
	}

	return bfv, nil
}

func bootFromVolume(m map[string]string) bool {
	v, ok := m[openstack.BOOT_FROM_VOLUME]
	if !ok {
		return false
	}

	switch v {
	case "true", "enabled":
		return true
	default:
		return false
	}
}
