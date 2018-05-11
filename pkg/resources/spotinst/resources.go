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

package spotinst

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/resources/aws"
	"k8s.io/kops/pkg/resources/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	gceup "k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

type Resources struct {
	Cloud       spotinst.Cloud
	ClusterName string
}

// ListResources fetches all spotinst resources into tracker.Resources
func (r *Resources) ListResources() (map[string]*resources.Resource, error) {
	var err error
	var allResources map[string]*resources.Resource

	glog.V(2).Info("Listing external resources")
	cloud := r.Cloud.Cloud()

	switch id := cloud.ProviderID(); id {
	case kops.CloudProviderAWS:
		allResources, err = listResourcesAWS(cloud, r.ClusterName)
	case kops.CloudProviderGCE:
		allResources, err = listResourcesGCE(cloud, r.ClusterName)
	default:
		return nil, fmt.Errorf("spotinst: unknown cloud provider: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to list resources: %v", err)
	}

	glog.V(2).Info("Listing internal resources")
	spotinstResources, err := r.listResourcesSpotinst()
	if err != nil {
		return nil, err
	}

	glog.V(2).Info("Merging resource lists")
	for k, resource := range spotinstResources {
		allResources[k] = resource
	}

	glog.V(2).Info("Removing deleted resources")
	for k, resource := range allResources {
		if resource.Done {
			delete(allResources, k)
		}
	}

	return allResources, nil
}

// DeleteResources deletes all resources passed in the form in tracker.Resources
func (r *Resources) DeleteResources(resources map[string]*resources.Resource) error {
	for _, resource := range resources {
		if err := deleter(r.Cloud, resource); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resources) listResourcesSpotinst() (map[string]*resources.Resource, error) {
	spotinstResources, err := r.Cloud.ListResources(r.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to list resources: %v", err)
	}

	trackers := make(map[string]*resources.Resource)
	for k, resource := range spotinstResources {
		trackers[k] = &resources.Resource{
			ID:      resource.ID,
			Name:    resource.Name,
			Type:    resource.Type,
			Obj:     resource,
			Deleter: deleter,
			Dumper:  dumper,
		}
	}

	return trackers, nil
}

func listResourcesAWS(cloud fi.Cloud, clusterName string) (map[string]*resources.Resource, error) {
	awsCloud := cloud.(awsup.AWSCloud)

	awsResources, err := aws.ListResourcesAWS(awsCloud, clusterName)
	if err != nil {
		return nil, err
	}

	for _, resource := range awsResources {
		if deleter := resource.Deleter; deleter != nil {
			resource.Deleter = func(cloud fi.Cloud, resource *resources.Resource) error {
				return deleter(awsCloud, resource)
			}
		}
		if deleter := resource.GroupDeleter; deleter != nil {
			resource.GroupDeleter = func(cloud fi.Cloud, resources []*resources.Resource) error {
				return deleter(awsCloud, resources)
			}
		}
	}

	return awsResources, nil
}

func listResourcesGCE(cloud fi.Cloud, clusterName string) (map[string]*resources.Resource, error) {
	gceCloud := cloud.(gceup.GCECloud)

	gceResources, err := gce.ListResourcesGCE(gceCloud, clusterName, "")
	if err != nil {
		return nil, err
	}

	for _, resource := range gceResources {
		if deleter := resource.Deleter; deleter != nil {
			resource.Deleter = func(cloud fi.Cloud, resource *resources.Resource) error {
				return deleter(gceCloud, resource)
			}
		}
		if deleter := resource.GroupDeleter; deleter != nil {
			resource.GroupDeleter = func(cloud fi.Cloud, resources []*resources.Resource) error {
				return deleter(gceCloud, resources)
			}
		}
	}

	return gceResources, nil
}

func deleter(cloud fi.Cloud, resource *resources.Resource) error {
	glog.V(2).Infof("Deleting resource: %s (%s)", resource.ID, resource.Name)
	return cloud.(spotinst.Cloud).DeleteResource(resource.Obj)
}

func dumper(op *resources.DumpOperation, resource *resources.Resource) error {
	glog.V(2).Infof("Dumping resource: %s (%s)", resource.ID, resource.Name)

	data := make(map[string]interface{})
	data["id"] = resource.ID
	data["type"] = resource.Type
	data["raw"] = resource.Obj

	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}
