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

package yandexmodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandextasks"
)

// YandexModelContext Yandex Model Context
type YandexModelContext struct {
	*model.KopsModelContext
}

// LinkToNetwork returns the Yandex.Cloud Network object the cluster is located in
func (c *YandexModelContext) LinkToNetwork() (*yandextasks.Network, error) {
	return &yandextasks.Network{Name: fi.String(c.ClusterName())}, nil
}

// LinkToSubnet returns a link to the Yandex.Cloud subnet object
func (c *YandexModelContext) LinkToSubnet(subnet *kops.ClusterSubnetSpec) *yandextasks.Subnet {
	name := subnet.Name
	return &yandextasks.Subnet{Name: fi.String(name)}
}

// LinkToTargetGroup returns a link to the Yandex.Cloud targetGroup object
