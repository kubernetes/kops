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

package designate

import (
	"github.com/gophercloud/gophercloud"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

var _ dnsprovider.Interface = Interface{}

type Interface struct {
	sc *gophercloud.ServiceClient
}

// New builds an Interface, with a specified Designate implementation.
// This is useful for testing purposes, but also if we want an instance with custom OpenStack options.
func New(sc *gophercloud.ServiceClient) *Interface {
	return &Interface{sc}
}

func (i Interface) Zones() (zones dnsprovider.Zones, supported bool) {
	return Zones{&i}, true
}
