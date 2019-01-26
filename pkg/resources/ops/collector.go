/*
Copyright 2017 The Kubernetes Authors.

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

package ops

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/pkg/resources/ali"
	"k8s.io/kops/pkg/resources/aws"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/pkg/resources/gce"
	"k8s.io/kops/pkg/resources/openstack"
	"k8s.io/kops/upup/pkg/fi"
	cloudali "k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	cloudgce "k8s.io/kops/upup/pkg/fi/cloudup/gce"
	cloudopenstack "k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// ListResources collects the resources from the specified cloud
func ListResources(cloud fi.Cloud, clusterName string, region string) (map[string]*resources.Resource, error) {
	switch cloud.ProviderID() {
	case kops.CloudProviderAWS:
		return aws.ListResourcesAWS(cloud.(awsup.AWSCloud), clusterName)
	case kops.CloudProviderDO:
		return digitalocean.ListResources(cloud.(*digitalocean.Cloud), clusterName)
	case kops.CloudProviderGCE:
		return gce.ListResourcesGCE(cloud.(cloudgce.GCECloud), clusterName, region)
	case kops.CloudProviderOpenstack:
		return openstack.ListResources(cloud.(cloudopenstack.OpenstackCloud), clusterName)
	case kops.CloudProviderVSphere:
		return resources.ListResourcesVSphere(cloud.(*vsphere.VSphereCloud), clusterName)
	case kops.CloudProviderALI:
		return ali.ListResourcesALI(cloud.(cloudali.ALICloud), clusterName, region)
	default:
		return nil, fmt.Errorf("delete on clusters on %q not (yet) supported", cloud.ProviderID())
	}
}
