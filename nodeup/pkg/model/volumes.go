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
	if !b.UseVolumeMounts() {
		glog.V(1).Info("skipping the volume builder, no volumes defined for this instancegroup")

		return nil
	}

	var mountUnits []string

	// @step: iterate the volume mounts and create the format and mount units
	for _, x := range b.InstanceGroup.Spec.VolumeMounts {
		glog.V(2).Infof("attempting to provision device: %s, path: %s", x.Device, x.Path)

		// @step: create the formatting unit
		fsvc, err := buildDeviceFormatService(c, x)
		if err != nil {
			return fmt.Errorf("failed to provision format service for device: %s, error: %s", x.Device, err)
		}
		c.AddTask(fsvc)

		msvc, err := buildDeviceMountService(c, x, fsvc.Name)
		if err != nil {
			return fmt.Errorf("failed to provision format service for device: %s, error: %s", x.Device, err)
		}
		c.AddTask(msvc)

		mountUnits = append(mountUnits, msvc.Name)
	}

	// @step: create a unit for restart the docker daemon once everything is mounted
	u := &systemd.Manifest{}
	u.Set("Unit", "Description", "Used to start the docker daemon post volume mounts")
	for _, x := range mountUnits {
		u.Set("Unit", "After", x)
		u.Set("Unit", "Requires", x)
	}
	u.Set("Service", "Type", "oneshot")
	u.Set("Service", "RemainAfterExit", "yes")
	u.Set("Service", "ExecStartPre", "/usr/bin/systemctl restart docker.service")
	u.Set("Service", "ExecStart", "/usr/bin/systemctl restart --no-block kops-configuration.service")

	c.AddTask(&nodetasks.Service{
		Name:         b.EnsureSystemdSuffix(b.VolumesServiceName()),
		Definition:   s(u.Render()),
		Enabled:      fi.Bool(true),
		ManageState:  fi.Bool(true),
		Running:      fi.Bool(true),
		SmartRestart: fi.Bool(true),
	})

	return nil
}

// buildDeviceFormatService is responsible for constructing the systemd unit to format the device
func buildDeviceFormatService(c *fi.ModelBuilderContext, volume *kops.VolumeMountSpec) (*nodetasks.Service, error) {
	device := volume.Device
	deviceFmt := strings.TrimPrefix(strings.Replace(device, "/", "-", -1), "-")
	name := fmt.Sprintf("format-%s.service", deviceFmt)

	u := &systemd.Manifest{}
	u.Set("Unit", "Description", fmt.Sprintf("Formats the device: %s", device))
	u.Set("Unit", "After", fmt.Sprintf("%s.device", deviceFmt))
	u.Set("Unit", "Requires", fmt.Sprintf("%s.device", deviceFmt))
	u.Set("Service", "Type", "oneshot")
	u.Set("Service", "RemainAfterExit", "yes")

	// @TODO this was written to work on CoreOS need to check other OS's, add a switch on the distro and potentionally an api override
	command := fmt.Sprintf("/usr/bin/bash -c '/usr/sbin/blkid %s || (/usr/sbin/wipefs -f %s && /usr/sbin/mkfs.%s %s)'",
		device, device, volume.Filesystem, device)

	u.Set("Service", "ExecStart", command)

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
func buildDeviceMountService(c *fi.ModelBuilderContext, volume *kops.VolumeMountSpec, formatName string) (*nodetasks.Service, error) {
	device := volume.Device
	mountpath := volume.Path
	name := fmt.Sprintf("%s.mount", strings.TrimPrefix(strings.Replace(mountpath, "/", "-", -1), "-"))

	// @step: create the mounting unit
	u := &systemd.Manifest{}
	u.Set("Unit", "Description", fmt.Sprintf("Mounting volume: %s from device: %s", mountpath, device))
	u.Set("Unit", "Requires", formatName)
	u.Set("Unit", "After", formatName)
	u.Set("Unit", "Before", "docker.service")
	u.Set("Mount", "What", device)
	u.Set("Mount", "Where", mountpath)
	u.Set("Mount", "Type", volume.Filesystem)
	u.Set("Mount", "Options", "defaults")

	return &nodetasks.Service{
		Name:         name,
		Definition:   s(u.Render()),
		Enabled:      fi.Bool(true),
		ManageState:  fi.Bool(true),
		Running:      fi.Bool(true),
		SmartRestart: fi.Bool(true),
	}, nil
}

func santizeDeviceName(name string) string {
	return strings.TrimPrefix(strings.Replace(name, "/", "-", -1), "-")
}
