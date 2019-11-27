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

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

func (c *openstackCloud) GetImage(name string) (*images.Image, error) {
	opts := images.ListOpts{Name: name}
	pager := images.List(c.glanceClient, opts)
	page, err := pager.AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	i, err := images.ExtractImages(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract images: %v", err)
	}

	switch len(i) {
	case 1:
		return &i[0], nil
	case 0:
		return nil, fmt.Errorf("no image found with name %v", name)
	default:
		return nil, fmt.Errorf("multiple images found with name %v", name)
	}
}
