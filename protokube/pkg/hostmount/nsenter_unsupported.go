// +build !linux

/*
Copyright 2014 The Kubernetes Authors.

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

package hostmount

import (
	"errors"

	"k8s.io/utils/mount"
	"k8s.io/utils/nsenter"
)

func New(ne *nsenter.Nsenter) *Mounter {
	return &Mounter{}
}

// Mounter implements mount.Interface for unsupported platforms
type Mounter struct {
}

var errUnsupported = errors.New("util/mount on this platform is not supported")

// Mount always returns an error on unsupported platforms
func (mounter *Mounter) Mount(source string, target string, fstype string, options []string) error {
	return errUnsupported
}

// Unmount always returns an error on unsupported platforms
func (mounter *Mounter) Unmount(target string) error {
	return errUnsupported
}

// List always returns an error on unsupported platforms
func (mounter *Mounter) List() ([]mount.MountPoint, error) {
	return []mount.MountPoint{}, errUnsupported
}

// IsLikelyNotMountPoint always returns an error on unsupported platforms
func (mounter *Mounter) IsLikelyNotMountPoint(file string) (bool, error) {
	return true, errUnsupported
}

// GetMountRefs always returns an error on unsupported platforms
func (mounter *Mounter) GetMountRefs(pathname string) ([]string, error) {
	return nil, errUnsupported
}
