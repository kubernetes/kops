/*
Copyright 2024 The Kubernetes Authors.

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

package controllerclientset

import (
	"context"
	"fmt"
	"os"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/util/pkg/vfs"
)

type addonsClient struct {
	basePath vfs.Path

	cluster *kops.Cluster
}

var _ simple.AddonsClient = &addonsClient{}

func newAddonsClient(basePath vfs.Path, cluster *kops.Cluster) *addonsClient {
	if cluster == nil || cluster.Name == "" {
		klog.Fatalf("cluster / cluster.Name is required")
	}

	r := &addonsClient{
		basePath: basePath,
		cluster:  cluster,
	}

	return r
}

func (c *addonsClient) Replace(addons kubemanifest.ObjectList) error {
	return fmt.Errorf("server-side addons client does not support Addons::Replace")
}

func (c *addonsClient) List(ctx context.Context) (kubemanifest.ObjectList, error) {
	configPath := c.basePath.Join("default")

	b, err := configPath.ReadFile(ctx)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading addons file %s: %v", configPath, err)
	}

	objects, err := kubemanifest.LoadObjectsFrom(b)
	if err != nil {
		return nil, err
	}

	return objects, nil
}
