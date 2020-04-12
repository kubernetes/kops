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

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/schedulerhints"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Instance
type Instance struct {
	ID               *string
	Name             *string
	NamePrefix       *string
	Port             *Port
	Region           *string
	Flavor           *string
	Image            *string
	SSHKey           *string
	ServerGroup      *ServerGroup
	Tags             []string
	Role             *string
	UserData         *string
	Metadata         map[string]string
	AvailabilityZone *string
	SecurityGroups   []string
	Lifecycle        *fi.Lifecycle
}

var _ fi.HasAddress = &Instance{}

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
	serverPage, err := servers.List(c.Cloud.(openstack.OpenstackCloud).ComputeClient(), servers.ListOpts{
		Name: fmt.Sprintf("^%s", fi.StringValue(e.NamePrefix)),
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
		_, detachTag := server.Metadata[openstack.TagNameDetach]
		// name or metadata tag should match to instance name
		// this is needed for backwards compatibility
		if server.Name == fi.StringValue(e.Name) || metadataName == fi.StringValue(e.Name) {
			if !detachTag {
				filteredList = append(filteredList, server)
			}
		}
	}
	serverList = nil
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
		AvailabilityZone: e.AvailabilityZone,
		NamePrefix:       e.NamePrefix,
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
	return a == nil, nil
}

// makeServerName generates name for the instance
// the instance format is [namePrefix]-[6 character hash]
func makeServerName(e *Instance) (string, error) {
	secret, err := fi.CreateSecret()
	if err != nil {
		return "", err
	}

	hash, err := secret.AsString()
	if err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%s-%s", fi.StringValue(e.NamePrefix), hash[0:5])), nil
}

func (_ *Instance) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Instance) error {
	if a == nil {
		e.Metadata[openstack.TagKopsName] = fi.StringValue(e.Name)
		serverName, err := makeServerName(e)
		if err != nil {
			return err
		}
		klog.V(2).Infof("Creating Instance with name: %q", serverName)
		opt := servers.CreateOpts{
			Name:       serverName,
			ImageName:  fi.StringValue(e.Image),
			FlavorName: fi.StringValue(e.Flavor),
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
			opt.UserData = []byte(*e.UserData)
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

		klog.V(2).Infof("Creating a new Openstack instance, id=%s", v.ID)

		return nil
	}

	klog.V(2).Infof("Openstack task Instance::RenderOpenstack did nothing")
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
