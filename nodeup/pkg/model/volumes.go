/*
Copyright 2018 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"github.com/golang/glog"
)

// VolumesBuilder maintains the volume mounting
type VolumesBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &VolumesBuilder{}

// Build is responsible for handling the mounting additional volumes onto the instance
func (b *VolumesBuilder) Build(c *fi.ModelBuilderContext) error {
	// @step: check if the instancegroup has any volumes to mount
	ig := b.InstanceGroup.Spec
	if len(ig.Volumes) <= 0 {
		glog.V(1).Info("skipping the volume builder, no volumes defined for this instancegroup")

		return nil
	}

	for _, x := range ig.Volumes {
		// @check does the volume have a filesystem specification?
		if x.Filesystem == nil {
			continue
		}
		device := fi.StringValue(x.DeviceName)
		mountpath := x.Filesystem.Path

		glog.V(2).Infof("attempting to provision device: %s, path: %s", device, mountpath)

		// @step: create the formatting unit
		fsvc, err := buildDeviceFormatService(c, x)
		if err != nil {
			return fmt.Errorf("failed to provision format service for device: %s, error: %s", device, err)
		}
		c.AddTask(fsvc)

		msvc, err := buildDeviceMountService(c, x, fsvc.Name)
		if err != nil {
			return fmt.Errorf("failed to provision format service for device: %s, error: %s", device, err)
		}
		c.AddTask(msvc)
	}

	return nil
}

// buildDeviceFormatService is responsible for constructing the systemd unit to format the device
func buildDeviceFormatService(c *fi.ModelBuilderContext, volume *kops.InstanceGroupVolumeSpec) (*nodetasks.Service, error) {
	device := fi.StringValue(volume.DeviceName)
	deviceFmt := strings.Replace(device, "/", "-", -1)
	name := fmt.Sprintf("format%s.service", deviceFmt)

	u := &systemd.Manifest{}
	u.Set("Unit", "Description", fmt.Sprintf("Formats the device: %s", device))
	u.Set("Unit", "After", fmt.Sprintf("%s.device", deviceFmt))
	u.Set("Unit", "Requires", fmt.Sprintf("%s.device", deviceFmt))
	u.Set("Service", "Type", "oneshot")
	u.Set("Service", "RemainAfterExit", "yes")
	u.Set("Service", "ExecStart", fmt.Sprintf("/usr/bin/bash -c '/usr/sbin/blkid %s || (/usr/sbin/wipefs -f %s && /usr/sbin/mkfs.ext4 %s)'", device, device, device))

	return &nodetasks.Service{
		Name:         name,
		Definition:   s(u.Render()),
		Enabled:      fi.Bool(true),
		ManageState:  fi.Bool(true),
		Running:      fi.Bool(true),
		SmartRestart: fi.Bool(true),
	}, nil
}

// buildDeviceMountService is responsible for building the mount service
func buildDeviceMountService(c *fi.ModelBuilderContext, volume *kops.InstanceGroupVolumeSpec, formatName string) (*nodetasks.Service, error) {
	device := fi.StringValue(volume.DeviceName)
	mountpath := volume.Filesystem.Path
	name := fmt.Sprintf("%s.mount", strings.TrimPrefix(strings.Replace(mountpath, "/", "-", -1), "-"))

	// @step: create the mounting unit
	u := &systemd.Manifest{}
	u.Set("Unit", "Description", fmt.Sprintf("Mounting volume: %s from device: %s", mountpath, device))
	u.Set("Unit", "Requires", formatName)
	u.Set("Unit", "After", formatName)
	u.Set("Unit", "Before", "docker.service")
	u.Set("Mount", "What", device)
	u.Set("Mount", "Where", mountpath)
	u.Set("Mount", "Type", "ext4")

	return &nodetasks.Service{
		Name:         name,
		Definition:   s(u.Render()),
		Enabled:      fi.Bool(true),
		ManageState:  fi.Bool(true),
		Running:      fi.Bool(true),
		SmartRestart: fi.Bool(true),
		OnChangeExecute: [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "restart", "docker.service"},
			{"systemctl", "restart", "kops-configuration.service", "&"},
		},
	}, nil
}
