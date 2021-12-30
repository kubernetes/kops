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
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// VMScaleSetModelBuilder configures VMScaleSet objects
type VMScaleSetModelBuilder struct {
	*AzureModelContext
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
}

var _ fi.ModelBuilder = &VMScaleSetModelBuilder{}

// Build is responsible for constructing the VM ScaleSet from the kops spec.
func (b *VMScaleSetModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		name := b.AutoscalingGroupName(ig)
		vmss, err := b.buildVMScaleSetTask(c, name, ig)
		if err != nil {
			return err
		}
		c.AddTask(vmss)

		// Create tasks for assigning built-in roles to VM Scale Sets.
		// See https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles
		// for the ID definitions.
		roleDefIDs := map[string]string{
			// Owner
			"owner": "8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
			// Storage Blob Data Contributor
			"blob": "ba92f5b4-2d11-453d-a403-e96b0029c9fe",
		}
		for k, roleDefID := range roleDefIDs {
			c.AddTask(b.buildRoleAssignmentTask(vmss, k, roleDefID))
		}
	}

	return nil
}

func (b *VMScaleSetModelBuilder) buildVMScaleSetTask(
	c *fi.ModelBuilderContext,
	name string,
	ig *kops.InstanceGroup,
) (*azuretasks.VMScaleSet, error) {
	var azNumbers []string
	for _, zone := range ig.Spec.Zones {
		az, err := azure.ZoneToAvailabilityZoneNumber(zone)
		if err != nil {
			return nil, err
		}
		azNumbers = append(azNumbers, az)
	}
	t := &azuretasks.VMScaleSet{
		Name:               fi.String(name),
		Lifecycle:          b.Lifecycle,
		ResourceGroup:      b.LinkToResourceGroup(),
		VirtualNetwork:     b.LinkToVirtualNetwork(),
		SKUName:            fi.String(ig.Spec.MachineType),
		ComputerNamePrefix: fi.String(ig.Name),
		AdminUser:          fi.String(b.Cluster.Spec.CloudProvider.Azure.AdminUser),
		Zones:              azNumbers,
	}

	var err error
	if t.Capacity, err = getCapacity(&ig.Spec); err != nil {
		return nil, err
	}

	sp, err := getStorageProfile(&ig.Spec)
	if err != nil {
		return nil, err
	}
	t.StorageProfile = &azuretasks.VMScaleSetStorageProfile{
		VirtualMachineScaleSetStorageProfile: sp,
	}

	if n := len(b.SSHPublicKeys); n > 0 {
		if n > 1 {
			return nil, fmt.Errorf("expected at most one SSH public key; found %d keys", n)
		}
		t.SSHPublicKey = fi.String(string(b.SSHPublicKeys[0]))
	}

	if t.CustomData, err = b.BootstrapScriptBuilder.ResourceNodeUp(c, ig); err != nil {
		return nil, err
	}

	subnets, err := b.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	if len(subnets) != 1 {
		return nil, fmt.Errorf("expected exactly one subnet for InstanceGroup %q; subnets was %s", ig.Name, ig.Spec.Subnets)
	}
	subnet := subnets[0]
	t.Subnet = b.LinkToAzureSubnet(subnet)

	switch subnet.Type {
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		t.RequirePublicIP = fi.Bool(true)
		if ig.Spec.AssociatePublicIP != nil {
			t.RequirePublicIP = ig.Spec.AssociatePublicIP
		}
	case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
		t.RequirePublicIP = fi.Bool(false)
	default:
		return nil, fmt.Errorf("unexpected subnet type: for InstanceGroup %q; type was %s", ig.Name, subnet.Type)
	}

	if ig.Spec.Role == kops.InstanceGroupRoleMaster && b.Cluster.Spec.API.LoadBalancer != nil {
		t.LoadBalancer = &azuretasks.LoadBalancer{
			Name: to.StringPtr(b.NameForLoadBalancer()),
		}
	}

	t.Tags = b.CloudTagsForInstanceGroup(ig)

	return t, nil
}

func getCapacity(spec *kops.InstanceGroupSpec) (*int64, error) {
	// Follow the convention that all other CSPs have.
	minSize := int32(1)
	maxSize := int32(1)
	if spec.MinSize != nil {
		minSize = fi.Int32Value(spec.MinSize)
	} else if spec.Role == kops.InstanceGroupRoleNode {
		minSize = 2
	}
	if spec.MaxSize != nil {
		maxSize = *spec.MaxSize
	} else if spec.Role == kops.InstanceGroupRoleNode {
		maxSize = 2
	}
	if minSize != maxSize {
		return nil, fmt.Errorf("instance group must have the same min and max size in Azure, but got %d and %d", minSize, maxSize)
	}
	return fi.Int64(int64(minSize)), nil
}

func getStorageProfile(spec *kops.InstanceGroupSpec) (*compute.VirtualMachineScaleSetStorageProfile, error) {
	var volumeSize int32
	if spec.RootVolumeSize != nil {
		volumeSize = *spec.RootVolumeSize
	} else {
		var err error
		volumeSize, err = defaults.DefaultInstanceGroupVolumeSize(spec.Role)
		if err != nil {
			return nil, err
		}
	}

	var storageAccountType compute.StorageAccountTypes
	if spec.RootVolumeType != nil {
		storageAccountType = compute.StorageAccountTypes(*spec.RootVolumeType)
	} else {
		storageAccountType = compute.StorageAccountTypesPremiumLRS
	}

	imageReference, err := parseImage(spec.Image)
	if err != nil {
		return nil, err
	}

	return &compute.VirtualMachineScaleSetStorageProfile{
		ImageReference: imageReference,
		OsDisk: &compute.VirtualMachineScaleSetOSDisk{
			// TODO(kenji): Support Windows.
			OsType:       compute.OperatingSystemTypes(compute.Linux),
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			DiskSizeGB:   to.Int32Ptr(volumeSize),
			ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
				StorageAccountType: storageAccountType,
			},
			Caching: compute.CachingTypes(compute.HostCachingReadWrite),
		},
	}, nil
}

func parseImage(image string) (*compute.ImageReference, error) {
	if strings.HasPrefix(image, "/subscriptions/") {
		return &compute.ImageReference{
			ID: to.StringPtr(image),
		}, nil
	}

	l := strings.Split(image, ":")
	if len(l) != 4 {
		return nil, fmt.Errorf("malformed format of image urn: %s", image)
	}
	return &compute.ImageReference{
		Publisher: to.StringPtr(l[0]),
		Offer:     to.StringPtr(l[1]),
		Sku:       to.StringPtr(l[2]),
		Version:   to.StringPtr(l[3]),
	}, nil
}

func (b *VMScaleSetModelBuilder) buildRoleAssignmentTask(vmss *azuretasks.VMScaleSet, roleKey, roleDefID string) *azuretasks.RoleAssignment {
	name := fmt.Sprintf("%s-%s", *vmss.Name, roleKey)
	return &azuretasks.RoleAssignment{
		Name:          to.StringPtr(name),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		VMScaleSet:    vmss,
		RoleDefID:     to.StringPtr(roleDefID),
	}
}
