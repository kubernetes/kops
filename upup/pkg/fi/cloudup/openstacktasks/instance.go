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
	"strings"

	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/schedulerhints"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type Instance struct {
	ID               *string
	Name             *string
	GroupName        *string
	Port             *Port
	Region           *string
	Flavor           *string
	Image            *string
	SSHKey           *string
	ServerGroup      *ServerGroup
	Role             *string
	UserData         fi.Resource
	Metadata         map[string]string
	AvailabilityZone *string
	SecurityGroups   []string
	FloatingIP       *FloatingIP
	ConfigDrive      *bool

	Lifecycle    fi.Lifecycle
	ForAPIServer bool
}

var (
	_ fi.Task            = &Instance{}
	_ fi.HasAddress      = &Instance{}
	_ fi.HasDependencies = &Instance{}
)

// Constants for truncating Tags
const MAX_TAG_LENGTH_OPENSTACK = 60

var TRUNCATE_OPT = truncate.TruncateStringOptions{
	MaxLength:     MAX_TAG_LENGTH_OPENSTACK,
	AlwaysAddHash: false,
	HashLength:    6,
}

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
		deps = append(deps, fi.FindDependencies(tasks, e.UserData)...)
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

func (e *Instance) FindAddresses(context *fi.Context) ([]string, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	if e.Port == nil {
		return nil, nil
	}

	ports, err := cloud.GetPort(fi.StringValue(e.Port.ID))
	if err != nil {
		return nil, err
	}

	for _, port := range ports.FixedIPs {
		return []string{port.IPAddress}, nil
	}

	return nil, nil
}

// filterInstancePorts tries to get all ports of an instance tagged with the cluster name.
// If no tagged ports are found it will return all ports of the instance, to not change the legacy behavior when there weren't tagged ports
func filterInstancePorts(allPorts []ports.Port, clusterName string) []ports.Port {
	clusterNameTag := truncate.TruncateString(fmt.Sprintf("%s=%s", openstack.TagClusterName, clusterName), TRUNCATE_OPT)

	var taggedPorts []ports.Port

	for _, port := range allPorts {
		for _, tag := range port.Tags {
			if tag == clusterNameTag {
				taggedPorts = append(taggedPorts, port)
				break
			}
		}
	}

	if len(taggedPorts) == 0 {
		return allPorts
	}

	return taggedPorts
}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	if e == nil || e.Name == nil {
		return nil, nil
	}
	cloud := c.Cloud.(openstack.OpenstackCloud)
	computeClient := cloud.ComputeClient()
	serverPage, err := servers.List(computeClient, servers.ListOpts{
		Name: fmt.Sprintf("^%s", fi.StringValue(e.GroupName)),
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("error listing servers: %v", err)
	}
	serverList, err := servers.ExtractServers(serverPage)
	if err != nil {
		return nil, fmt.Errorf("error extracting server page: %v", err)
	}

	var filteredList []servers.Server
	for _, server := range serverList {
		val, ok := server.Metadata["k8s"]
		if !ok || val != fi.StringValue(e.ServerGroup.ClusterName) {
			continue
		}
		metadataName := ""
		val, ok = server.Metadata[openstack.TagKopsName]
		if ok {
			metadataName = val
		}
		// name or metadata tag should match to instance name
		// this is needed for backwards compatibility
		if server.Name == fi.StringValue(e.Name) || metadataName == fi.StringValue(e.Name) {
			filteredList = append(filteredList, server)
		}
	}

	if filteredList == nil {
		return nil, nil
	}
	if len(filteredList) > 1 {
		return nil, fmt.Errorf("Multiple servers found with name %s", fi.StringValue(e.Name))
	}

	server := filteredList[0]
	actual := &Instance{
		ID:               fi.String(server.ID),
		Name:             e.Name,
		SSHKey:           fi.String(server.KeyName),
		Lifecycle:        e.Lifecycle,
		Metadata:         server.Metadata,
		Role:             fi.String(server.Metadata["KopsRole"]),
		AvailabilityZone: e.AvailabilityZone,
		GroupName:        e.GroupName,
		ConfigDrive:      e.ConfigDrive,
	}

	ports, err := cloud.ListPorts(ports.ListOpts{
		DeviceID: server.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port for instance %v: %v", server.ID, err)
	}

	ports = filterInstancePorts(ports, fi.StringValue(e.ServerGroup.ClusterName))

	if len(ports) == 1 {
		port := ports[0]
		porttask, err := newPortTaskFromCloud(cloud, e.Lifecycle, &port, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch port for instance %v: %v", server.ID, err)
		}
		actual.Port = porttask

	} else if len(ports) > 1 {
		return nil, fmt.Errorf("found more than one port for instance %v", server.ID)
	}

	if e.FloatingIP != nil && e.Port != nil {
		fips, err := cloud.ListL3FloatingIPs(l3floatingip.ListOpts{
			PortID: fi.StringValue(e.Port.ID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch floating ips for instance %v: %v", server.ID, err)
		}

		if len(fips) == 1 {
			fip := fips[0]
			fipTask := &FloatingIP{
				ID:   fi.String(fip.ID),
				Name: fi.String(fip.Description),
			}

			actual.FloatingIP = fipTask
		} else if len(fips) > 1 {
			return nil, fmt.Errorf("found more than one floating ip for instance %v", server.ID)
		}
	}

	// Avoid flapping
	e.ID = actual.ID
	actual.ForAPIServer = e.ForAPIServer

	// Immutable fields
	actual.Flavor = e.Flavor
	actual.Image = e.Image
	actual.UserData = e.UserData
	actual.Region = e.Region
	actual.SSHKey = e.SSHKey
	actual.ServerGroup = e.ServerGroup

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
	if changes.Port != nil {
		return true, nil
	}
	if changes.FloatingIP != nil {
		return true, nil
	}

	return false, nil
}

// generateInstanceName generates name for the instance
// the instance format is [GroupName]-[6 character hash]
func generateInstanceName(e *Instance) (string, error) {
	secret, err := fi.CreateSecret()
	if err != nil {
		return "", err
	}

	hash, err := secret.AsString()
	if err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%s-%s", fi.StringValue(e.GroupName), hash[0:6])), nil
}

func (_ *Instance) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Instance) error {
	cloud := t.Cloud
	if a == nil {
		serverName, err := generateInstanceName(e)
		if err != nil {
			return err
		}
		klog.V(2).Infof("Creating Instance with name: %q", serverName)

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
			Name:      serverName,
			ImageRef:  image.ID,
			FlavorRef: flavor.ID,
			Networks: []servers.Network{
				{
					Port: fi.StringValue(e.Port.ID),
				},
			},
			Metadata:       e.Metadata,
			SecurityGroups: e.SecurityGroups,
			ConfigDrive:    e.ConfigDrive,
		}
		if e.UserData != nil {
			bytes, err := fi.ResourceAsBytes(e.UserData)
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

		v, err := t.Cloud.CreateInstance(opts, fi.StringValue(e.Port.ID))
		if err != nil {
			return fmt.Errorf("Error creating instance: %v", err)
		}
		e.ID = fi.String(v.ID)
		e.ServerGroup.AddNewMember(fi.StringValue(e.ID))

		if e.FloatingIP != nil {
			err = associateFloatingIP(t, e)
			if err != nil {
				return err
			}
		}

		klog.V(2).Infof("Creating a new Openstack instance, id=%s", v.ID)

		return nil
	}
	if changes.Port != nil {
		ports.Update(cloud.NetworkingClient(), fi.StringValue(changes.Port.ID), ports.UpdateOpts{
			DeviceID: e.ID,
		})
	}
	if changes.FloatingIP != nil {
		err := associateFloatingIP(t, e)
		if err != nil {
			return err
		}
	}
	return nil
}

func associateFloatingIP(t *openstack.OpenstackAPITarget, e *Instance) error {
	client := t.Cloud.NetworkingClient()

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
