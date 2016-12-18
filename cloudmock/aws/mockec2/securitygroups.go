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
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"strings"
)

func (m *MockEC2) CreateSecurityGroupRequest(*ec2.CreateSecurityGroupInput) (*request.Request, *ec2.CreateSecurityGroupOutput) {
	panic("MockEC2 CreateSecurityGroupRequest not implemented")
	return nil, nil
}

func (m *MockEC2) CreateSecurityGroup(request *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	glog.Infof("CreateSecurityGroup: %v", request)

	m.securityGroupNumber++
	n := m.securityGroupNumber

	sg := &ec2.SecurityGroup{
		GroupName:   request.GroupName,
		GroupId:     s(fmt.Sprintf("sg-%d", n)),
		VpcId:       request.VpcId,
		Description: request.Description,
	}
	m.SecurityGroups = append(m.SecurityGroups, sg)
	response := &ec2.CreateSecurityGroupOutput{
		GroupId: sg.GroupId,
	}
	return response, nil
}

func (m *MockEC2) DeleteSecurityGroupRequest(*ec2.DeleteSecurityGroupInput) (*request.Request, *ec2.DeleteSecurityGroupOutput) {
	panic("MockEC2 DeleteSecurityGroupRequest not implemented")
	return nil, nil
}
func (m *MockEC2) DeleteSecurityGroup(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	panic("MockEC2 DeleteSecurityGroup not implemented")
	return nil, nil
}

func (m *MockEC2) DescribeSecurityGroupReferencesRequest(*ec2.DescribeSecurityGroupReferencesInput) (*request.Request, *ec2.DescribeSecurityGroupReferencesOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeSecurityGroupReferences(*ec2.DescribeSecurityGroupReferencesInput) (*ec2.DescribeSecurityGroupReferencesOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeSecurityGroupsRequest(*ec2.DescribeSecurityGroupsInput) (*request.Request, *ec2.DescribeSecurityGroupsOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeSecurityGroups(request *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	glog.Infof("DescribeSecurityGroups: %v", request)

	var groups []*ec2.SecurityGroup

	for _, sg := range m.SecurityGroups {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "vpc-id":
				for _, v := range filter.Values {
					if sg.VpcId != nil && *sg.VpcId == *v {
						match = true
					}
				}

			case "group-name":
				for _, v := range filter.Values {
					if sg.GroupName != nil && *sg.GroupName == *v {
						match = true
					}
				}

			default:
				if strings.HasPrefix(*filter.Name, "tag:") {
					match = m.hasTag(ec2.ResourceTypeSecurityGroup, *sg.GroupId, filter)
				} else {
					return nil, fmt.Errorf("unknown filter name: %q", *filter.Name)
				}
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := *sg
		copy.Tags = m.getTags(ec2.ResourceTypeSecurityGroup, *sg.GroupId)
		groups = append(groups, &copy)
	}

	response := &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: groups,
	}

	return response, nil
}

func (m *MockEC2) DescribeStaleSecurityGroupsRequest(*ec2.DescribeStaleSecurityGroupsInput) (*request.Request, *ec2.DescribeStaleSecurityGroupsOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) DescribeStaleSecurityGroups(*ec2.DescribeStaleSecurityGroupsInput) (*ec2.DescribeStaleSecurityGroupsOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) RevokeSecurityGroupEgressRequest(*ec2.RevokeSecurityGroupEgressInput) (*request.Request, *ec2.RevokeSecurityGroupEgressOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) RevokeSecurityGroupEgress(*ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) RevokeSecurityGroupIngressRequest(*ec2.RevokeSecurityGroupIngressInput) (*request.Request, *ec2.RevokeSecurityGroupIngressOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) RevokeSecurityGroupIngress(*ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) AuthorizeSecurityGroupEgressRequest(*ec2.AuthorizeSecurityGroupEgressInput) (*request.Request, *ec2.AuthorizeSecurityGroupEgressOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) AuthorizeSecurityGroupEgress(*ec2.AuthorizeSecurityGroupEgressInput) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) AuthorizeSecurityGroupIngressRequest(*ec2.AuthorizeSecurityGroupIngressInput) (*request.Request, *ec2.AuthorizeSecurityGroupIngressOutput) {
	panic("Not implemented")
	return nil, nil
}
func (m *MockEC2) AuthorizeSecurityGroupIngress(*ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	panic("Not implemented")
	return nil, nil
}
