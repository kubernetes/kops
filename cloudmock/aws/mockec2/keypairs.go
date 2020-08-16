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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/pki"
)

func (m *MockEC2) DescribeKeyPairsRequest(*ec2.DescribeKeyPairsInput) (*request.Request, *ec2.DescribeKeyPairsOutput) {
	panic("MockEC2 DescribeKeyPairsRequest not implemented")
}
func (m *MockEC2) DescribeKeyPairsWithContext(aws.Context, *ec2.DescribeKeyPairsInput, ...request.Option) (*ec2.DescribeKeyPairsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) ImportKeyPairRequest(*ec2.ImportKeyPairInput) (*request.Request, *ec2.ImportKeyPairOutput) {
	panic("MockEC2 ImportKeyPairRequest not implemented")
}
func (m *MockEC2) ImportKeyPairWithContext(aws.Context, *ec2.ImportKeyPairInput, ...request.Option) (*ec2.ImportKeyPairOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) ImportKeyPair(request *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ImportKeyPair: %v", request)

	fp, err := pki.ComputeAWSKeyFingerprint(string(request.PublicKeyMaterial))
	if err != nil {
		return nil, err
	}

	n := len(m.KeyPairs) + 1
	id := fmt.Sprintf("key-%d", n)

	kp := &ec2.KeyPairInfo{
		KeyFingerprint: aws.String(fp),
		KeyName:        request.KeyName,
		KeyPairId:      aws.String(id),
	}
	if m.KeyPairs == nil {
		m.KeyPairs = make(map[string]*ec2.KeyPairInfo)
	}
	m.KeyPairs[id] = kp
	response := &ec2.ImportKeyPairOutput{
		KeyFingerprint: kp.KeyFingerprint,
		KeyName:        kp.KeyName,
	}

	m.addTags(id, tagSpecificationsToTags(request.TagSpecifications, ec2.ResourceTypeKeyPair)...)

	return response, nil
}
func (m *MockEC2) CreateKeyPairRequest(*ec2.CreateKeyPairInput) (*request.Request, *ec2.CreateKeyPairOutput) {
	panic("MockEC2 CreateKeyPairRequest not implemented")
}
func (m *MockEC2) CreateKeyPairWithContext(aws.Context, *ec2.CreateKeyPairInput, ...request.Option) (*ec2.CreateKeyPairOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) CreateKeyPair(*ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	panic("MockEC2 CreateKeyPair not implemented")
}

func (m *MockEC2) DescribeKeyPairs(request *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeKeyPairs: %v", request)

	var keypairs []*ec2.KeyPairInfo

	for _, keypair := range m.KeyPairs {
		allFiltersMatch := true

		if len(request.KeyNames) != 0 {
			match := false
			for _, keyname := range request.KeyNames {
				if aws.StringValue(keyname) == aws.StringValue(keypair.KeyName) {
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
					if aws.StringValue(keypair.KeyName) == aws.StringValue(v) {
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
		copy.Tags = m.getTags(ec2.ResourceTypeKeyPair, *copy.KeyPairId)
		keypairs = append(keypairs, &copy)
	}

	response := &ec2.DescribeKeyPairsOutput{
		KeyPairs: keypairs,
	}

	return response, nil
}

func (m *MockEC2) DeleteKeyPair(request *ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteKeyPair: %v", request)

	keyID := aws.StringValue(request.KeyPairId)
	found := false
	for id, kp := range m.KeyPairs {
		if aws.StringValue(kp.KeyPairId) == keyID {
			found = true
			delete(m.KeyPairs, id)
		}
	}
	if !found {
		return nil, fmt.Errorf("KeyPairs %q not found", keyID)
	}

	return &ec2.DeleteKeyPairOutput{}, nil
}

func (m *MockEC2) DeleteKeyPairWithContext(aws.Context, *ec2.DeleteKeyPairInput, ...request.Option) (*ec2.DeleteKeyPairOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DeleteKeyPairRequest(*ec2.DeleteKeyPairInput) (*request.Request, *ec2.DeleteKeyPairOutput) {
	panic("Not implemented")
}
