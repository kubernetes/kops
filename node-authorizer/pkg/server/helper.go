/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"k8s.io/kops/node-authorizer/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClientAddress returns the client address
func getClientAddress(address string) (string, error) {
	host, _, err := net.SplitHostPort(address)

	return host, err
}

// isNodeRegistered checks if the node is already registered with kubernetes
func isNodeRegistered(ctx context.Context, client kubernetes.Interface, nodename string) (bool, error) {
	var registered bool

	maxInterval := 1000 * time.Millisecond
	maxTime := 10 * time.Second

	// @lets try multiple times
	err := utils.Retry(ctx, maxInterval, maxTime, func() error {
		// @step: get a lit of nodes
		nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("kubernetes.io/hostname=%s", nodename),
		})
		if err != nil {
			return err
		}

		// @check if we found a registered node
		if len(nodes.Items) > 0 {
			registered = true
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return registered, nil
}
