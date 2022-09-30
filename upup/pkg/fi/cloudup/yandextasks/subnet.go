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

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
)

// +kops:fitask
type Subnet struct {
	Name      *string
	Lifecycle fi.Lifecycle

	SubnetId    string
	Network     *Network
	FolderId    string //TODO(yb): figure out how to fill FolderID
	Description *string
	// Labels       map[string]string
	ZoneId       string
	V4CidrBlocks []string
	RouteTableId string
	DhcpOptions  *vpc.DhcpOptions
}

//var _ fi.CompareWithID = &Network{}

//func (v *Network) CompareWithID() string {
//	return v.NetworkId //fi.String(strconv.Itoa(fi.IntValue(v.NetworkId)))
//}

func (v *Subnet) Find(c *fi.Context) (*Subnet, error) {
	sdk := c.Cloud.(yandex.YandexCloud).SDK()
	filter := fmt.Sprintf("name=\"%s\"", *v.Name) // only filter by name supported atm 08.2022
	resp, err := sdk.VPC().Subnet().List(context.TODO(), &vpc.ListSubnetsRequest{
		FolderId: v.FolderId,
		Filter:   filter,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	for _, subnet := range resp.Subnets {
		if subnet.Name != *v.Name {
			continue
		}

		matches := &Subnet{
			Name:      fi.String(subnet.Name),
			Lifecycle: v.Lifecycle,
			//Labels:      subnet.Labels,
			SubnetId:     subnet.Id,
			Network:      v.Network, //&Network{NetworkId: subnet.NetworkId},
			FolderId:     subnet.FolderId,
			Description:  fi.String(subnet.Description),
			ZoneId:       subnet.ZoneId,
			V4CidrBlocks: subnet.V4CidrBlocks,
		}
		v.SubnetId = subnet.Id
		return matches, nil
	}

	return nil, nil
}

func (v *Subnet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Subnet) CheckChanges(a, e, changes *Subnet) error {
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

func (_ *Subnet) RenderYandex(t *yandex.YandexAPITarget, a, e, changes *Subnet) error {
	sdk := t.Cloud.SDK()
	// network required by subnet
	// in case network not exists return fi.NewTryAgainLaterError("waiting for the IAM Instance Profile to be propagated")
	if a == nil {
		op, err := sdk.WrapOperation(
			sdk.VPC().Subnet().Create(context.TODO(), &vpc.CreateSubnetRequest{
				FolderId:     e.FolderId,
				Name:         *e.Name,
				Description:  *e.Description,
				NetworkId:    e.Network.NetworkId,
				ZoneId:       e.ZoneId,
				V4CidrBlocks: e.V4CidrBlocks,
			}))
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
		subnet := resp.(*vpc.Subnet)
		klog.Infof("Yandex subnet: %q", subnet.Id)
		e.SubnetId = subnet.Id

		return nil
	} else {
		if changes.Name != nil || changes.Description != nil {
			filter := fmt.Sprintf("name=\"%s\"", *e.Name) // only filter by name supported atm 08.2022
			list, err := sdk.VPC().Subnet().List(context.TODO(), &vpc.ListSubnetsRequest{
				FolderId: e.FolderId,
				Filter:   filter,
				PageSize: 100,
			})
			if err != nil {
				return err
			}

			var subnetId string
			for _, subnet := range list.Subnets {
				if subnet.Name != *e.Name {
					continue
				}
				subnetId = subnet.Id
				break
			}
			if subnetId == "" {
				return nil
			}
			op, err := sdk.WrapOperation(
				sdk.VPC().Subnet().Update(context.TODO(), &vpc.UpdateSubnetRequest{
					SubnetId:    subnetId, // TODO(yb): might be a better way to get this id
					Name:        *e.Name,
					Description: *e.Description,
					//Labels:      e.Labels,
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
			klog.Infof("Yandex subnet has been updated: %q", subnet.Id)
		}
		return nil
	}
}
