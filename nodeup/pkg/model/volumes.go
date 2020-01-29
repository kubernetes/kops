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

package model

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"

	"k8s.io/klog"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
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
		klog.V(1).Info("Skipping the volume builder, no volumes defined for this instancegroup")

		return nil
	}

	// @step: iterate the volume mounts and attempt to mount the devices
	for _, x := range b.InstanceGroup.Spec.VolumeMounts {
		// @check the directory exists, else create it
		if err := b.EnsureDirectory(x.Path); err != nil {
			return fmt.Errorf("failed to ensure the directory: %s, error: %s", x.Path, err)
		}

		m := &mount.SafeFormatAndMount{
			Exec:      utilexec.New(),
			Interface: mount.New(""),
		}

		// @check if the device is already mounted
		if found, err := b.IsMounted(m, x.Device, x.Path); err != nil {
			return fmt.Errorf("Failed to check if device: %s is mounted, error: %s", x.Device, err)
		} else if found {
			klog.V(3).Infof("Skipping device: %s, path: %s as already mounted", x.Device, x.Path)
			continue
		}

		klog.Infof("Attempting to format and mount device: %s, path: %s", x.Device, x.Path)

		if err := m.FormatAndMount(x.Device, x.Path, x.Filesystem, x.MountOptions); err != nil {
			klog.Errorf("failed to mount the device: %s on: %s, error: %s", x.Device, x.Path, err)

			return err
		}
	}

	return nil
}
