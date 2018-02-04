/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package imagebuilder

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"k8s.io/kube-deploy/imagebuilder/pkg/imagebuilder/executor"
)

// TODO: We should replace most of this code with a fast-install manifest
// This would also allow more customization, and get rid of half of this code
// BUT... there's a circular dependency in the PRs here... :-)

// GCEInstance manages an GCE instance, used for building an image
type GCEInstance struct {
	cloud    *GCECloud
	name     string
	instance *compute.Instance
}

// Shutdown terminates the running instance
func (i *GCEInstance) Shutdown() error {
	glog.Infof("Terminating instance %q", i.name)
	return i.cloud.deleteInstance(i.name)
}

// DialSSH establishes an SSH client connection to the instance
func (i *GCEInstance) DialSSH(config *ssh.ClientConfig) (executor.Executor, error) {
	publicIP, err := i.WaitPublicIP()
	if err != nil {
		return nil, err
	}

	for {
		// TODO: Timeout, check error code
		sshClient, err := ssh.Dial("tcp", publicIP+":22", config)
		if err != nil {
			glog.Warningf("error connecting to SSH on server %q: %v", publicIP, err)
			time.Sleep(5 * time.Second)
			continue
			//	return nil, fmt.Errorf("error connecting to SSH on server %q", publicIP)
		}

		return executor.NewSSH(sshClient), nil
	}
}

// WaitPublicIP waits for the instance to get a public IP, returning it
func (i *GCEInstance) WaitPublicIP() (string, error) {
	// TODO: Timeout
	for {
		instance, err := i.cloud.describeInstance(i.name)
		if err != nil {
			return "", err
		}

		for _, ni := range instance.NetworkInterfaces {
			for _, ac := range ni.AccessConfigs {
				if ac.NatIP != "" {
					glog.Infof("Instance public IP is %q", ac.NatIP)
					return ac.NatIP, nil
				}
			}
		}
		glog.V(2).Infof("Sleeping before requerying instance for public IP: %q", i.name)
		time.Sleep(5 * time.Second)
	}
}

// GCECloud is a helper type for talking to an GCE acccount
type GCECloud struct {
	config *GCEConfig

	computeClient *compute.Service
}

var _ Cloud = &GCECloud{}

func NewGCECloud(computeClient *compute.Service, config *GCEConfig) *GCECloud {
	return &GCECloud{
		computeClient: computeClient,
		config:        config,
	}
}

func (a *GCECloud) GetExtraEnv() (map[string]string, error) {
	// No extra env needed on GCE
	env := make(map[string]string)
	return env, nil
}

func IsGCENotFound(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	return apiErr.Code == 404
}

func (c *GCECloud) describeInstance(name string) (*compute.Instance, error) {
	glog.V(2).Infof("GCE Instances List Name=%q", name)
	instances, err := c.computeClient.Instances.List(c.config.Project, c.config.Zone).Filter("name eq " + name).Do()
	if err != nil {
		return nil, fmt.Errorf("error making GCE Instances List call: %v", err)
	}

	if len(instances.Items) == 0 {
		return nil, nil
	}
	if len(instances.Items) != 1 {
		return nil, fmt.Errorf("found multiple instances with name %q", name)
	}
	return instances.Items[0], nil
}

// deleteInstance terminates the specified instance
func (c *GCECloud) deleteInstance(name string) error {
	glog.V(2).Infof("GCE Delete Instances name=%q", name)
	_, err := c.computeClient.Instances.Delete(c.config.Project, c.config.Zone, name).Do()
	if err != nil {
		return fmt.Errorf("error terminating instance %q: %v", name, err)
	}
	return nil
}

// GetInstance returns the GCE instance matching our tags, or nil if not found
func (c *GCECloud) GetInstance() (Instance, error) {
	name := c.config.MachineName

	instance, err := c.describeInstance(name)
	if err != nil {
		return nil, err
	}

	if instance != nil {
		glog.Infof("Found existing instance: %q", instance.Name)
		return &GCEInstance{
			cloud:    c,
			instance: instance,
			name:     instance.Name,
		}, nil
	}

	return nil, nil
}

// CreateInstance creates an instance for building an image instance
func (c *GCECloud) CreateInstance() (Instance, error) {
	name := c.config.MachineName
	zone := c.config.Zone

	machineType := "zones/" + zone + "/machineTypes/" + c.config.MachineType
	glog.Infof("creating instance with machinetype %s", machineType)

	var disks []*compute.AttachedDisk
	disks = append(disks, &compute.AttachedDisk{
		InitializeParams: &compute.AttachedDiskInitializeParams{
			SourceImage: c.config.Image,
			DiskType:    "zones/" + zone + "/diskTypes/pd-ssd",
		},
		Boot:       true,
		DeviceName: "disk-0",
		Index:      0,
		AutoDelete: true,
		Mode:       "READ_WRITE",
		Type:       "PERSISTENT",
	})

	metadata := &compute.Metadata{}

	if c.config.SSHPublicKey != "" {
		publicKey, err := ReadFile(c.config.SSHPublicKey)
		if err != nil {
			return nil, err
		}
		sshKey := "admin:" + string(publicKey)

		metadata.Items = append(metadata.Items, &compute.MetadataItems{
			Key:   "ssh-keys",
			Value: &sshKey,
		})
	}

	scopes := []string{
		"https://www.googleapis.com/auth/devstorage.read_write",
		"https://www.googleapis.com/auth/compute",
	}

	instance := &compute.Instance{
		Name: name,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "nat",
						Type: "ONE_TO_ONE_NAT",
					},
				},
			},
		},
		MachineType: machineType,
		Disks:       disks,
		Metadata:    metadata,
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email:  "default",
				Scopes: scopes,
			},
		},
	}
	_, err := c.computeClient.Instances.Insert(c.config.Project, c.config.Zone, instance).Do()
	if err != nil {
		return nil, fmt.Errorf("error running instance: %v", err)
	}
	return &GCEInstance{
		cloud: c,
		name:  name,
	}, nil
}

// FindImage finds a registered image, matching by the name tag
func (c *GCECloud) FindImage(imageName string) (Image, error) {
	image, err := findGCEImage(c.computeClient, c.config.Project, imageName)
	if err != nil {
		return nil, err
	}

	if image == nil {
		return nil, nil
	}

	return &GCEImage{
		computeClient: c.computeClient,
		name:          imageName,
		//image:   image,
	}, nil
}

func findGCEImage(computeClient *compute.Service, project string, imageName string) (*compute.Image, error) {
	images, err := computeClient.Images.List(project).Filter("name eq " + imageName).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing images: %v", err)
	}

	glog.V(2).Infof("GCE Images List Filter:Name=%q", imageName)

	if len(images.Items) == 0 {
		return nil, nil
	}

	if len(images.Items) != 1 {
		return nil, fmt.Errorf("found multiple matching images for name: %q", imageName)
	}

	return images.Items[0], nil
}

// GCEImage represents an image on GCE
type GCEImage struct {
	computeClient *compute.Service
	name          string
}

var _ Image = &GCEImage{}

// String returns a string representation of the image
func (i *GCEImage) String() string {
	return "GCEImage[" + i.name + "]"
}

// EnsurePublic makes the image accessible outside the current account
func (i *GCEImage) EnsurePublic() error {
	return fmt.Errorf("GCE does not currently support public images")
}

// AddTags adds the specified tags on the image
func (i *GCEImage) AddTags(tags map[string]string) error {
	return fmt.Errorf("Tagging of GCE images not yet implemented")
}

// ReplicateImage copies the image to all accessible GCE regions
func (i *GCEImage) ReplicateImage(makePublic bool) (map[string]Image, error) {
	if makePublic {
		return nil, fmt.Errorf("GCE does not currently support public images")
	}

	images := make(map[string]Image)
	// All images are already global
	return images, nil
}
