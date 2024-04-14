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

package mockec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/pki"
)

func (m *MockEC2) ImportKeyPair(ctx context.Context, request *ec2.ImportKeyPairInput, optFns ...func(*ec2.Options)) (*ec2.ImportKeyPairOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ImportKeyPair: %v", request)

	fp, err := pki.ComputeAWSKeyFingerprint(string(request.PublicKeyMaterial))
	if err != nil {
		return nil, err
	}

	n := len(m.KeyPairs) + 1
	id := fmt.Sprintf("key-%d", n)

	kp := &ec2types.KeyPairInfo{
		KeyFingerprint: aws.String(fp),
		KeyName:        request.KeyName,
		KeyPairId:      aws.String(id),
	}
	if m.KeyPairs == nil {
		m.KeyPairs = make(map[string]*ec2types.KeyPairInfo)
	}
	m.KeyPairs[id] = kp
	response := &ec2.ImportKeyPairOutput{
		KeyFingerprint: kp.KeyFingerprint,
		KeyName:        kp.KeyName,
	}

	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeKeyPair)...)

	return response, nil
}

func (m *MockEC2) DescribeKeyPairs(ctx context.Context, request *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeKeyPairs: %v", request)

	var keypairs []ec2types.KeyPairInfo

	for _, keypair := range m.KeyPairs {
		allFiltersMatch := true

		if len(request.KeyNames) != 0 {
			match := false
			for _, keyname := range request.KeyNames {
				if keyname == aws.ToString(keypair.KeyName) {
					match = true
				}
			}
			if !match {
				allFiltersMatch = false
			}
		}

		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {

			case "key-name":
				for _, v := range filter.Values {
					if aws.ToString(keypair.KeyName) == v {
						match = true
					}
				}
			default:
				return nil, fmt.Errorf("unknown filter name: %q", *filter.Name)
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := *keypair
		copy.Tags = m.getTags(ec2types.ResourceTypeKeyPair, *copy.KeyPairId)
		keypairs = append(keypairs, copy)
	}

	response := &ec2.DescribeKeyPairsOutput{
		KeyPairs: keypairs,
	}

	return response, nil
}

func (m *MockEC2) DeleteKeyPair(ctx context.Context, request *ec2.DeleteKeyPairInput, optFns ...func(*ec2.Options)) (*ec2.DeleteKeyPairOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteKeyPair: %v", request)

	keyID := aws.ToString(request.KeyPairId)
	found := false
	for id, kp := range m.KeyPairs {
		if aws.ToString(kp.KeyPairId) == keyID {
			found = true
			delete(m.KeyPairs, id)
		}
	}
	if !found {
		return nil, fmt.Errorf("KeyPairs %q not found", keyID)
	}

	return &ec2.DeleteKeyPairOutput{}, nil
}
