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
)

func (m *MockEC2) CreateSecurityGroup(ctx context.Context, request *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateSecurityGroup: %v", request)

	m.securityGroupNumber++
	n := m.securityGroupNumber
	id := fmt.Sprintf("sg-%d", n)
	tags := tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeSecurityGroup)

	sg := &ec2types.SecurityGroup{
		GroupName:   request.GroupName,
		GroupId:     s(id),
		VpcId:       request.VpcId,
		Description: request.Description,
		Tags:        tags,
	}
	if m.SecurityGroups == nil {
		m.SecurityGroups = make(map[string]*ec2types.SecurityGroup)
	}
	m.SecurityGroups[*sg.GroupId] = sg

	m.addTags(id, tags...)

	response := &ec2.CreateSecurityGroupOutput{
		GroupId: sg.GroupId,
	}
	return response, nil
}

func (m *MockEC2) DeleteSecurityGroup(ctx context.Context, request *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteSecurityGroup: %v", request)

	id := aws.ToString(request.GroupId)
	o := m.SecurityGroups[id]
	if o == nil {
		return nil, fmt.Errorf("SecurityGroup %q not found", id)
	}
	delete(m.SecurityGroups, id)

	return &ec2.DeleteSecurityGroupOutput{}, nil
}

func (m *MockEC2) DescribeSecurityGroups(ctx context.Context, request *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeSecurityGroups: %v", request)

	if len(request.GroupIds) != 0 {
		request.Filters = append(request.Filters, ec2types.Filter{Name: s("group-id"), Values: request.GroupIds})
	}

	var groups []ec2types.SecurityGroup

	for _, sg := range m.SecurityGroups {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "vpc-id":
				for _, v := range filter.Values {
					if sg.VpcId != nil && *sg.VpcId == v {
						match = true
					}
				}

			case "group-name":
				for _, v := range filter.Values {
					if sg.GroupName != nil && *sg.GroupName == v {
						match = true
					}
				}
			case "group-id":
				for _, v := range filter.Values {
					if sg.GroupId != nil && *sg.GroupId == v {
						match = true
					}
				}

			default:
				match = m.hasTag(ec2types.ResourceTypeSecurityGroup, *sg.GroupId, filter)
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
		copy.Tags = m.getTags(ec2types.ResourceTypeSecurityGroup, *sg.GroupId)
		groups = append(groups, copy)
	}

	response := &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: groups,
	}

	return response, nil
}

func (m *MockEC2) RevokeSecurityGroupEgress(ctx context.Context, request *ec2.RevokeSecurityGroupEgressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) RevokeSecurityGroupIngress(ctx context.Context, request *ec2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("RevokeSecurityGroupIngress: %v", request)

	if aws.ToString(request.GroupId) == "" {
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

	klog.Warningf("RevokeSecurityGroupIngress mock not implemented - does not actually revoke permissions")

	response := &ec2.RevokeSecurityGroupIngressOutput{}
	return response, nil
}

func (m *MockEC2) AuthorizeSecurityGroupEgress(ctx context.Context, request *ec2.AuthorizeSecurityGroupEgressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AuthorizeSecurityGroupEgress: %v", request)

	if aws.ToString(request.GroupId) == "" {
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

		p := ec2types.IpPermission{
			FromPort:   request.FromPort,
			ToPort:     request.ToPort,
			IpProtocol: request.IpProtocol,
		}

		if request.CidrIp != nil {
			p.IpRanges = append(p.IpRanges, ec2types.IpRange{CidrIp: request.CidrIp})
		}

		sg.IpPermissionsEgress = append(sg.IpPermissionsEgress, p)
	}

	sg.IpPermissionsEgress = append(sg.IpPermissionsEgress, request.IpPermissions...)

	// TODO: We need to fold permissions

	if m.SecurityGroupRules == nil {
		m.SecurityGroupRules = make(map[string]*ec2types.SecurityGroupRule)
	}

	for _, permission := range request.IpPermissions {

		for _, iprange := range permission.IpRanges {

			n := len(m.SecurityGroupRules) + 1
			id := fmt.Sprintf("sgr-%d", n)
			rule := &ec2types.SecurityGroupRule{
				SecurityGroupRuleId: &id,
				GroupId:             sg.GroupId,
				FromPort:            permission.FromPort,
				ToPort:              permission.ToPort,
				IsEgress:            aws.Bool(true),
				CidrIpv4:            iprange.CidrIp,
				IpProtocol:          permission.IpProtocol,
				Tags:                tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeSecurityGroupRule),
			}
			if permission.FromPort == nil {
				rule.FromPort = aws.Int32(int32(-1))
			}
			if permission.ToPort == nil {
				rule.ToPort = aws.Int32(int32(-1))
			}

			m.SecurityGroupRules[id] = rule
		}

		for _, iprange := range permission.Ipv6Ranges {

			n := len(m.SecurityGroupRules) + 1
			id := fmt.Sprintf("sgr-%d", n)
			rule := &ec2types.SecurityGroupRule{
				SecurityGroupRuleId: &id,
				GroupId:             sg.GroupId,
				FromPort:            permission.FromPort,
				ToPort:              permission.ToPort,
				IsEgress:            aws.Bool(true),
				CidrIpv6:            iprange.CidrIpv6,
				IpProtocol:          permission.IpProtocol,
				Tags:                tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeSecurityGroupRule),
			}
			if permission.FromPort == nil {
				rule.FromPort = aws.Int32(int32(-1))
			}
			if permission.ToPort == nil {
				rule.ToPort = aws.Int32(int32(-1))
			}

			m.SecurityGroupRules[id] = rule
		}
	}

	response := &ec2.AuthorizeSecurityGroupEgressOutput{}
	return response, nil
}

func (m *MockEC2) AuthorizeSecurityGroupIngress(ctx context.Context, request *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AuthorizeSecurityGroupIngress: %v", request)

	if aws.ToString(request.GroupId) == "" {
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

		p := ec2types.IpPermission{
			FromPort:   request.FromPort,
			ToPort:     request.ToPort,
			IpProtocol: request.IpProtocol,
		}

		if request.CidrIp != nil {
			p.IpRanges = append(p.IpRanges, ec2types.IpRange{CidrIp: request.CidrIp})
		}

		sg.IpPermissions = append(sg.IpPermissions, p)
	}

	sg.IpPermissions = append(sg.IpPermissions, request.IpPermissions...)

	// TODO: We need to fold permissions

	if m.SecurityGroupRules == nil {
		m.SecurityGroupRules = make(map[string]*ec2types.SecurityGroupRule)
	}

	newSecurityGroupRule := func(permission ec2types.IpPermission) (string, *ec2types.SecurityGroupRule) {
		n := len(m.SecurityGroupRules) + 1
		id := fmt.Sprintf("sgr-%d", n)
		rule := &ec2types.SecurityGroupRule{
			SecurityGroupRuleId: &id,
			GroupId:             sg.GroupId,
			FromPort:            permission.FromPort,
			ToPort:              permission.ToPort,
			IsEgress:            aws.Bool(false),
			IpProtocol:          permission.IpProtocol,
			Tags:                tagSpecificationsToTags(request.TagSpecifications, ec2types.ResourceTypeSecurityGroupRule),
		}
		if permission.FromPort == nil {
			rule.FromPort = aws.Int32(int32(-1))
		}
		if permission.ToPort == nil {
			rule.ToPort = aws.Int32(int32(-1))
		}

		return id, rule
	}

	for _, permission := range request.IpPermissions {

		for _, iprange := range permission.IpRanges {
			id, rule := newSecurityGroupRule(permission)
			rule.CidrIpv4 = iprange.CidrIp
			m.SecurityGroupRules[id] = rule
		}

		for _, iprange := range permission.Ipv6Ranges {
			id, rule := newSecurityGroupRule(permission)
			rule.CidrIpv6 = iprange.CidrIpv6
			m.SecurityGroupRules[id] = rule
		}

		for _, prefixListId := range permission.PrefixListIds {
			id, rule := newSecurityGroupRule(permission)
			rule.PrefixListId = prefixListId.PrefixListId
			m.SecurityGroupRules[id] = rule

		}

		for _, group := range permission.UserIdGroupPairs {
			id, rule := newSecurityGroupRule(permission)
			rule.ReferencedGroupInfo = &ec2types.ReferencedSecurityGroup{
				GroupId: group.GroupId,
			}
			m.SecurityGroupRules[id] = rule
		}
	}

	response := &ec2.AuthorizeSecurityGroupIngressOutput{}
	return response, nil
}

func (m *MockEC2) DescribeSecurityGroupRules(ctx context.Context, request *ec2.DescribeSecurityGroupRulesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rules := []ec2types.SecurityGroupRule{}

	sgid := ""
	for _, filter := range request.Filters {
		if aws.ToString(filter.Name) == "group-id" {
			sgid = filter.Values[0]
		}
	}

	for _, rule := range m.SecurityGroupRules {
		if aws.ToString(rule.GroupId) == sgid {
			rules = append(rules, *rule)
		}
	}

	return &ec2.DescribeSecurityGroupRulesOutput{
		SecurityGroupRules: rules,
	}, nil
}
