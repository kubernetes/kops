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

package openstack

import (
	"fmt"

	cinder "github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) ListVolumes(opt cinder.ListOptsBuilder) ([]cinder.Volume, error) {
	var volumes []cinder.Volume

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := cinder.List(c.cinderClient, opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing volumes %v: %v", opt, err)
		}

		vs, err := cinder.ExtractVolumes(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting volumes from pages: %v", err)
		}
		volumes = vs
		return true, nil
	})
	if err != nil {
		return volumes, err
	} else if done {
		return volumes, nil
	} else {
		return volumes, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateVolume(opt cinder.CreateOptsBuilder) (*cinder.Volume, error) {
	var volume *cinder.Volume

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := cinder.Create(c.cinderClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating volume %v: %v", opt, err)
		}
		volume = v
		return true, nil
	})
	if err != nil {
		return volume, err
	} else if done {
		return volume, nil
	} else {
		return volume, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) AttachVolume(serverID string, opts volumeattach.CreateOpts) (attachment *volumeattach.VolumeAttachment, err error) {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		volumeAttachment, err := volumeattach.Create(c.ComputeClient(), serverID, opts).Extract()
		if err != nil {
			return false, fmt.Errorf("error attaching volume %s to server %s: %v", opts.VolumeID, serverID, err)
		}
		attachment = volumeAttachment
		return true, nil
	})
	if !done {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return attachment, err
	}
	return attachment, err
}

func (c *openstackCloud) SetVolumeTags(id string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}
	if id == "" {
		return fmt.Errorf("error setting tags to unknown volume")
	}
	klog.V(4).Infof("setting tags to cinder volume %q: %v", id, tags)

	opt := cinder.UpdateOpts{Metadata: tags}
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		_, err := cinder.Update(c.cinderClient, id, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error setting tags to cinder volume %q: %v", id, err)
		}
		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) DeleteVolume(volumeID string) error {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err := cinder.Delete(c.cinderClient, volumeID, cinder.DeleteOpts{}).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting volume: %v", err)
		}
		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}
