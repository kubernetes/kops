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
	"context"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/resources/aws"
	"k8s.io/kops/upup/pkg/fi"
)

type listFn func(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error)

// ListResources fetches all spotinst resources into tracker.Resources
func ListResources(cloud fi.Cloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listResourcesAWS,
		listResourcesSpotinst,
	}
	for _, fn := range listFunctions {
		rt, err := fn(cloud, clusterName)
		if err != nil {
			return nil, err
		}
		for _, t := range rt {
			resourceTrackers[t.Type+":"+t.ID] = t
		}
	}

	glog.V(2).Info("Removing deleted resources")
	for k, resource := range resourceTrackers {
		if resource.Done {
			delete(resourceTrackers, k)
		}
	}

	return resourceTrackers, nil
}

func listResourcesAWS(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	resources, err := aws.ListResourcesAWS(cloud.(AWSCloud), clusterName)
	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		resourceTrackers = append(resourceTrackers, resource)
	}

	return resourceTrackers, nil
}

func listResourcesSpotinst(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	svc := cloud.(AWSCloud).Elastigroup()
	out, err := svc.List(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	for _, group := range out.Groups {
		groupID := fi.StringValue(group.ID)
		groupName := fi.StringValue(group.Name)

		if strings.HasSuffix(groupName, clusterName) {
			resource := &resources.Resource{
				ID:      groupID,
				Name:    groupName,
				Obj:     group,
				Deleter: deleteElastigroup,
				Dumper:  dumpElastigroup,
			}
			resourceTrackers = append(resourceTrackers, resource)
		}
	}

	return resourceTrackers, nil
}

func deleteElastigroup(cloud fi.Cloud, resource *resources.Resource) error {
	glog.V(2).Infof("Deleting Elastigroup: %s (%s)", resource.ID, resource.Name)
	return cloud.DeleteGroup(&cloudinstances.CloudInstanceGroup{Raw: resource.Obj})
}

func dumpElastigroup(op *resources.DumpOperation, resource *resources.Resource) error {
	glog.V(2).Infof("Dumping Elastigroup: %s (%s)", resource.ID, resource.Name)

	data := make(map[string]interface{})
	data["id"] = resource.ID
	data["type"] = resource.Type
	data["raw"] = resource.Obj

	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}
