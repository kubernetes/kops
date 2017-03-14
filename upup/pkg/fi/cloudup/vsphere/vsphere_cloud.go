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
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	k8sroute53 "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	"net/url"
	"os"
)

type VSphereCloud struct {
	Server     string
	Datacenter string
	Cluster    string
	Username   string
	Password   string
	Client     *govmomi.Client
}

const (
	snapshotName string = "LinkCloneSnapshotPoint"
	snapshotDesc string = "Snapshot created by kops"
)

var _ fi.Cloud = &VSphereCloud{}

func (c *VSphereCloud) ProviderID() fi.CloudProviderID {
	return fi.CloudProviderVSphere
}

func NewVSphereCloud(spec *kops.ClusterSpec) (*VSphereCloud, error) {
	server := *spec.CloudConfig.VSphereServer
	datacenter := *spec.CloudConfig.VSphereDatacenter
	cluster := *spec.CloudConfig.VSphereResourcePool
	glog.V(2).Infof("Creating vSphere Cloud with server(%s), datacenter(%s), cluster(%s)", server, datacenter, cluster)

	username := os.Getenv("VSPHERE_USERNAME")
	password := os.Getenv("VSPHERE_PASSWORD")
	if username == "" || password == "" {
		return nil, fmt.Errorf("Failed to detect vSphere username and password. Please set env variables: VSPHERE_USERNAME and VSPHERE_PASSWORD accordingly.")
	}

	u, err := url.Parse(fmt.Sprintf("https://%s/sdk", server))
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("Creating vSphere Cloud URL is %s", u)

	// set username and password in URL
	u.User = url.UserPassword(username, password)

	c, err := govmomi.NewClient(context.TODO(), u, true)
	if err != nil {
		return nil, err
	}
	// Add retry functionality
	c.RoundTripper = vim25.Retry(c.RoundTripper, vim25.TemporaryNetworkError(5))
	vsphereCloud := &VSphereCloud{Server: server, Datacenter: datacenter, Cluster: cluster, Username: username, Password: password, Client: c}
	glog.V(2).Infof("Created vSphere Cloud successfully: %+v", vsphereCloud)
	return vsphereCloud, nil
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
	glog.Warning("FindVPCInfo not (yet) implemented on VSphere")
	return nil, nil
}

func (c *VSphereCloud) CreateLinkClonedVm(vmName, vmImage *string) (string, error) {
	f := find.NewFinder(c.Client.Client, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc, err := f.Datacenter(ctx, c.Datacenter)
	if err != nil {
		return "", err
	}
	f.SetDatacenter(dc)

	templateVm, err := f.VirtualMachine(ctx, *vmImage)
	if err != nil {
		return "", err
	}

	glog.V(2).Infof("Template VM ref is %+v", templateVm)
	datacenterFolders, err := dc.Folders(ctx)
	if err != nil {
		return "", err
	}

	// Create snapshot of the template VM if not already snapshotted.
	snapshot, err := createSnapshot(ctx, templateVm, snapshotName, snapshotDesc)
	if err != nil {
		return "", err
	}

	clsComputeRes, err := f.ClusterComputeResource(ctx, c.Cluster)
	glog.V(4).Infof("Cluster compute resource is %+v", clsComputeRes)
	if err != nil {
		return "", err
	}

	resPool, err := clsComputeRes.ResourcePool(ctx)
	glog.V(4).Infof("Cluster resource pool is %+v", resPool)
	if err != nil {
		return "", err
	}

	if resPool == nil {
		return "", errors.New(fmt.Sprintf("No resource pool found for cluster %s", c.Cluster))
	}

	resPoolRef := resPool.Reference()
	snapshotRef := snapshot.Reference()

	cloneSpec := &types.VirtualMachineCloneSpec{
		Config: &types.VirtualMachineConfigSpec{},
		Location: types.VirtualMachineRelocateSpec{
			Pool:         &resPoolRef,
			DiskMoveType: "createNewChildDiskBacking",
		},
		Snapshot: &snapshotRef,
	}

	// Create a link cloned VM from the template VM's snapshot
	clonedVmTask, err := templateVm.Clone(ctx, datacenterFolders.VmFolder, *vmName, *cloneSpec)
	if err != nil {
		return "", err
	}

	clonedVmTaskInfo, err := clonedVmTask.WaitForResult(ctx, nil)
	if err != nil {
		return "", err
	}

	clonedVm := clonedVmTaskInfo.Result.(object.Reference)
	glog.V(2).Infof("Created VM %s successfully", clonedVm)

	return clonedVm.Reference().Value, nil
}
