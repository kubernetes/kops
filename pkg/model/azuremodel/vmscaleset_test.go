/*
Copyright 2020 The Kubernetes Authors.

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

package azuremodel

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

func TestVMScaleSetModelBuilder_Build(t *testing.T) {
	b := VMScaleSetModelBuilder{
		AzureModelContext: newTestAzureModelContext(),
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle: fi.LifecycleSync,
			KopsModelContext: &model.KopsModelContext{
				IAMModelContext: iam.IAMModelContext{
					Cluster: &kops.Cluster{
						Spec: kops.ClusterSpec{
							Networking: kops.NetworkingSpec{},
						},
					},
				},
			},
		},
	}
	c := &fi.CloudupModelBuilderContext{
		Tasks: make(map[string]fi.CloudupTask),
	}

	caTask := &fitasks.Keypair{
		Name:    fi.PtrTo(fi.CertificateIDCA),
		Subject: "cn=kubernetes",
		Type:    "ca",
	}
	c.AddTask(caTask)
	etcdCaTask := &fitasks.Keypair{
		Name:    fi.PtrTo("etcd-clients-ca"),
		Subject: "cn=etcd-clients-ca",
		Type:    "ca",
	}
	c.AddTask(etcdCaTask)
	for _, cert := range []string{
		"kubelet",
		"kube-proxy",
	} {
		c.AddTask(&fitasks.Keypair{
			Name:    fi.PtrTo(cert),
			Subject: "cn=" + cert,
			Signer:  caTask,
			Type:    "client",
		})
	}

	err := b.Build(c)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
}

func TestGetCapacity(t *testing.T) {
	testCases := []struct {
		spec     kops.InstanceGroupSpec
		success  bool
		capacity int64
	}{
		{
			spec: kops.InstanceGroupSpec{
				Role:    kops.InstanceGroupRoleControlPlane,
				MinSize: fi.PtrTo(int32(3)),
				MaxSize: fi.PtrTo(int32(3)),
			},
			success:  true,
			capacity: 3,
		},
		{
			spec: kops.InstanceGroupSpec{
				Role: kops.InstanceGroupRoleControlPlane,
			},
			success:  true,
			capacity: 1,
		},
		{
			spec: kops.InstanceGroupSpec{
				Role: kops.InstanceGroupRoleNode,
			},
			success:  true,
			capacity: 2,
		},
		{
			spec: kops.InstanceGroupSpec{
				Role:    kops.InstanceGroupRoleControlPlane,
				MinSize: fi.PtrTo(int32(1)),
				MaxSize: fi.PtrTo(int32(2)),
			},
			success: false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			c, err := getCapacity(&tc.spec)
			if !tc.success {
				if err == nil {
					t.Fatalf("unexpected success")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if *c != tc.capacity {
				t.Fatalf("expected %d, but got %d", *c, tc.capacity)
			}
		})
	}
}

func TestGetStorageProfile(t *testing.T) {
	testCases := []struct {
		spec    kops.InstanceGroupSpec
		success bool
		profile *compute.VirtualMachineScaleSetStorageProfile
	}{
		{
			spec: kops.InstanceGroupSpec{
				Image: "Canonical:UbuntuServer:18.04-LTS:latest",
				RootVolume: &kops.InstanceRootVolumeSpec{
					Type: fi.PtrTo(string(compute.StorageAccountTypesUltraSSDLRS)),
					Size: fi.PtrTo(int32(128)),
				},
			},
			profile: &compute.VirtualMachineScaleSetStorageProfile{
				ImageReference: &compute.ImageReference{
					Publisher: to.Ptr("Canonical"),
					Offer:     to.Ptr("UbuntuServer"),
					SKU:       to.Ptr("18.04-LTS"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &compute.VirtualMachineScaleSetOSDisk{
					OSType:       to.Ptr(compute.OperatingSystemTypesLinux),
					CreateOption: to.Ptr(compute.DiskCreateOptionTypesFromImage),
					DiskSizeGB:   to.Ptr[int32](128),
					ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
						StorageAccountType: to.Ptr(compute.StorageAccountTypesUltraSSDLRS),
					},
					Caching: to.Ptr(compute.CachingTypesReadWrite),
				},
			},
		},
		{
			spec: kops.InstanceGroupSpec{
				Image: "Canonical:UbuntuServer:18.04-LTS:latest",
				Role:  kops.InstanceGroupRoleControlPlane,
			},
			profile: &compute.VirtualMachineScaleSetStorageProfile{
				ImageReference: &compute.ImageReference{
					Publisher: to.Ptr("Canonical"),
					Offer:     to.Ptr("UbuntuServer"),
					SKU:       to.Ptr("18.04-LTS"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &compute.VirtualMachineScaleSetOSDisk{
					OSType:       to.Ptr(compute.OperatingSystemTypesLinux),
					CreateOption: to.Ptr(compute.DiskCreateOptionTypesFromImage),
					DiskSizeGB:   to.Ptr[int32](defaults.DefaultVolumeSizeMaster),
					ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
						StorageAccountType: to.Ptr(compute.StorageAccountTypesStandardSSDLRS),
					},
					Caching: to.Ptr(compute.CachingTypesReadWrite),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			profile, err := getStorageProfile(&tc.spec)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(profile, tc.profile) {
				t.Fatalf("expected %+v, but got %+v", profile, tc.profile)
			}
		})
	}
}

func TestParseImage(t *testing.T) {
	testCases := []struct {
		image    string
		success  bool
		imageRef *compute.ImageReference
	}{
		{
			image:   "Canonical:UbuntuServer:18.04-LTS:latest",
			success: true,
			imageRef: &compute.ImageReference{
				Publisher: to.Ptr("Canonical"),
				Offer:     to.Ptr("UbuntuServer"),
				SKU:       to.Ptr("18.04-LTS"),
				Version:   to.Ptr("latest"),
			},
		},
		{
			image:   "/subscriptions/<subscription id>/resourceGroups/<resource group>/providers/<provider>/images/<image>",
			success: true,
			imageRef: &compute.ImageReference{
				ID: to.Ptr("/subscriptions/<subscription id>/resourceGroups/<resource group>/providers/<provider>/images/<image>"),
			},
		},
		{
			image:   "invalidformat",
			success: false,
		},
		{
			image:   "inv:ali:dfo:rma:t",
			success: false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			imageRef, err := parseImage(tc.image)
			if !tc.success {
				if err == nil {
					t.Fatalf("unexpected success")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(imageRef, tc.imageRef) {
				t.Fatalf("expected %+v, but got %+v", imageRef, tc.imageRef)
			}
		})
	}
}
