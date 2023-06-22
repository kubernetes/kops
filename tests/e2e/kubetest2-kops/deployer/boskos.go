/*
Copyright 2023 The Kubernetes Authors.

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

package deployer

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/klog/v2"
	"sigs.k8s.io/boskos/client"
	"sigs.k8s.io/boskos/common"
	"sigs.k8s.io/kubetest2/pkg/boskos"
)

type boskosHelper struct {
	// boskos holds the client we use for boskos communication.
	boskos *client.Client

	// this channel serves as a signal channel for the boskos heartbeat goroutine
	// so that it can be explicitly closed
	boskosHeartbeatClose chan struct{}

	// resources tracks acquired resources so they can be freed in Cleanup.
	resources []*common.Resource
}

func (h *boskosHelper) Acquire(ctx context.Context, resourceType string) (*common.Resource, error) {
	if h.boskos == nil {
		h.boskosHeartbeatClose = make(chan struct{})

		boskosURL := os.Getenv("BOSKOS_HOST")
		if boskosURL == "" {
			boskosURL = "http://boskos.test-pods.svc.cluster.local."
		}
		boskosClient, err := boskos.NewClient(boskosURL)
		if err != nil {
			return nil, fmt.Errorf("failed to make boskos client for %q: %w", boskosURL, err)
		}
		h.boskos = boskosClient
	}

	resource, err := boskos.Acquire(
		h.boskos,
		resourceType,
		5*time.Minute,
		5*time.Minute,
		h.boskosHeartbeatClose,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get %q resource from boskos: %w", resourceType, err)
	}
	h.resources = append(h.resources, resource)

	return resource, nil
}

// Cleanup releases any resources acquired from boskos
func (h *boskosHelper) Cleanup(ctx context.Context) error {
	if h.boskos != nil {
		var resourceNames []string
		for _, resource := range h.resources {
			klog.V(2).Info("releasing boskos resource %v %q", resource.Type, resource.Name)
			resourceNames = append(resourceNames, resource.Name)
		}
		err := boskos.Release(
			h.boskos,
			resourceNames,
			h.boskosHeartbeatClose,
		)
		if err != nil {
			return fmt.Errorf("failed to release boskos resources %v: %w", resourceNames, err)
		}
	}

	return nil
}
