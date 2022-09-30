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

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
)

// +kops:fitask
type LoadBalancer struct {
	FolderId             string
	Name                 *string
	Lifecycle            fi.Lifecycle
	Description          string
	Labels               map[string]string
	RegionId             string
	Type                 loadbalancer.NetworkLoadBalancer_Type
	ListenerSpecs        []*loadbalancer.ListenerSpec
	AttachedTargetGroups []*loadbalancer.AttachedTargetGroup
}

func (e *LoadBalancer) IsForAPIServer() bool {
	return true
}

func (e *LoadBalancer) Find(c *fi.Context) (*LoadBalancer, error) {
	sdk := c.Cloud.(yandex.YandexCloud).SDK()
	filter := fmt.Sprintf("name=\"%s\"", *e.Name) // only filter by name supported atm 08.2022
	r, err := sdk.LoadBalancer().NetworkLoadBalancer().List(context.TODO(), &loadbalancer.ListNetworkLoadBalancersRequest{
		FolderId: e.FolderId,
		Filter:   filter,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	for _, nlb := range r.NetworkLoadBalancers {
		if nlb.Name != *e.Name {
			continue
		}

		matches := &LoadBalancer{
			Name:        fi.String(nlb.Name),
			Lifecycle:   e.Lifecycle,
			Labels:      nlb.Labels,
			FolderId:    nlb.FolderId,
			Description: nlb.Description,
			Type:        nlb.Type,
		}
		return matches, nil
	}

	return nil, nil
}

func (v *LoadBalancer) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *LoadBalancer) CheckChanges(a, e, changes *LoadBalancer) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}
	return nil
}

func (_ *LoadBalancer) RenderYandex(t *yandex.YandexAPITarget, a, e, changes *LoadBalancer) error {
	// TODO(YuraBeznos): loadbalancer/tagergroup logic should be refactored
	sdk := t.Cloud.SDK()
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
	var tgId string
	for _, tg := range r.TargetGroups {
		if tg.Name != tgName {
			continue
		}
		tgId = tg.Id
	}
	if tgId == "" {
		return fmt.Errorf("error target group not found: %s", tgName)
	}
	// create empty lb with that tg
	if a == nil {
		listener := []*loadbalancer.ListenerSpec{
			{
				Name:       tgName,
				Port:       443,
				Protocol:   loadbalancer.Listener_TCP,
				TargetPort: 443,
			},
		}
		attachedTargetGroups := []*loadbalancer.AttachedTargetGroup{
			{
				TargetGroupId: tgId,
				HealthChecks: []*loadbalancer.HealthCheck{
					{
						Name:               tgName,
						UnhealthyThreshold: 3,
						HealthyThreshold:   3,
						Options: &loadbalancer.HealthCheck_TcpOptions_{
							TcpOptions: &loadbalancer.HealthCheck_TcpOptions{
								Port: 443,
							},
						},
					},
				},
			},
		}
		o := &loadbalancer.CreateNetworkLoadBalancerRequest{
			FolderId:             e.FolderId,
			Name:                 *e.Name,
			Description:          e.Description,
			Labels:               e.Labels,
			Type:                 loadbalancer.NetworkLoadBalancer_EXTERNAL,
			ListenerSpecs:        listener,
			AttachedTargetGroups: attachedTargetGroups,
		}
		op, err := sdk.WrapOperation(sdk.LoadBalancer().NetworkLoadBalancer().Create(context.TODO(), o))
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
		lb := resp.(*loadbalancer.NetworkLoadBalancer)
		klog.Infof("Yandex NetworkLoadBalancer: %q", lb.Id)

		return nil

	} else {
		return nil
	}

}
