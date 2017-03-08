/*
Copyright 2017 The Kubernetes Authors.

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

package vsphere

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	k8sroute53 "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	"os"
)

type VSphereCloud struct {
	Server     string
	Datacenter string
	Cluster    string
	Username   string
	Password   string
}

var _ fi.Cloud = &VSphereCloud{}

func (c *VSphereCloud) ProviderID() fi.CloudProviderID {
	return fi.CloudProviderVSphere
}

func NewVSphereCloud(spec *kops.ClusterSpec) (*VSphereCloud, error) {
	server := spec.CloudConfig.VSphereServer
	datacenter := spec.CloudConfig.VSphereDatacenter
	cluster := spec.CloudConfig.VSphereResourcePool
	username := os.Getenv("VSPHERE_USERNAME")
	password := os.Getenv("VSPHERE_PASSWORD")
	if username == "" || password == "" {
		return nil, fmt.Errorf("Failed to detect vSphere username and password. Please set env variables: VSPHERE_USERNAME and VSPHERE_PASSWORD accordingly.")
	}

	c := &VSphereCloud{Server: server, Datacenter: datacenter, Cluster: cluster, Username: username, Password: password}
	// TODO: create a client of govmomi here?
	return c, nil
}

func (c *VSphereCloud) DNS() (dnsprovider.Interface, error) {
	glog.Warning("DNS() not implemented on VSphere")
	provider, err := dnsprovider.GetDnsProvider(k8sroute53.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	}
	return provider, nil

}

func (c *VSphereCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	glog.Warningf("FindVPCInfo not (yet) implemented on VSphere")
	return nil, nil
}
