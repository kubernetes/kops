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
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"k8s.io/client-go/tools/cache"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/config"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
	"k8s.io/klog/v2"
)

const instanceTopologyManagerCacheTimeout = 24 * time.Hour

/*
We need to ensure that instance types that we expect a response will not successfully complete syncing unless
we get a response, so we can track known instance types that we expect to get a response for.

https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-topology-prerequisites.html
*/
var defaultSupportedTopologyInstanceTypePattern = regexp.MustCompile(`^(hpc|trn|p)[0-9]+[a-z]*(-[a-z0-9]+)?(\.[0-9a-z]*)$`)

// stringKeyFunc is a string as cache key function
func topStringKeyFunc(obj interface{}) (string, error) {
	// Type should already be a string, so just return as is.
	s, ok := obj.(string)
	if !ok {
		return "", fmt.Errorf("failed to cast to string: %+v", obj)
	}

	return s, nil
}

// InstanceTopologyManager enables mocking the InstanceTopologyManager.
type InstanceTopologyManager interface {
	GetNodeTopology(ctx context.Context, instanceType string, region string, instanceID string) (*types.InstanceTopology, error)
	DoesInstanceTypeRequireResponse(instanceType string) bool
}

// instanceTopologyManager manages getting instance topology for nodes.
type instanceTopologyManager struct {
	ec2                                  iface.EC2
	unsupportedKeyStore                  cache.Store
	supportedTopologyInstanceTypePattern *regexp.Regexp
}

// NewInstanceTopologyManager generates a new InstanceTopologyManager.
func NewInstanceTopologyManager(ec2 iface.EC2, cfg *config.CloudConfig) InstanceTopologyManager {
	var supportedTopologyInstanceTypePattern *regexp.Regexp
	if cfg.Global.SupportedTopologyInstanceTypePattern != "" {
		supportedTopologyInstanceTypePattern = regexp.MustCompile(cfg.Global.SupportedTopologyInstanceTypePattern)
	} else {
		supportedTopologyInstanceTypePattern = defaultSupportedTopologyInstanceTypePattern
	}

	return &instanceTopologyManager{
		ec2:                                  ec2,
		supportedTopologyInstanceTypePattern: supportedTopologyInstanceTypePattern,
		// These should change very infrequently, if ever, so checking once a day sounds fair.
		unsupportedKeyStore: cache.NewTTLStore(topStringKeyFunc, instanceTopologyManagerCacheTimeout),
	}
}

// GetNodeTopology gets the instance topology for a node.
func (t *instanceTopologyManager) GetNodeTopology(ctx context.Context, instanceType string, region string, instanceID string) (*types.InstanceTopology, error) {
	if t.mightSupportTopology(instanceID, instanceType, region) {
		request := &ec2.DescribeInstanceTopologyInput{InstanceIds: []string{instanceID}}
		topologies, err := t.ec2.DescribeInstanceTopology(ctx, request)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				code := apiErr.ErrorCode()
				switch code {
				case "UnsupportedOperation":
					klog.Infof("ec2:DescribeInstanceTopology is not available in %s: %q", region, err)
					// If region is unsupported, track it to avoid making the call in the future.
					t.addUnsupported(region)
					return nil, nil
				case "UnauthorizedOperation":
					// Gracefully handle the DecribeInstanceTopology access missing error
					klog.Warningf("Not authorized to perform: ec2:DescribeInstanceTopology, permission missing: %q", err)
					// Mark region as unsupported to back off on attempts to get network topology.
					t.addUnsupported(region)
					return nil, nil
				case "RequestLimitExceeded":
					klog.Warningf("Exceeded ec2:DescribeInstanceTopology request limits. Try again later: %q", err)
					return nil, err
				}
			}

			// Unhandled error
			klog.Errorf("Error describing instance topology: %q", err)
			return nil, err
		} else if len(topologies) == 0 {
			// If no topology is returned, track the instance type as unsupported if we don't require a response.
			if t.DoesInstanceTypeRequireResponse(instanceType) {
				// While the instance type could be unsupported, it's also possible that the instance is deleting or shut down
				// and has no active instance topology. In this case, we don't want to track it as unsupported.
				klog.Warningf("Instance %s of type %s has no instance topology listed but may be a supported type.", instanceID, instanceType)
				// Track that the instance ID is does not include a response. This will prevent us from calling again unnecessarily.
				t.addUnsupported(instanceID)
			} else {
				klog.Infof("Instance type %s unsupported for getting instance topology", instanceType)
				t.addUnsupported(instanceType)
			}
			return nil, nil
		}

		return &topologies[0], nil
	}
	return nil, nil
}

// DoesInstanceTypeRequireResponse verifies whether or not we expect an instance to have an instance topology response.
func (t *instanceTopologyManager) DoesInstanceTypeRequireResponse(instanceType string) bool {
	return t.supportedTopologyInstanceTypePattern.MatchString(instanceType)
}

func (t *instanceTopologyManager) addUnsupported(key string) {
	err := t.unsupportedKeyStore.Add(key)
	if err != nil {
		klog.Errorf("Failed to cache unsupported key %s: %q", key, err)
	}
}

func (t *instanceTopologyManager) mightSupportTopology(instanceID string, instanceType string, region string) bool {
	// In the case of fargate and possibly other variants, the instance type will be empty.
	if len(instanceType) == 0 {
		return false
	}

	if _, exists, err := t.unsupportedKeyStore.GetByKey(region); exists {
		return false
	} else if err != nil {
		klog.Errorf("Failed to get cached unsupported region: %q:", err)
	}

	if _, exists, err := t.unsupportedKeyStore.GetByKey(instanceID); exists {
		return false
	} else if err != nil {
		klog.Errorf("Failed to get cached unsupported instance ID: %q:", err)
	}

	if _, exists, err := t.unsupportedKeyStore.GetByKey(instanceType); exists {
		return false
	} else if err != nil {
		klog.Errorf("Failed to get cached unsupported instance type: %q:", err)
	}

	return true
}
