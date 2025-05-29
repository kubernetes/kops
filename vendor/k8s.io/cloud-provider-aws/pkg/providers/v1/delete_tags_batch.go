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

// deleteTagsBatcher contains the batcher details
type deleteTagsBatcher struct {
	batcher *batcher.Batcher[ec2.DeleteTagsInput, ec2.DeleteTagsOutput]
}

// newDeleteTagsBatcher creates a newDeleteTagsBatcher object
func newDeleteTagsBatcher(ctx context.Context, ec2api iface.EC2) *deleteTagsBatcher {
	options := batcher.Options[ec2.DeleteTagsInput, ec2.DeleteTagsOutput]{
		Name:          "delete_tags",
		IdleTimeout:   100 * time.Millisecond,
		MaxTimeout:    1 * time.Second,
		MaxItems:      50,
		RequestHasher: deleteTagsHasher,
		BatchExecutor: execDeleteTagsBatch(ec2api),
	}
	return &deleteTagsBatcher{batcher: batcher.NewBatcher(ctx, options)}
}

// DeleteTags adds delete tag input to batcher
func (b *deleteTagsBatcher) deleteTags(ctx context.Context, DeleteTagsInput *ec2.DeleteTagsInput) (*ec2.DeleteTagsOutput, error) {
	if len(DeleteTagsInput.Resources) != 1 {
		return nil, fmt.Errorf("expected to receive a single instance only, found %d", len(DeleteTagsInput.Resources))
	}
	result := b.batcher.Add(ctx, DeleteTagsInput)
	return result.Output, result.Err
}

// DeleteTagsHasher generates hash for different delete tag inputs
// Same set of tags have same hash, so they get executed together
func deleteTagsHasher(ctx context.Context, input *ec2.DeleteTagsInput) uint64 {
	hash, err := hashstructure.Hash(input.Tags, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		log.FromContext(ctx).Error(err, "failed hashing input tags")
	}
	return hash
}

func execDeleteTagsBatch(ec2api iface.EC2) batcher.BatchExecutor[ec2.DeleteTagsInput, ec2.DeleteTagsOutput] {
	return func(ctx context.Context, inputs []*ec2.DeleteTagsInput) []batcher.Result[ec2.DeleteTagsOutput] {
		results := make([]batcher.Result[ec2.DeleteTagsOutput], len(inputs))
		firstInput := inputs[0]
		// aggregate instanceIDs into 1 input
		for _, input := range inputs[1:] {
			firstInput.Resources = append(firstInput.Resources, input.Resources...)
		}
		batchedInput := &ec2.DeleteTagsInput{
			Resources: firstInput.Resources,
			Tags:      firstInput.Tags,
		}
		klog.Infof("Batched delete tags %v", batchedInput)
		output, err := ec2api.DeleteTags(batchedInput)

		if err != nil {
			klog.Errorf("Error occurred trying to batch tag resources, trying individually, %v", err)
			var wg sync.WaitGroup
			for idx, input := range inputs {
				wg.Add(1)
				go func(input *ec2.DeleteTagsInput) {
					defer wg.Done()
					out, err := ec2api.DeleteTags(input)
					results[idx] = batcher.Result[ec2.DeleteTagsOutput]{Output: out, Err: err}

				}(input)
			}
			wg.Wait()
		} else {
			for idx := range inputs {
				results[idx] = batcher.Result[ec2.DeleteTagsOutput]{Output: output}
			}
		}
		return results
	}
}
