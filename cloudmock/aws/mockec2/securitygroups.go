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
	"k8s.io/klog"
)

func (m *MockEC2) CreateSecurityGroupRequest(*ec2.CreateSecurityGroupInput) (*request.Request, *ec2.CreateSecurityGroupOutput) {
	panic("MockEC2 CreateSecurityGroupRequest not implemented")
}

func (m *MockEC2) CreateSecurityGroupWithContext(aws.Context, *ec2.CreateSecurityGroupInput, ...request.Option) (*ec2.CreateSecurityGroupOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) CreateSecurityGroup(request *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateSecurityGroup: %v", request)

	m.securityGroupNumber++
	n := m.securityGroupNumber

	sg := &ec2.SecurityGroup{
		GroupName:   request.GroupName,
		GroupId:     s(fmt.Sprintf("sg-%d", n)),
		VpcId:       request.VpcId,
		Description: request.Description,
	}
	if m.SecurityGroups == nil {
		m.SecurityGroups = make(map[string]*ec2.SecurityGroup)
	}
	m.SecurityGroups[*sg.GroupId] = sg

	response := &ec2.CreateSecurityGroupOutput{
		GroupId: sg.GroupId,
	}
	return response, nil
}

func (m *MockEC2) DeleteSecurityGroupRequest(request *ec2.DeleteSecurityGroupInput) (*request.Request, *ec2.DeleteSecurityGroupOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteSecurityGroupWithContext(aws.Context, *ec2.DeleteSecurityGroupInput, ...request.Option) (*ec2.DeleteSecurityGroupOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteSecurityGroup(request *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteSecurityGroup: %v", request)

	id := aws.StringValue(request.GroupId)
	o := m.SecurityGroups[id]
	if o == nil {
		return nil, fmt.Errorf("SecurityGroup %q not found", id)
	}
	delete(m.SecurityGroups, id)

	return &ec2.DeleteSecurityGroupOutput{}, nil
}

func (m *MockEC2) DescribeSecurityGroupReferencesRequest(*ec2.DescribeSecurityGroupReferencesInput) (*request.Request, *ec2.DescribeSecurityGroupReferencesOutput) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeSecurityGroupReferencesWithContext(aws.Context, *ec2.DescribeSecurityGroupReferencesInput, ...request.Option) (*ec2.DescribeSecurityGroupReferencesOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeSecurityGroupReferences(*ec2.DescribeSecurityGroupReferencesInput) (*ec2.DescribeSecurityGroupReferencesOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeSecurityGroupsRequest(*ec2.DescribeSecurityGroupsInput) (*request.Request, *ec2.DescribeSecurityGroupsOutput) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeSecurityGroupsWithContext(aws.Context, *ec2.DescribeSecurityGroupsInput, ...request.Option) (*ec2.DescribeSecurityGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeSecurityGroups(request *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeSecurityGroups: %v", request)

	if len(request.GroupIds) != 0 {
		request.Filters = append(request.Filters, &ec2.Filter{Name: s("group-id"), Values: request.GroupIds})
	}

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
			case "group-id":
				for _, v := range filter.Values {
					if sg.GroupId != nil && *sg.GroupId == *v {
						match = true
					}
				}

			default:
				match = m.hasTag(ec2.ResourceTypeSecurityGroup, *sg.GroupId, filter)
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
}
func (m *MockEC2) DescribeStaleSecurityGroupsWithContext(aws.Context, *ec2.DescribeStaleSecurityGroupsInput, ...request.Option) (*ec2.DescribeStaleSecurityGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeStaleSecurityGroups(*ec2.DescribeStaleSecurityGroupsInput) (*ec2.DescribeStaleSecurityGroupsOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) RevokeSecurityGroupEgressRequest(*ec2.RevokeSecurityGroupEgressInput) (*request.Request, *ec2.RevokeSecurityGroupEgressOutput) {
	panic("Not implemented")
}
func (m *MockEC2) RevokeSecurityGroupEgressWithContext(aws.Context, *ec2.RevokeSecurityGroupEgressInput, ...request.Option) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) RevokeSecurityGroupEgress(*ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) RevokeSecurityGroupIngressRequest(*ec2.RevokeSecurityGroupIngressInput) (*request.Request, *ec2.RevokeSecurityGroupIngressOutput) {
	panic("Not implemented")
}
func (m *MockEC2) RevokeSecurityGroupIngressWithContext(aws.Context, *ec2.RevokeSecurityGroupIngressInput, ...request.Option) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) RevokeSecurityGroupIngress(request *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("RevokeSecurityGroupIngress: %v", request)

	if aws.StringValue(request.GroupId) == "" {
		return nil, fmt.Errorf("GroupId not specified")
	}

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	if request.GroupName != nil {
		klog.Fatalf("GroupName not implemented")
	}
	sg := m.SecurityGroups[*request.GroupId]
	if sg == nil {
		return nil, fmt.Errorf("SecurityGroup not found")
	}

	klog.Warningf("RevokeSecurityGroupIngress not implemented - does not actually revoke permissions")

	response := &ec2.RevokeSecurityGroupIngressOutput{}
	return response, nil
}

func (m *MockEC2) AuthorizeSecurityGroupEgressRequest(*ec2.AuthorizeSecurityGroupEgressInput) (*request.Request, *ec2.AuthorizeSecurityGroupEgressOutput) {
	panic("Not implemented")
}
func (m *MockEC2) AuthorizeSecurityGroupEgressWithContext(aws.Context, *ec2.AuthorizeSecurityGroupEgressInput, ...request.Option) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) AuthorizeSecurityGroupEgress(request *ec2.AuthorizeSecurityGroupEgressInput) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AuthorizeSecurityGroupEgress: %v", request)

	if aws.StringValue(request.GroupId) == "" {
		return nil, fmt.Errorf("GroupId not specified")
	}

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	sg := m.SecurityGroups[*request.GroupId]
	if sg == nil {
		return nil, fmt.Errorf("sg not found")
	}

	if request.CidrIp != nil {
		if request.SourceSecurityGroupName != nil {
			klog.Fatalf("SourceSecurityGroupName not implemented")
		}
		if request.SourceSecurityGroupOwnerId != nil {
			klog.Fatalf("SourceSecurityGroupOwnerId not implemented")
		}

		p := &ec2.IpPermission{
			FromPort:   request.FromPort,
			ToPort:     request.ToPort,
			IpProtocol: request.IpProtocol,
		}

		if request.CidrIp != nil {
			p.IpRanges = append(p.IpRanges, &ec2.IpRange{CidrIp: request.CidrIp})
		}

		sg.IpPermissionsEgress = append(sg.IpPermissionsEgress, p)
	}

	sg.IpPermissionsEgress = append(sg.IpPermissionsEgress, request.IpPermissions...)

	// TODO: We need to fold permissions

	response := &ec2.AuthorizeSecurityGroupEgressOutput{}
	return response, nil
}
func (m *MockEC2) AuthorizeSecurityGroupIngressRequest(*ec2.AuthorizeSecurityGroupIngressInput) (*request.Request, *ec2.AuthorizeSecurityGroupIngressOutput) {
	panic("Not implemented")
}
func (m *MockEC2) AuthorizeSecurityGroupIngressWithContext(aws.Context, *ec2.AuthorizeSecurityGroupIngressInput, ...request.Option) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) AuthorizeSecurityGroupIngress(request *ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AuthorizeSecurityGroupIngress: %v", request)

	if aws.StringValue(request.GroupId) == "" {
		return nil, fmt.Errorf("GroupId not specified")
	}

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	if request.GroupName != nil {
		klog.Fatalf("GroupName not implemented")
	}
	sg := m.SecurityGroups[*request.GroupId]
	if sg == nil {
		return nil, fmt.Errorf("sg not found")
	}

	if request.CidrIp != nil {
		if request.SourceSecurityGroupName != nil {
			klog.Fatalf("SourceSecurityGroupName not implemented")
		}
		if request.SourceSecurityGroupOwnerId != nil {
			klog.Fatalf("SourceSecurityGroupOwnerId not implemented")
		}

		p := &ec2.IpPermission{
			FromPort:   request.FromPort,
			ToPort:     request.ToPort,
			IpProtocol: request.IpProtocol,
		}

		if request.CidrIp != nil {
			p.IpRanges = append(p.IpRanges, &ec2.IpRange{CidrIp: request.CidrIp})
		}

		sg.IpPermissions = append(sg.IpPermissions, p)
	}

	sg.IpPermissions = append(sg.IpPermissions, request.IpPermissions...)

	// TODO: We need to fold permissions

	response := &ec2.AuthorizeSecurityGroupIngressOutput{}
	return response, nil
}
