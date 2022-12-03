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

package server

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	pb "k8s.io/kops/proto/generated/kops/kopscontroller/v1"
)

// Finds addresses for well-known hosts (apiserver etc) without using DNS
func (s *Server) DiscoverHosts(req *pb.DiscoverHostsRequest, stream pb.KopsControllerService_DiscoverHostsServer) error {
	ctx := stream.Context()

	// TODO: Authentication

	for {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion("v1")
		u.SetKind("ConfigMap")
		configMapID := types.NamespacedName{
			Namespace: "kube-system",
			Name:      "coredns",
		}
		if err := s.client.Get(ctx, configMapID, u); err != nil {
			// TODO: Maybe retry or wait?
			klog.Warningf("error getting coredns configmap: %v", err)
			return fmt.Errorf("error getting coredns configmap: %w", err)
		}

		hosts, _, err := unstructured.NestedString(u.Object, "data", "hosts")
		if err != nil {
			// TODO: Maybe retry or wait?
			klog.Warningf("error getting data.hosts from coredns configmap: %v", err)
			return fmt.Errorf("error getting data.hosts from coredns configmap: %w", err)
		}

		records := make(map[string]*pb.HostRecord)

		for _, line := range strings.Split(hosts, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			tokens := strings.Fields(line)
			if len(tokens) < 2 {
				klog.Warningf("unexpected line in coredns configmap: %q", line)
				continue
			}
			addr := tokens[0]
			for _, host := range tokens[1:] {
				record := records[host]
				if record == nil {
					record = &pb.HostRecord{}
					records[host] = record
				}
				record.Addresses = append(record.Addresses, &pb.Address{Address: addr})
			}
		}

		if len(records) == 0 {
			klog.Warningf("data.hosts from coredns confgimap is empty")
		}

		msg := &pb.DiscoverHostsResponse{
			Complete: true,
		}
		for _, record := range records {
			msg.Records = append(msg.Records, record)
		}

		// TODO: Only if changed (may need to normalize, but can also just check if configmap itself has changed)

		if err := stream.Send(msg); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}
