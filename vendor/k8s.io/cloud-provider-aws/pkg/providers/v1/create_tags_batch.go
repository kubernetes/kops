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
	"github.com/mitchellh/hashstructure/v2"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/batcher"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// createTagsBatcher contains the batcher details
type createTagsBatcher struct {
	batcher *batcher.Batcher[ec2.CreateTagsInput, ec2.CreateTagsOutput]
}

// newCreateTagsBatcher creates a newCreateTagsBatcher object
func newCreateTagsBatcher(ctx context.Context, ec2api iface.EC2) *createTagsBatcher {
	options := batcher.Options[ec2.CreateTagsInput, ec2.CreateTagsOutput]{
		Name:          "create_tags",
		IdleTimeout:   100 * time.Millisecond,
		MaxTimeout:    1 * time.Second,
		MaxItems:      50,
		RequestHasher: createTagsHasher,
		BatchExecutor: execCreateTagsBatch(ec2api),
	}
	return &createTagsBatcher{batcher: batcher.NewBatcher(ctx, options)}
}

// CreateTags adds create tag input to batcher
func (b *createTagsBatcher) createTags(ctx context.Context, CreateTagsInput *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	if len(CreateTagsInput.Resources) != 1 {
		return nil, fmt.Errorf("expected to receive a single instance only, found %d", len(CreateTagsInput.Resources))
	}
	result := b.batcher.Add(ctx, CreateTagsInput)
	return result.Output, result.Err
}

// CreateTagsHasher generates hash for different create tag inputs
// Same set of tags have same hash, so they get executed together
func createTagsHasher(ctx context.Context, input *ec2.CreateTagsInput) uint64 {
	hash, err := hashstructure.Hash(input.Tags, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		log.FromContext(ctx).Error(err, "failed hashing input tags")
	}
	return hash
}

func execCreateTagsBatch(ec2api iface.EC2) batcher.BatchExecutor[ec2.CreateTagsInput, ec2.CreateTagsOutput] {
	return func(ctx context.Context, inputs []*ec2.CreateTagsInput) []batcher.Result[ec2.CreateTagsOutput] {
		results := make([]batcher.Result[ec2.CreateTagsOutput], len(inputs))
		firstInput := inputs[0]
		// aggregate instanceIDs into 1 input
		for _, input := range inputs[1:] {
			firstInput.Resources = append(firstInput.Resources, input.Resources...)
		}
		batchedInput := &ec2.CreateTagsInput{
			Resources: firstInput.Resources,
			Tags:      firstInput.Tags,
		}
		klog.Infof("Batched create tags %v", batchedInput)
		output, err := ec2api.CreateTags(batchedInput)

		if err != nil {
			klog.Errorf("Error occurred trying to batch tag resources, trying individually, %v", err)
			var wg sync.WaitGroup
			for idx, input := range inputs {
				wg.Add(1)
				go func(input *ec2.CreateTagsInput) {
					defer wg.Done()
					out, err := ec2api.CreateTags(input)
					results[idx] = batcher.Result[ec2.CreateTagsOutput]{Output: out, Err: err}

				}(input)
			}
			wg.Wait()
		} else {
			for idx := range inputs {
				results[idx] = batcher.Result[ec2.CreateTagsOutput]{Output: output}
			}
		}
		return results
	}
}
