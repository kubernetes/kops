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

package awsdiscovery

import (
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/protokube/pkg/gossip/aws"
)

// New builds a cloud-backed resolver.
func New(ec2 ec2iface.EC2API, clusterName string) (resolver.Resolver, error) {
	tags := make(map[string]string)
	tags["kubernetes.io/cluster/"+clusterName] = "owned"

	return aws.NewSeedProvider(ec2, tags)
}
