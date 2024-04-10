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
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"k8s.io/kops/pkg/resources"
)

func DumpNatGateway(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func buildNatGatewayResource(ngw ec2types.NatGateway, forceShared bool, clusterName string) *resources.Resource {
	id := aws.ToString(ngw.NatGatewayId)

	r := &resources.Resource{
		Name:    id,
		ID:      id,
		Obj:     ngw,
		Type:    TypeNatGateway,
		Dumper:  DumpNatGateway,
		Deleter: DeleteNatGateway,
		Shared:  forceShared,
	}

	if HasSharedTag(r.Type+":"+r.Name, ngw.Tags, clusterName) {
		r.Shared = true
	}

	// The NAT gateway blocks deletion of any associated Elastic IPs
	for _, address := range ngw.NatGatewayAddresses {
		if address.AllocationId != nil {
			r.Blocks = append(r.Blocks, TypeElasticIp+":"+aws.ToString(address.AllocationId))
		}
	}

	return r
}
