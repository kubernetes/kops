/*
Copyright 2025 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	cloudazure "k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type dumpState struct {
	cloud cloudazure.AzureCloud

	mutex    sync.Mutex
	vmssNICs map[string][]*network.Interface // key: "rg/vmss"
}

func getDumpState(op *resources.DumpOperation) *dumpState {
	if op.CloudState == nil {
		op.CloudState = &dumpState{cloud: op.Cloud.(cloudazure.AzureCloud)}
	}
	return op.CloudState.(*dumpState)
}

func (s *dumpState) getVMSSNICs(ctx context.Context, resourceGroupName, vmssName string) ([]*network.Interface, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.vmssNICs == nil {
		s.vmssNICs = make(map[string][]*network.Interface)
	}

	key := resourceGroupName + "/" + vmssName
	if cached, ok := s.vmssNICs[key]; ok {
		return cached, nil
	}

	nis, err := s.cloud.NetworkInterface().ListScaleSetsNetworkInterfaces(ctx, resourceGroupName, vmssName)
	if err != nil {
		return nil, err
	}
	s.vmssNICs[key] = nis
	return nis, nil
}

func DumpVMScaleSet(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	return nil
}

func DumpVMScaleSetVM(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	vm, ok := r.Obj.(*compute.VirtualMachineScaleSetVM)
	if !ok {
		return fmt.Errorf("expected VirtualMachineScaleSetVM, got %T", r.Obj)
	}

	rid, err := arm.ParseResourceID(r.ID)
	if err != nil {
		return err
	}
	vmssName := ""
	if rid.Parent != nil {
		vmssName = rid.Parent.Name
	}

	instance := &resources.Instance{Name: r.Name}
	if instance.Name == "" {
		if vm.Properties != nil && vm.Properties.OSProfile != nil && vm.Properties.OSProfile.ComputerName != nil {
			instance.Name = *vm.Properties.OSProfile.ComputerName
		} else {
			instance.Name = rid.Name
		}
	}

	if vm.Properties != nil && vm.Properties.OSProfile != nil && vm.Properties.OSProfile.AdminUsername != nil {
		instance.SSHUser = *vm.Properties.OSProfile.AdminUsername
	}

	// Resolve NIC private IPs by listing VMSS NICs once and matching by virtualMachine.id.
	if vmssName != "" {
		nis, err := getDumpState(op).getVMSSNICs(op.Context, rid.ResourceGroupName, vmssName)
		if err != nil {
			return err
		}
		for _, ni := range nis {
			if ni == nil || ni.Properties == nil || ni.Properties.VirtualMachine == nil || ni.Properties.VirtualMachine.ID == nil {
				continue
			}
			if *ni.Properties.VirtualMachine.ID != r.ID {
				continue
			}
			for _, ip := range ni.Properties.IPConfigurations {
				if ip == nil || ip.Properties == nil {
					continue
				}
				if ip.Properties.PrivateIPAddress != nil {
					instance.PrivateAddresses = append(instance.PrivateAddresses, *ip.Properties.PrivateIPAddress)
				}
			}
		}
	}

	instance.Roles = append(instance.Roles, rolesFromTags(vm.Tags)...)
	op.Dump.Instances = append(op.Dump.Instances, instance)
	return nil
}

func rolesFromTags(tags map[string]*string) []string {
	var roles []string
	if tags == nil {
		return roles
	}

	isControlPlane := false
	for k, v := range tags {
		if v == nil || *v != "1" {
			continue
		}
		if !strings.HasPrefix(k, cloudazure.TagNameRolePrefix) {
			continue
		}
		role := strings.TrimPrefix(k, cloudazure.TagNameRolePrefix)
		switch role {
		case cloudazure.TagRoleControlPlane, cloudazure.TagRoleMaster:
			isControlPlane = true
		default:
			roles = append(roles, role)
		}
	}
	if isControlPlane {
		roles = append(roles, "control-plane")
	}
	return roles
}

func DumpLoadBalancer(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	if lb, ok := r.Obj.(*network.LoadBalancer); ok {
		op.Dump.LoadBalancers = append(op.Dump.LoadBalancers, &resources.LoadBalancer{
			Name: fi.ValueOf(lb.Name),
		})

	}
	return nil
}
