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

package openstackmodel

import (
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

type OpenstackModelContext struct {
	*model.KopsModelContext
}

func (c *OpenstackModelContext) LinkToNetwork() *openstacktasks.Network {
	return &openstacktasks.Network{Name: s(c.ClusterName())}
}

func (c *OpenstackModelContext) LinkToRouter(name *string) *openstacktasks.Router {
	return &openstacktasks.Router{Name: name}
}

func (c *OpenstackModelContext) LinkToSubnet(name *string) *openstacktasks.Subnet {
	return &openstacktasks.Subnet{Name: name}
}

func (c *OpenstackModelContext) LinkToPort(name *string) *openstacktasks.Port {
	return &openstacktasks.Port{Name: name}
}

func (c *OpenstackModelContext) LinkToSecurityGroup(name string) *openstacktasks.SecurityGroup {
	return &openstacktasks.SecurityGroup{Name: fi.String(name)}
}
