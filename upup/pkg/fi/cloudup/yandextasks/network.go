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
type Network struct {
	Name      *string
	Lifecycle fi.Lifecycle

	NetworkId   string
	FolderId    string
	Description *string
	Labels      map[string]string
}

//var _ fi.CompareWithID = &Network{}

//func (v *Network) CompareWithID() string {
//	return v.NetworkId //fi.String(strconv.Itoa(fi.IntValue(v.NetworkId)))
//}

func (v *Network) Find(c *fi.Context) (*Network, error) {
	sdk := c.Cloud.(yandex.YandexCloud).SDK()
	filter := fmt.Sprintf("name=\"%s\"", *v.Name) // only filter by name supported atm 08.2022
	resp, err := sdk.VPC().Network().List(context.TODO(), &vpc.ListNetworksRequest{
		FolderId: v.FolderId,
		Filter:   filter,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	for _, network := range resp.Networks {
		if network.Name != *v.Name {
			continue
		}
		matches := &Network{
			Name:      fi.String(network.Name),
			Lifecycle: v.Lifecycle,
			Labels:    network.Labels,
			//NetworkId:   network.Id,
			FolderId:    network.FolderId,
			Description: fi.String(network.Description),
		}
		return matches, nil
	}

	return nil, nil
}

func (v *Network) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Network) CheckChanges(a, e, changes *Network) error {
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

func (_ *Network) RenderYandex(t *yandex.YandexAPITarget, a, e, changes *Network) error {
	sdk := t.Cloud.SDK()
	// network required by subnet
	if a == nil {
		op, err := sdk.WrapOperation(
			sdk.VPC().Network().Create(context.TODO(), &vpc.CreateNetworkRequest{
				FolderId:    e.FolderId,
				Name:        *e.Name,
				Description: *e.Description,
				Labels:      e.Labels,
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
		network := resp.(*vpc.Network)
		klog.Infof("Yandex network: %q", network.Id)
		e.NetworkId = network.Id

		return nil
	} else {
		if changes.Name != nil || len(changes.Labels) != 0 || changes.Description != nil {

			filter := fmt.Sprintf("name=\"%s\"", *e.Name) // only filter by name supported atm 08.2022
			list, err := sdk.VPC().Network().List(context.TODO(), &vpc.ListNetworksRequest{
				FolderId: e.FolderId,
				Filter:   filter,
				PageSize: 100,
			})
			if err != nil {
				return err
			}

			var networkId string
			for _, network := range list.Networks {
				if network.Name != *e.Name {
					continue
				}
				networkId = network.Id
				break
			}
			if networkId == "" {
				return nil
			}
			op, err := sdk.WrapOperation(
				sdk.VPC().Network().Update(context.TODO(), &vpc.UpdateNetworkRequest{
					NetworkId:   networkId, // TODO(yb): might be a better way to get this id
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
			network := resp.(*vpc.Network)
			klog.Infof("Yandex network has been updated: %q", network.Id)
		}
		return nil
	}
}
