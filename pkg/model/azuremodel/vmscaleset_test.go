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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

func TestVMScaleSetModelBuilder_Build(t *testing.T) {
	b := VMScaleSetModelBuilder{
		AzureModelContext: newTestAzureModelContext(),
		BootstrapScriptBuilder: &model.BootstrapScriptBuilder{
			Lifecycle: fi.LifecycleSync,
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{},
				},
			},
		},
	}
	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	caTask := &fitasks.Keypair{
		Name:    fi.String(fi.CertificateIDCA),
		Subject: "cn=kubernetes",
		Type:    "ca",
	}
	c.AddTask(caTask)
	etcdCaTask := &fitasks.Keypair{
		Name:    fi.String("etcd-clients-ca"),
		Subject: "cn=etcd-clients-ca",
		Type:    "ca",
	}
	c.AddTask(etcdCaTask)
	for _, cert := range []string{
		"kubelet",
		"kube-proxy",
	} {
		c.AddTask(&fitasks.Keypair{
			Name:    fi.String(cert),
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
				Role:    kops.InstanceGroupRoleMaster,
				MinSize: fi.Int32(3),
				MaxSize: fi.Int32(3),
			},
			success:  true,
			capacity: 3,
		},
		{
			spec: kops.InstanceGroupSpec{
				Role: kops.InstanceGroupRoleMaster,
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
				Role:    kops.InstanceGroupRoleMaster,
				MinSize: fi.Int32(1),
				MaxSize: fi.Int32(2),
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
				Image:          "Canonical:UbuntuServer:18.04-LTS:latest",
				RootVolumeType: fi.String(string(compute.StorageAccountTypesStandardLRS)),
				RootVolumeSize: fi.Int32(128),
			},
			profile: &compute.VirtualMachineScaleSetStorageProfile{
				ImageReference: &compute.ImageReference{
					Publisher: to.StringPtr("Canonical"),
					Offer:     to.StringPtr("UbuntuServer"),
					Sku:       to.StringPtr("18.04-LTS"),
					Version:   to.StringPtr("latest"),
				},
				OsDisk: &compute.VirtualMachineScaleSetOSDisk{
					OsType:       compute.OperatingSystemTypes(compute.Linux),
					CreateOption: compute.DiskCreateOptionTypesFromImage,
					DiskSizeGB:   to.Int32Ptr(128),
					ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
						StorageAccountType: compute.StorageAccountTypesStandardLRS,
					},
					Caching: compute.CachingTypes(compute.HostCachingReadWrite),
				},
			},
		},
		{
			spec: kops.InstanceGroupSpec{
				Image: "Canonical:UbuntuServer:18.04-LTS:latest",
				Role:  kops.InstanceGroupRoleMaster,
			},
			profile: &compute.VirtualMachineScaleSetStorageProfile{
				ImageReference: &compute.ImageReference{
					Publisher: to.StringPtr("Canonical"),
					Offer:     to.StringPtr("UbuntuServer"),
					Sku:       to.StringPtr("18.04-LTS"),
					Version:   to.StringPtr("latest"),
				},
				OsDisk: &compute.VirtualMachineScaleSetOSDisk{
					OsType:       compute.OperatingSystemTypes(compute.Linux),
					CreateOption: compute.DiskCreateOptionTypesFromImage,
					DiskSizeGB:   to.Int32Ptr(defaults.DefaultVolumeSizeMaster),
					ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
						StorageAccountType: compute.StorageAccountTypesPremiumLRS,
					},
					Caching: compute.CachingTypes(compute.HostCachingReadWrite),
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
				Publisher: to.StringPtr("Canonical"),
				Offer:     to.StringPtr("UbuntuServer"),
				Sku:       to.StringPtr("18.04-LTS"),
				Version:   to.StringPtr("latest"),
			},
		},
		{
			image:   "/subscriptions/<subscription id>/resourceGroups/<resource group>/providers/<provider>/images/<image>",
			success: true,
			imageRef: &compute.ImageReference{
				ID: to.StringPtr("/subscriptions/<subscription id>/resourceGroups/<resource group>/providers/<provider>/images/<image>"),
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
