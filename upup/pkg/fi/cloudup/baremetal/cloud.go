/*
Copyright 2016 The Kubernetes Authors.

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

package baremetal

import (
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type Cloud struct {
	dns dnsprovider.Interface
}

var _ fi.Cloud = &Cloud{}

func NewCloud(dns dnsprovider.Interface) (*Cloud, error) {
	return &Cloud{dns: dns}, nil
}

func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderBareMetal
}

func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return c.dns, nil
}

func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("baremetal FindVPCInfo not supported")
}
