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

package openstackmodel

import (
	"fmt"

	openstackutil "k8s.io/cloud-provider-openstack/pkg/util/openstack"
	"k8s.io/klog"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

type OpenstackModelContext struct {
	*model.KopsModelContext
}

func (c *OpenstackModelContext) UseVIPACL() bool {
	tags := make(map[string]string)
	tags[openstack.TagClusterName] = c.ClusterName()
	osCloud, err := openstack.NewOpenstackCloud(tags, &c.Cluster.Spec)
	if err != nil {
		klog.Errorf("Failed with error %v", err)
		return false
	}
	return openstackutil.IsOctaviaFeatureSupported(osCloud.LoadBalancerClient(), openstackutil.OctaviaFeatureVIPACL)
}

func (c *OpenstackModelContext) GetNetworkName() (string, error) {
	if c.Cluster.Spec.NetworkID == "" {
		return c.ClusterName(), nil
	}

	tags := make(map[string]string)
	tags[openstack.TagClusterName] = c.ClusterName()
	osCloud, err := openstack.NewOpenstackCloud(tags, &c.Cluster.Spec)
	if err != nil {
		return "", fmt.Errorf("error loading cloud: %v", err)
	}

	network, err := osCloud.GetNetwork(c.Cluster.Spec.NetworkID)
	if err != nil {
		return "", err
	}
	return network.Name, nil
}

func (c *OpenstackModelContext) findSubnetClusterSpec(subnet string) (string, error) {
	for _, sp := range c.Cluster.Spec.Subnets {
		if sp.Name == subnet {
			name, err := c.findSubnetNameByID(sp.ProviderID, sp.Name)
			if err != nil {
				return "", err
			}
			return name, nil
		}
	}
	return "", fmt.Errorf("could not find subnet %s from clusterSpec", subnet)
}

func (c *OpenstackModelContext) findSubnetNameByID(subnetID string, subnetName string) (string, error) {
	if subnetID == "" {
		return subnetName + "." + c.ClusterName(), nil
	}

	tags := make(map[string]string)
	tags[openstack.TagClusterName] = c.ClusterName()
	osCloud, err := openstack.NewOpenstackCloud(tags, &c.Cluster.Spec)
	if err != nil {
		return "", fmt.Errorf("error loading cloud: %v", err)
	}

	subnet, err := osCloud.GetSubnet(subnetID)
	if err != nil {
		return "", err
	}
	return subnet.Name, nil
}

func (c *OpenstackModelContext) LinkToNetwork() *openstacktasks.Network {
	netName, err := c.GetNetworkName()
	if err != nil {
		klog.Fatalf("Could not find networkname")
		return nil
	}
	return &openstacktasks.Network{Name: s(netName)}
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
