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

// vsphere_cloud is the entry point to vSphere. All operations that need access to vSphere should be housed here.

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	k8scoredns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/coredns"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

// VSphereCloud represents a vSphere cloud instance.
type VSphereCloud struct {
	Server        string
	Datacenter    string
	Cluster       string
	Username      string
	Password      string
	Client        *govmomi.Client
	CoreDNSServer string
	DNSZone       string
}

const (
	snapshotName  string = "LinkCloneSnapshotPoint"
	snapshotDesc  string = "Snapshot created by kops"
	cloudInitFile string = "cloud-init.iso"
)

var _ fi.Cloud = &VSphereCloud{}

// ProviderID returns ID for vSphere type cloud provider.
func (c *VSphereCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderVSphere
}

// Region returns the region bound to the VsphereCloud.
func (c *VSphereCloud) Region() string {
	// TODO: map region with vCenter cluster, or datacenter, or datastore?
	region := c.Cluster
	return region
}

// NewVSphereCloud returns VSphereCloud instance for given ClusterSpec.
func NewVSphereCloud(spec *kops.ClusterSpec) (*VSphereCloud, error) {
	server := *spec.CloudConfig.VSphereServer
	datacenter := *spec.CloudConfig.VSphereDatacenter
	cluster := *spec.CloudConfig.VSphereResourcePool
	klog.V(2).Infof("Creating vSphere Cloud with server(%s), datacenter(%s), cluster(%s)", server, datacenter, cluster)

	dns_server := *spec.CloudConfig.VSphereCoreDNSServer
	dns_zone := spec.DNSZone
	username := os.Getenv("VSPHERE_USERNAME")
	password := os.Getenv("VSPHERE_PASSWORD")
	if username == "" || password == "" {
		return nil, fmt.Errorf("Failed to detect vSphere username and password. Please set env variables: VSPHERE_USERNAME and VSPHERE_PASSWORD accordingly.")
	}

	u, err := url.Parse(fmt.Sprintf("https://%s/sdk", server))
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("Creating vSphere Cloud URL is %s", u)

	// set username and password in URL
	u.User = url.UserPassword(username, password)

	c, err := govmomi.NewClient(context.TODO(), u, true)
	if err != nil {
		return nil, err
	}
	// Add retry functionality
	c.RoundTripper = vim25.Retry(c.RoundTripper, vim25.TemporaryNetworkError(5))
	vsphereCloud := &VSphereCloud{Server: server, Datacenter: datacenter, Cluster: cluster, Username: username, Password: password, Client: c, CoreDNSServer: dns_server, DNSZone: dns_zone}
	spec.CloudConfig.VSphereUsername = fi.String(username)
	spec.CloudConfig.VSpherePassword = fi.String(password)
	klog.V(2).Infof("Created vSphere Cloud successfully: %+v", vsphereCloud)
	return vsphereCloud, nil
}

// GetCloudGroups is not implemented yet, that needs to return the instances and groups that back a kops cluster.
func (c *VSphereCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Infof("vSphere cloud provider GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("vSphere cloud provider does not support getting cloud groups at this time")
}

// DeleteGroup is not implemented yet, is a func that needs to delete a vSphere instance group.
func (c *VSphereCloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	klog.V(8).Infof("vSphere cloud provider DeleteGroup not implemented yet")
	return fmt.Errorf("vSphere cloud provider does not support deleting cloud groups at this time.")
}

// DeleteInstance is not implemented yet, is func needs to delete a vSphereCloud instance.
func (c *VSphereCloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Infof("vSphere cloud provider DeleteInstance not implemented yet")
	return fmt.Errorf("vSphere cloud provider does not support deleting cloud instances at this time.")
}

// DNS returns dnsprovider interface for this vSphere cloud.
func (c *VSphereCloud) DNS() (dnsprovider.Interface, error) {
	var provider dnsprovider.Interface
	var err error
	var lines []string
	lines = append(lines, "etcd-endpoints = "+c.CoreDNSServer)
	lines = append(lines, "zones = "+c.DNSZone)
	config := "[global]\n" + strings.Join(lines, "\n") + "\n"
	file := bytes.NewReader([]byte(config))
	provider, err = dnsprovider.GetDnsProvider(k8scoredns.ProviderName, file)
	if err != nil {
		return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	}

	return provider, nil

}

// FindVPCInfo doesn't perform any operation for now. No VPC is present for vSphere.
func (c *VSphereCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	klog.Warning("FindVPCInfo not (yet) implemented on VSphere")
	return nil, nil
}

// CreateLinkClonedVm creates linked clone of given VM image. This method will perform all necessary steps, like creating snapshot if it's not already present.
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

	klog.V(2).Infof("Template VM ref is %+v", templateVm)
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
	klog.V(4).Infof("Cluster compute resource is %+v", clsComputeRes)
	if err != nil {
		return "", err
	}

	resPool, err := clsComputeRes.ResourcePool(ctx)
	klog.V(4).Infof("Cluster resource pool is %+v", resPool)
	if err != nil {
		return "", err
	}

	if resPool == nil {
		return "", errors.New(fmt.Sprintf("No resource pool found for cluster %s", c.Cluster))
	}

	resPoolRef := resPool.Reference()
	snapshotRef := snapshot.Reference()

	cloneSpec := &types.VirtualMachineCloneSpec{
		Config: &types.VirtualMachineConfigSpec{
			Flags: &types.VirtualMachineFlagInfo{
				DiskUuidEnabled: fi.Bool(true),
			},
		},
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
	klog.V(2).Infof("Created VM %s successfully", clonedVm)

	return clonedVm.Reference().Value, nil
}

// PowerOn powers on given VM.
func (c *VSphereCloud) PowerOn(vm string) error {
	f := find.NewFinder(c.Client.Client, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc, err := f.Datacenter(ctx, c.Datacenter)
	if err != nil {
		return err
	}
	f.SetDatacenter(dc)

	vmRef, err := f.VirtualMachine(ctx, vm)
	if err != nil {
		return err
	}
	task, err := vmRef.PowerOn(ctx)
	if err != nil {
		return err
	}
	task.Wait(ctx)
	return nil
}

// UploadAndAttachISO uploads the ISO to datastore and attaches it to the given VM.
func (c *VSphereCloud) UploadAndAttachISO(vm *string, isoFile string) error {
	f := find.NewFinder(c.Client.Client, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc, err := f.Datacenter(ctx, c.Datacenter)
	if err != nil {
		return err
	}
	f.SetDatacenter(dc)

	vmRef, err := f.VirtualMachine(ctx, *vm)
	if err != nil {
		return err
	}

	var vmResult mo.VirtualMachine

	pc := property.DefaultCollector(c.Client.Client)
	err = pc.RetrieveOne(ctx, vmRef.Reference(), []string{"datastore"}, &vmResult)
	if err != nil {
		klog.Fatalf("Unable to retrieve VM summary for VM %s", *vm)
	}
	klog.V(4).Infof("vm property collector result :%+v\n", vmResult)

	// We expect the VM to be on only 1 datastore
	dsRef := vmResult.Datastore[0].Reference()
	var dsResult mo.Datastore
	err = pc.RetrieveOne(ctx, dsRef, []string{"summary"}, &dsResult)
	if err != nil {
		klog.Fatalf("Unable to retrieve datastore summary for datastore  %s", dsRef)
	}
	klog.V(4).Infof("datastore property collector result :%+v\n", dsResult)
	dsObj, err := f.Datastore(ctx, dsResult.Summary.Name)
	if err != nil {
		return err
	}
	p := soap.DefaultUpload
	dstIsoFile := getCloudInitFileName(*vm)
	klog.V(2).Infof("Uploading ISO file %s to datastore %+v, destination iso is %s\n", isoFile, dsObj, dstIsoFile)
	err = dsObj.UploadFile(ctx, isoFile, dstIsoFile, &p)
	if err != nil {
		return err
	}
	klog.V(2).Infof("Uploaded ISO file %s", isoFile)

	// Find the cd-rom device and insert the cloud init iso file into it.
	devices, err := vmRef.Device(ctx)
	if err != nil {
		return err
	}

	// passing empty cd-rom name so that the first one gets returned
	cdrom, err := devices.FindCdrom("")
	cdrom.Connectable.StartConnected = true
	if err != nil {
		return err
	}
	iso := dsObj.Path(dstIsoFile)
	klog.V(2).Infof("Inserting ISO file %s into cd-rom", iso)
	return vmRef.EditDevice(ctx, devices.InsertIso(cdrom, iso))

}

// Returns VM's instance uuid
func (c *VSphereCloud) FindVMUUID(vm *string) (string, error) {
	f := find.NewFinder(c.Client.Client, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc, err := f.Datacenter(ctx, c.Datacenter)
	if err != nil {
		return "", err
	}
	f.SetDatacenter(dc)

	vmRef, err := f.VirtualMachine(ctx, *vm)
	if err != nil {
		return "", err
	}

	var vmResult mo.VirtualMachine

	pc := property.DefaultCollector(c.Client.Client)
	err = pc.RetrieveOne(ctx, vmRef.Reference(), []string{"config.uuid"}, &vmResult)
	if err != nil {
		return "", err
	}
	klog.V(4).Infof("vm property collector result :%+v\n", vmResult)
	klog.V(3).Infof("retrieved vm uuid as %q for vm %q", vmResult.Config.Uuid, *vm)
	return vmResult.Config.Uuid, nil
}

// GetVirtualMachines returns the VMs where the VM name matches the strings in the argument
func (c *VSphereCloud) GetVirtualMachines(args []string) ([]*object.VirtualMachine, error) {
	var out []*object.VirtualMachine

	// List virtual machines
	if len(args) == 0 {
		return nil, errors.New("no argument")
	}

	f := find.NewFinder(c.Client.Client, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc, err := f.Datacenter(ctx, c.Datacenter)
	if err != nil {
		return nil, err
	}
	f.SetDatacenter(dc)

	var nfe error

	// List virtual machines for every argument
	for _, arg := range args {
		vms, err := f.VirtualMachineList(context.TODO(), arg)
		if err != nil {
			if _, ok := err.(*find.NotFoundError); ok {
				// Let caller decide how to handle NotFoundError
				nfe = err
				continue
			}
			return nil, err
		}

		out = append(out, vms...)
	}

	return out, nfe
}

func (c *VSphereCloud) DeleteCloudInitISO(vm *string) error {
	f := find.NewFinder(c.Client.Client, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc, err := f.Datacenter(ctx, c.Datacenter)
	if err != nil {
		return err
	}
	f.SetDatacenter(dc)

	vmRef, err := f.VirtualMachine(ctx, *vm)
	if err != nil {
		return err
	}

	var vmResult mo.VirtualMachine

	pc := property.DefaultCollector(c.Client.Client)
	err = pc.RetrieveOne(ctx, vmRef.Reference(), []string{"datastore"}, &vmResult)
	if err != nil {
		klog.Fatalf("Unable to retrieve VM summary for VM %s", *vm)
	}
	klog.V(4).Infof("vm property collector result :%+v\n", vmResult)

	// We expect the VM to be on only 1 datastore
	dsRef := vmResult.Datastore[0].Reference()
	var dsResult mo.Datastore
	err = pc.RetrieveOne(ctx, dsRef, []string{"summary"}, &dsResult)
	if err != nil {
		klog.Fatalf("Unable to retrieve datastore summary for datastore  %s", dsRef)
	}
	klog.V(4).Infof("datastore property collector result :%+v\n", dsResult)
	dsObj, err := f.Datastore(ctx, dsResult.Summary.Name)
	if err != nil {
		return err
	}
	isoFileName := getCloudInitFileName(*vm)
	fileManager := dsObj.NewFileManager(dc, false)
	err = fileManager.DeleteFile(ctx, isoFileName)
	if err != nil {
		if types.IsFileNotFound(err) {
			klog.Warningf("ISO file not found: %q", isoFileName)
			return nil
		}
		return err
	}
	klog.V(2).Infof("Deleted ISO file %q", isoFileName)
	return nil
}

func getCloudInitFileName(vmName string) string {
	return vmName + "/" + cloudInitFile
}
