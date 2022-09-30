/*
Copyright 2022 The Kubernetes Authors.

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

package yandextasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
)

// MaxMetaDataSize is the max size of the metadata
const MaxMetaDataSize = 524288

// +kops:fitask
type Instance struct {
	Name       *string
	Lifecycle  fi.Lifecycle
	InstanceId string

	FolderId              string
	Description           *string
	Labels                map[string]string
	ZoneId                string
	PlatformId            string
	Subnet                *Subnet
	SSHPublicKeys         [][]byte
	UserData              fi.Resource
	ResourcesSpec         *compute.ResourcesSpec
	Metadata              map[string]string
	MetadataOptions       *compute.MetadataOptions
	BootDiskSpec          *compute.AttachedDiskSpec
	SecondaryDiskSpec     []*compute.AttachedDiskSpec
	LocalDiskSpec         []*compute.AttachedDiskSpec
	FilesystemsSpec       []*compute.AttachedFilesystemSpec
	NetworkInterfaceSpecs []*compute.NetworkInterfaceSpec
	Hostname              string
	SchedulingPolicy      *compute.SchedulingPolicy
	ServiceAccountId      string
	NetworkSettings       *compute.NetworkSettings
	PlacementPolicy       *compute.PlacementPolicy
}

//var _ fi.CompareWithID = &Network{}

//func (v *Network) CompareWithID() string {
//	return v.NetworkId //fi.String(strconv.Itoa(fi.IntValue(v.NetworkId)))
//}

func (e *Instance) Find(c *fi.Context) (*Instance, error) {
	sdk := c.Cloud.(yandex.YandexCloud).SDK()
	filter := fmt.Sprintf("name=\"%s\"", *e.Name) // only filter by name supported atm 08.2022
	r, err := sdk.Compute().Instance().List(context.TODO(), &compute.ListInstancesRequest{
		FolderId: e.FolderId,
		Filter:   filter,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	for _, instance := range r.Instances {
		if instance.Name != *e.Name {
			continue
		}
		matches := &Instance{
			InstanceId:       instance.Id,
			PlatformId:       instance.PlatformId,
			Name:             fi.String(instance.Name),
			Lifecycle:        e.Lifecycle,
			Labels:           instance.Labels,
			FolderId:         instance.FolderId,
			Description:      fi.String(instance.Description),
			ServiceAccountId: instance.ServiceAccountId,
			ZoneId:           instance.ZoneId,
		}
		e.InstanceId = instance.Id
		return matches, nil
	}

	return nil, nil
}

func (e *Instance) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Instance) CheckChanges(a, e, changes *Instance) error {
	if a != nil {
		if changes.FolderId != "" {
			return fi.CannotChangeField("FolderId")
		}
		//if changes.NetworkId != "" {
		//	return fi.CannotChangeField("NetworkId")
		//}
	} else {
		if e.FolderId == "" {
			return fi.RequiredField("FolderId")
		}
	}

	return nil
}

func (_ *Instance) RenderYandex(t *yandex.YandexAPITarget, a, e, changes *Instance) error {
	sdk := t.Cloud.SDK()
	// subnet and imageId are require by instance
	// in case network not exists return fi.NewTryAgainLaterError("waiting for the IAM Instance Profile to be propagated")
	sourceImageID := sourceImage(context.TODO(), sdk)
	// TODO(YuraBeznos): this block gets targetgroup id which is required to add instance to it, should be a better way.
	var tgId string
	//var targetGroup loadbalancer.TargetGroup
	if strings.HasPrefix(*e.Name, "master") {
		tgName := "api"
		// get tg by name
		filter := fmt.Sprintf("name=\"%s\"", tgName) // only filter by name supported atm 08.2022

		r, err := sdk.LoadBalancer().TargetGroup().List(context.TODO(), &loadbalancer.ListTargetGroupsRequest{
			FolderId: e.FolderId,
			Filter:   filter,
			PageSize: 100,
		})
		if err != nil {
			return err
		}
		for _, tg := range r.TargetGroups {
			if tg.Name != tgName {
				continue
			}
			tgId = tg.Id
			//targetGroup = *tg
		}
		if tgId == "" {
			return fmt.Errorf("error target group not found: %s", tgName)
		}
	}
	if a == nil {
		//TODO(YuraBeznos): should be a configurable way to define the node resources
		request := &compute.CreateInstanceRequest{
			FolderId:    e.FolderId,
			Name:        *e.Name,
			Description: *e.Description,
			Labels:      e.Labels,
			ZoneId:      e.ZoneId,

			PlatformId: "standard-v1",
			Metadata:   map[string]string{},
			ResourcesSpec: &compute.ResourcesSpec{
				Cores:  2,
				Memory: 2 * 1024 * 1024 * 1024,
			},
			ServiceAccountId: e.ServiceAccountId,
			BootDiskSpec: &compute.AttachedDiskSpec{
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskSpec_{
					DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
						TypeId: "network-hdd",
						Size:   20 * 1024 * 1024 * 1024,
						Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
							ImageId: sourceImageID,
						},
					},
				},
			},
			NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{
				{
					SubnetId: e.Subnet.SubnetId,
					PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{
						OneToOneNatSpec: &compute.OneToOneNatSpec{
							IpVersion: compute.IpVersion_IPV4,
						},
					},
				},
			},
		}
		if e.UserData != nil {

			d, err := fi.ResourceAsString(e.UserData)
			if err != nil {
				return fmt.Errorf("error rendering Instance UserData: %v", err)
			}

			if e.SSHPublicKeys != nil {
				// TODO(YuraBeznos): add all keys
				request.Metadata = map[string]string{"user-data": d, "ssh-keys": fmt.Sprintf("user:%s", string(e.SSHPublicKeys[0]))}
			} else {
				request.Metadata = map[string]string{"user-data": d}
			}

			//if len(request.Metadata) > MaxMetaDataSize {
			// TODO(YuraBeznos): check size of metadata
			//	return fmt.Errorf("Instance UserData was too large (%d bytes)", len(d))
			//}
		}

		op, err := sdk.WrapOperation(sdk.Compute().Instance().Create(context.TODO(), request))

		if err != nil {
			return err
		}
		//meta, err := op.Metadata()
		//if err != nil {
		//	return err
		//}
		//fmt.Printf("Creating network %s\n",
		//	meta.(*vpc.CreateNetworkMetadata).NetworkId)

		err = op.Wait(context.TODO())
		if err != nil {
			return err
		}
		resp, err := op.Response()
		if err != nil {
			return err
		}
		instance := resp.(*compute.Instance)
		klog.Infof("Yandex instance: %q", instance.Id)
		e.InstanceId = instance.Id

		// TODO(YuraBeznos): this block adds instance to targetgroup, should be a better way.
		if strings.HasPrefix(*e.Name, "master") {
			op, err := sdk.WrapOperation(sdk.LoadBalancer().TargetGroup().AddTargets(context.TODO(), &loadbalancer.AddTargetsRequest{
				TargetGroupId: tgId,
				Targets: []*loadbalancer.Target{
					{
						SubnetId: e.Subnet.SubnetId,
						Address:  instance.NetworkInterfaces[0].GetPrimaryV4Address().GetAddress(),
					},
				},
			}))
			if err != nil {
				return err
			}
			err = op.Wait(context.TODO())
			if err != nil {
				return err
			}
			_, err = op.Response()
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		if changes.Name != nil || changes.Description != nil {
			filter := fmt.Sprintf("name=\"%s\"", *e.Name) // only filter by name supported atm 08.2022
			list, err := sdk.Compute().Instance().List(context.TODO(), &compute.ListInstancesRequest{
				FolderId: e.FolderId,
				Filter:   filter,
				PageSize: 100,
			})
			if err != nil {
				return err
			}

			var instanceId string
			for _, instance := range list.Instances {
				if instance.Name != *e.Name {
					continue
				}
				instanceId = instance.Id
				break
			}
			if instanceId == "" {
				return nil
			}
			op, err := sdk.WrapOperation(
				sdk.Compute().Instance().Update(context.TODO(), &compute.UpdateInstanceRequest{
					InstanceId:  instanceId, // TODO(yb): might be a better way to get this id
					Name:        *e.Name,
					Description: *e.Description,
					Labels:      e.Labels,
				}))
			if err != nil {
				return err
			}
			err = op.Wait(context.TODO())
			if err != nil {
				return err
			}
			resp, err := op.Response()
			if err != nil {
				return err
			}
			subnet := resp.(*vpc.Subnet)
			klog.Infof("Yandex instance has been updated: %q", subnet.Id)
		}
		return nil
	}
}

func sourceImage(ctx context.Context, sdk *ycsdk.SDK) string {
	image, err := sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: "standard-images",
		Family:   "ubuntu-2004-lts",
	})
	if err != nil {
		klog.Fatal(err)
	}
	return image.Id
}
