/*
Copyright 2024 The Kubernetes Authors.

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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/batcher"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// describeInstanceBatcher contains the batcher details
type describeInstanceBatcher struct {
	batcher *batcher.Batcher[ec2.DescribeInstancesInput, ec2types.Instance]
}

// newdescribeInstanceBatcher creates a createdescribeInstanceBatcher object
func newdescribeInstanceBatcher(ctx context.Context, ec2api iface.EC2) *describeInstanceBatcher {
	options := batcher.Options[ec2.DescribeInstancesInput, ec2types.Instance]{
		Name:          "describe_instance",
		IdleTimeout:   100 * time.Millisecond,
		MaxTimeout:    1 * time.Second,
		MaxItems:      500,
		RequestHasher: describeInstanceHasher,
		BatchExecutor: execDescribeInstanceBatch(ec2api),
	}
	return &describeInstanceBatcher{batcher: batcher.NewBatcher(ctx, options)}
}

// DescribeInstances adds describe instances input to batcher
func (b *describeInstanceBatcher) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput) ([]*ec2types.Instance, error) {
	if len(input.InstanceIds) != 1 {
		return nil, fmt.Errorf("expected to receive a single instance only, found %d", len(input.InstanceIds))
	}
	result := b.batcher.Add(ctx, input)
	if result.Output == nil {
		return nil, result.Err
	}
	return []*ec2types.Instance{result.Output}, result.Err
}

// DescribeInstanceHasher generates hash for different describe instances inputs
// Same inputs have same hash, so they get executed together
func describeInstanceHasher(ctx context.Context, input *ec2.DescribeInstancesInput) uint64 {
	// We use filters for hashing because we want requests with same filters being made together so that we dont batch mutually exclusive filter requests together
	hash, err := hashstructure.Hash(input.Filters, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		log.FromContext(ctx).Error(err, "failed hashing input filters")
	}
	return hash
}

func execDescribeInstanceBatch(ec2api iface.EC2) batcher.BatchExecutor[ec2.DescribeInstancesInput, ec2types.Instance] {
	return func(ctx context.Context, inputs []*ec2.DescribeInstancesInput) []batcher.Result[ec2types.Instance] {
		results := make([]batcher.Result[ec2types.Instance], len(inputs))

		firstInput := *inputs[0]
		// aggregate instanceIDs into 1 input
		for _, input := range inputs[1:] {
			firstInput.InstanceIds = append(firstInput.InstanceIds, input.InstanceIds...)
		}
		batchedInput := &ec2.DescribeInstancesInput{
			InstanceIds: firstInput.InstanceIds,
		}
		klog.Infof("Batched describe instances %v", batchedInput)
		output, err := ec2api.DescribeInstances(ctx, batchedInput)
		if err != nil {
			klog.Errorf("Error occurred trying to batch describe instance, trying individually, %v", err)
			var wg sync.WaitGroup
			for idx, input := range inputs {
				wg.Add(1)
				go func(input *ec2.DescribeInstancesInput) {
					defer wg.Done()
					out, err := ec2api.DescribeInstances(ctx, input)
					if err != nil || len(out) == 0 {
						results[idx] = batcher.Result[ec2types.Instance]{Output: nil, Err: err}
						return
					}
					results[idx] = batcher.Result[ec2types.Instance]{Output: &out[0], Err: err}
				}(input)
			}
			wg.Wait()
		} else {
			instanceIDToOutputMap := map[string]ec2types.Instance{}
			lo.ForEach(output, func(o ec2types.Instance, _ int) { instanceIDToOutputMap[lo.FromPtr(o.InstanceId)] = o })
			for idx, input := range inputs {
				o, ok := instanceIDToOutputMap[input.InstanceIds[0]]
				if !ok {
					results[idx] = batcher.Result[ec2types.Instance]{Output: nil}
					continue
				}
				results[idx] = batcher.Result[ec2types.Instance]{Output: &o}
			}
		}
		return results
	}
}
