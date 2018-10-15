/*
Copyright 2018 The Kubernetes Authors.

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

package spotinst

import (
	"context"

	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"k8s.io/kops/upup/pkg/fi"
)

type awsService struct {
	svc aws.Service
}

// List returns a list of Elastigroups.
func (s *awsService) List(ctx context.Context) ([]Elastigroup, error) {
	output, err := s.svc.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	groups := make([]Elastigroup, len(output.Groups))
	for i, group := range output.Groups {
		groups[i] = &awsElastigroup{group}
	}

	return groups, nil
}

// Create creates a new Elastigroup and returns its ID.
func (s *awsService) Create(ctx context.Context, group Elastigroup) (string, error) {
	input := &aws.CreateGroupInput{
		Group: group.Obj().(*aws.Group),
	}

	output, err := s.svc.Create(ctx, input)
	if err != nil {
		return "", err
	}

	return fi.StringValue(output.Group.ID), nil
}

// Read returns an existing Elastigroup by ID.
func (s *awsService) Read(ctx context.Context, groupID string) (Elastigroup, error) {
	input := &aws.ReadGroupInput{
		GroupID: fi.String(groupID),
	}

	output, err := s.svc.Read(ctx, input)
	if err != nil {
		return nil, err
	}

	return &awsElastigroup{output.Group}, nil
}

// Update updates an existing Elastigroup.
func (s *awsService) Update(ctx context.Context, group Elastigroup) error {
	input := &aws.UpdateGroupInput{
		Group: group.Obj().(*aws.Group),
	}

	_, err := s.svc.Update(ctx, input)
	return err

}

// Delete deletes an existing Elastigroup by ID.
func (s *awsService) Delete(ctx context.Context, groupID string) error {
	input := &aws.DeleteGroupInput{
		GroupID: fi.String(groupID),
	}

	_, err := s.svc.Delete(ctx, input)
	return err
}

// Detach removes one or more instances from the specified Elastigroup.
func (s *awsService) Detach(ctx context.Context, groupID string, instanceIDs []string) error {
	input := &aws.DetachGroupInput{
		GroupID:                       fi.String(groupID),
		InstanceIDs:                   instanceIDs,
		ShouldDecrementTargetCapacity: fi.Bool(false),
		ShouldTerminateInstances:      fi.Bool(true),
	}

	_, err := s.svc.Detach(ctx, input)
	return err
}

// Instances returns a list of all instances that belong to specified Elastigroup.
func (s *awsService) Instances(ctx context.Context, groupID string) ([]Instance, error) {
	input := &aws.StatusGroupInput{
		GroupID: fi.String(groupID),
	}

	output, err := s.svc.Status(ctx, input)
	if err != nil {
		return nil, err
	}

	instances := make([]Instance, len(output.Instances))
	for i, instance := range output.Instances {
		instances[i] = &awsInstance{instance}
	}

	return instances, err
}

type awsElastigroup struct {
	obj *aws.Group
}

// Id returns the ID of the Elastigroup.
func (e *awsElastigroup) Id() string { return fi.StringValue(e.obj.ID) }

// Name returns the name of the Elastigroup.
func (e *awsElastigroup) Name() string { return fi.StringValue(e.obj.Name) }

// MinSize returns the minimum size of the Elastigroup.
func (e *awsElastigroup) MinSize() int { return fi.IntValue(e.obj.Capacity.Minimum) }

// MaxSize returns the maximum size of the Elastigroup.
func (e *awsElastigroup) MaxSize() int { return fi.IntValue(e.obj.Capacity.Maximum) }

// Obj returns the raw object which is a cloud-specific implementation.
func (e *awsElastigroup) Obj() interface{} { return e.obj }

type awsInstance struct {
	obj *aws.Instance
}

// Id returns the ID of the instance.
func (i *awsInstance) Id() string { return fi.StringValue(i.obj.ID) }

// Obj returns the raw object which is a cloud-specific implementation.
func (i *awsInstance) Obj() interface{} { return i.obj }
