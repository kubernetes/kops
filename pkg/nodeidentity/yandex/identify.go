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

package yandex

import (
	"context"

	"cloud.google.com/go/compute/metadata"
	ycsdk "github.com/yandex-cloud/go-sdk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/nodeidentity"
)

//TODO(YuraBeznos): yandex node identify implementation

// nodeIdentifier identifies a node from Yandex Cloud
type nodeIdentifier struct {
	sdk *ycsdk.SDK
}

func New() (nodeidentity.Identifier, error) {
	sdk, err := ycsdk.Build(context.TODO(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}

	return &nodeIdentifier{
		sdk: sdk,
	}, nil
}

func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	instanceId, err := metadata.InstanceID()
	if err != nil {
		return nil, err
	}
	labels := map[string]string{}
	info := &nodeidentity.Info{
		InstanceID: instanceId,
		Labels:     labels,
	}
	return info, nil
}
