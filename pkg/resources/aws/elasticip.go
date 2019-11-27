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

package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"k8s.io/kops/pkg/resources"
)

func buildElasticIPResource(address *ec2.Address, forceShared bool, clusterName string) *resources.Resource {
	name := aws.StringValue(address.PublicIp)
	if name == "" {
		name = aws.StringValue(address.PrivateIpAddress)
	}
	if name == "" {
		name = aws.StringValue(address.AllocationId)
	}

	r := &resources.Resource{
		Name:    name,
		ID:      aws.StringValue(address.AllocationId),
		Type:    TypeElasticIp,
		Deleter: DeleteElasticIP,
		Shared:  forceShared,
	}

	if HasSharedTag(r.Type+":"+r.Name, address.Tags, clusterName) {
		r.Shared = true
	}

	return r
}
