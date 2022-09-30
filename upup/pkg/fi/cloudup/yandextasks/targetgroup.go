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
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
)

// TargetGroup represents a Yandex TargetGroup
// +kops:fitask
type TargetGroup struct {
	Name      *string
	FolderId  string
	Lifecycle fi.Lifecycle
}

var _ fi.CompareWithID = &TargetGroup{}

func (e *TargetGroup) CompareWithID() *string {
	return e.Name
}

func (e *TargetGroup) Find(c *fi.Context) (*TargetGroup, error) {
	sdk := c.Cloud.(yandex.YandexCloud).SDK()
	filter := fmt.Sprintf("name=\"%s\"", *e.Name) // only filter by name supported atm 08.2022

	r, err := sdk.LoadBalancer().TargetGroup().List(context.TODO(), &loadbalancer.ListTargetGroupsRequest{
		FolderId: c.Cluster.Spec.Project,
		Filter:   filter,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	for _, tg := range r.TargetGroups {
		if tg.Name != *e.Name {
			continue
		}
		actual := &TargetGroup{
			FolderId:  tg.FolderId,
			Name:      fi.String(tg.Name),
			Lifecycle: e.Lifecycle,
		}
		return actual, nil
	}

	return nil, nil
}

func (e *TargetGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *TargetGroup) CheckChanges(a, e, changes *TargetGroup) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (_ *TargetGroup) RenderYandex(t *yandex.YandexAPITarget, a, e, changes *TargetGroup) error {
	sdk := t.Cloud.SDK()
	name := fi.StringValue(e.Name)

	o := &loadbalancer.CreateTargetGroupRequest{
		FolderId: e.FolderId,
		Name:     name,
	}

	if a == nil {
		klog.V(4).Infof("Creating TargetGroup %q", o.Name)

		op, err := sdk.WrapOperation(sdk.LoadBalancer().TargetGroup().Create(context.TODO(), o))
		//if err != nil {
		//	return fmt.Errorf("error creating TargetGroup %q: %v", name, err)
		//}
		if err != nil {
			return err
		}
		//if err := t.Cloud.WaitForOp(op); err != nil {
		//	return fmt.Errorf("error creating TargetGroup: %v", err)
		//}

		err = op.Wait(context.TODO())
		if err != nil {
			return err
		}
		_, err = op.Response()
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("cannot apply changes to TargetGroup: %v", changes)
	}

	return nil
}
