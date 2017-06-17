/*
Copyright 2016 The Kubernetes Authors.

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

package gce

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/storage/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)

type GCECloud struct {
	Compute *compute.Service
	Storage *storage.Service

	Region  string
	Project string

	labels map[string]string
}

var _ fi.Cloud = &GCECloud{}

func (c *GCECloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderGCE
}

func NewGCECloud(region string, project string, labels map[string]string) (*GCECloud, error) {
	c := &GCECloud{Region: region, Project: project}

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("error building google API client: %v", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}
	c.Compute = computeService

	storageService, err := storage.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building storage API client: %v", err)
	}
	c.Storage = storageService

	c.labels = labels

	return c, nil
}

func (c *GCECloud) DNS() (dnsprovider.Interface, error) {
	provider, err := clouddns.CreateInterface(c.Project, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	}
	return provider, nil
}

func (c *GCECloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	glog.Warningf("FindVPCInfo not (yet) implemented on GCE")
	return nil, nil
}

func (c *GCECloud) Labels() map[string]string {
	// Defensive copy
	tags := make(map[string]string)
	for k, v := range c.labels {
		tags[k] = v
	}
	return tags
}

func (c *GCECloud) WaitForOp(op *compute.Operation) error {
	return WaitForOp(c.Compute, op)
}

func (c *GCECloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus

	// Note that this must match GCEModelContext::NameForForwardingRule
	name := SafeObjectName("api", cluster.ObjectMeta.Name)

	glog.V(2).Infof("Querying GCE to find ForwardingRules for API (%q)", name)
	forwardingRule, err := c.Compute.ForwardingRules.Get(c.Project, c.Region, name).Do()
	if err != nil {
		if !IsNotFound(err) {
			forwardingRule = nil
		} else {
			return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
		}
	}

	if forwardingRule != nil {
		if forwardingRule.IPAddress == "" {
			return nil, fmt.Errorf("Found forward rule %q, but it did not have an IPAddress", name)
		}

		ingresses = append(ingresses, kops.ApiIngressStatus{
			IP: forwardingRule.IPAddress,
		})
	}

	return ingresses, nil
}
