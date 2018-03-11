/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/golang/glog"

	"k8s.io/kops/pkg/pki"
)

func (m *MockEC2) DescribeKeyPairsRequest(*ec2.DescribeKeyPairsInput) (*request.Request, *ec2.DescribeKeyPairsOutput) {
	panic("MockEC2 DescribeKeyPairsRequest not implemented")
}
func (m *MockEC2) DescribeKeyPairsWithContext(aws.Context, *ec2.DescribeKeyPairsInput, ...request.Option) (*ec2.DescribeKeyPairsOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) ImportKeyPairRequest(*ec2.ImportKeyPairInput) (*request.Request, *ec2.ImportKeyPairOutput) {
	panic("MockEC2 ImportKeyPairRequest not implemented")
}
func (m *MockEC2) ImportKeyPairWithContext(aws.Context, *ec2.ImportKeyPairInput, ...request.Option) (*ec2.ImportKeyPairOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) ImportKeyPair(request *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error) {
	glog.Infof("ImportKeyPair: %v", request)

	fp, err := pki.ComputeAWSKeyFingerprint(string(request.PublicKeyMaterial))
	if err != nil {
		return nil, err
	}

	kp := &ec2.KeyPairInfo{
		KeyFingerprint: aws.String(fp),
		KeyName:        request.KeyName,
	}
	m.KeyPairs = append(m.KeyPairs, kp)
	response := &ec2.ImportKeyPairOutput{
		KeyFingerprint: kp.KeyFingerprint,
		KeyName:        kp.KeyName,
	}
	return response, nil
}
func (m *MockEC2) CreateKeyPairRequest(*ec2.CreateKeyPairInput) (*request.Request, *ec2.CreateKeyPairOutput) {
	panic("MockEC2 CreateKeyPairRequest not implemented")
}
func (m *MockEC2) CreateKeyPairWithContext(aws.Context, *ec2.CreateKeyPairInput, ...request.Option) (*ec2.CreateKeyPairOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) CreateKeyPair(*ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	panic("MockEC2 CreateKeyPair not implemented")
}

func (m *MockEC2) DescribeKeyPairs(request *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
	glog.Infof("DescribeKeyPairs: %v", request)

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
		keypairs = append(keypairs, &copy)
	}

	response := &ec2.DescribeKeyPairsOutput{
		KeyPairs: keypairs,
	}

	return response, nil
}
