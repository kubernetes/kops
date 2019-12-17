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

package spotinst

import (
	"context"
	"time"

	awseg "github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	awsoc "github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"k8s.io/kops/upup/pkg/fi"
)

type awsCloud struct {
	eg InstanceGroupService
	oc InstanceGroupService
	ls LaunchSpecService
}

func (x *awsCloud) Elastigroup() InstanceGroupService { return x.eg }
func (x *awsCloud) Ocean() InstanceGroupService       { return x.oc }
func (x *awsCloud) LaunchSpec() LaunchSpecService     { return x.ls }

type awsElastigroupService struct {
	svc awseg.Service
}

// List returns a list of InstanceGroups.
func (x *awsElastigroupService) List(ctx context.Context) ([]InstanceGroup, error) {
	output, err := x.svc.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	groups := make([]InstanceGroup, len(output.Groups))
	for i, group := range output.Groups {
		groups[i] = &awsElastigroupInstanceGroup{group}
	}

	return groups, nil
}

// Create creates a new InstanceGroup and returns its ID.
func (x *awsElastigroupService) Create(ctx context.Context, group InstanceGroup) (string, error) {
	input := &awseg.CreateGroupInput{
		Group: group.Obj().(*awseg.Group),
	}

	output, err := x.svc.Create(ctx, input)
	if err != nil {
		return "", err
	}

	return fi.StringValue(output.Group.ID), nil
}

// Read returns an existing InstanceGroup by ID.
func (x *awsElastigroupService) Read(ctx context.Context, groupID string) (InstanceGroup, error) {
	input := &awseg.ReadGroupInput{
		GroupID: fi.String(groupID),
	}

	output, err := x.svc.Read(ctx, input)
	if err != nil {
		return nil, err
	}

	return &awsElastigroupInstanceGroup{output.Group}, nil
}

// Update updates an existing InstanceGroup.
func (x *awsElastigroupService) Update(ctx context.Context, group InstanceGroup) error {
	input := &awseg.UpdateGroupInput{
		Group: group.Obj().(*awseg.Group),
	}

	_, err := x.svc.Update(ctx, input)
	return err

}

// Delete deletes an existing InstanceGroup by ID.
func (x *awsElastigroupService) Delete(ctx context.Context, groupID string) error {
	input := &awseg.DeleteGroupInput{
		GroupID: fi.String(groupID),
	}

	_, err := x.svc.Delete(ctx, input)
	return err
}

// Detach removes one or more instances from the specified InstanceGroup.
func (x *awsElastigroupService) Detach(ctx context.Context, groupID string, instanceIDs []string) error {
	input := &awseg.DetachGroupInput{
		GroupID:                       fi.String(groupID),
		InstanceIDs:                   instanceIDs,
		ShouldDecrementTargetCapacity: fi.Bool(false),
		ShouldTerminateInstances:      fi.Bool(true),
	}

	_, err := x.svc.Detach(ctx, input)
	return err
}

// Instances returns a list of all instances that belong to specified InstanceGroup.
func (x *awsElastigroupService) Instances(ctx context.Context, groupID string) ([]Instance, error) {
	input := &awseg.StatusGroupInput{
		GroupID: fi.String(groupID),
	}

	output, err := x.svc.Status(ctx, input)
	if err != nil {
		return nil, err
	}

	instances := make([]Instance, len(output.Instances))
	for i, instance := range output.Instances {
		instances[i] = &awsElastigroupInstance{instance}
	}

	return instances, err
}

type awsOceanService struct {
	svc awsoc.Service
}

// List returns a list of InstanceGroups.
func (x *awsOceanService) List(ctx context.Context) ([]InstanceGroup, error) {
	output, err := x.svc.ListClusters(ctx, nil)
	if err != nil {
		return nil, err
	}

	groups := make([]InstanceGroup, len(output.Clusters))
	for i, group := range output.Clusters {
		groups[i] = &awsOceanInstanceGroup{group}
	}

	return groups, nil
}

// Create creates a new InstanceGroup and returns its ID.
func (x *awsOceanService) Create(ctx context.Context, group InstanceGroup) (string, error) {
	input := &awsoc.CreateClusterInput{
		Cluster: group.Obj().(*awsoc.Cluster),
	}

	output, err := x.svc.CreateCluster(ctx, input)
	if err != nil {
		return "", err
	}

	return fi.StringValue(output.Cluster.ID), nil
}

// Read returns an existing InstanceGroup by ID.
func (x *awsOceanService) Read(ctx context.Context, clusterID string) (InstanceGroup, error) {
	input := &awsoc.ReadClusterInput{
		ClusterID: fi.String(clusterID),
	}

	output, err := x.svc.ReadCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	return &awsOceanInstanceGroup{output.Cluster}, nil
}

// Update updates an existing InstanceGroup.
func (x *awsOceanService) Update(ctx context.Context, group InstanceGroup) error {
	input := &awsoc.UpdateClusterInput{
		Cluster: group.Obj().(*awsoc.Cluster),
	}

	_, err := x.svc.UpdateCluster(ctx, input)
	return err

}

// Delete deletes an existing InstanceGroup by ID.
func (x *awsOceanService) Delete(ctx context.Context, clusterID string) error {
	input := &awsoc.DeleteClusterInput{
		ClusterID: fi.String(clusterID),
	}

	_, err := x.svc.DeleteCluster(ctx, input)
	return err
}

// Detach removes one or more instances from the specified InstanceGroup.
func (x *awsOceanService) Detach(ctx context.Context, clusterID string, instanceIDs []string) error {
	input := &awsoc.DetachClusterInstancesInput{
		ClusterID:                     fi.String(clusterID),
		InstanceIDs:                   instanceIDs,
		ShouldDecrementTargetCapacity: fi.Bool(false),
		ShouldTerminateInstances:      fi.Bool(true),
	}

	_, err := x.svc.DetachClusterInstances(ctx, input)
	return err
}

// Instances returns a list of all instances that belong to specified InstanceGroup.
func (x *awsOceanService) Instances(ctx context.Context, clusterID string) ([]Instance, error) {
	input := &awsoc.ListClusterInstancesInput{
		ClusterID: fi.String(clusterID),
	}

	output, err := x.svc.ListClusterInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	instances := make([]Instance, len(output.Instances))
	for i, instance := range output.Instances {
		instances[i] = &awsOceanInstance{instance}
	}

	return instances, err
}

type awsOceanLaunchSpecService struct {
	svc awsoc.Service
}

// List returns a list of LaunchSpecs.
func (x *awsOceanLaunchSpecService) List(ctx context.Context, oceanID string) ([]LaunchSpec, error) {
	input := &awsoc.ListLaunchSpecsInput{
		OceanID: fi.String(oceanID),
	}

	output, err := x.svc.ListLaunchSpecs(ctx, input)
	if err != nil {
		return nil, err
	}

	specs := make([]LaunchSpec, len(output.LaunchSpecs))
	for i, spec := range output.LaunchSpecs {
		specs[i] = &awsOceanLaunchSpec{spec}
	}

	return specs, nil
}

// Create creates a new LaunchSpec and returns its ID.
func (x *awsOceanLaunchSpecService) Create(ctx context.Context, spec LaunchSpec) (string, error) {
	input := &awsoc.CreateLaunchSpecInput{
		LaunchSpec: spec.Obj().(*awsoc.LaunchSpec),
	}

	output, err := x.svc.CreateLaunchSpec(ctx, input)
	if err != nil {
		return "", err
	}

	return fi.StringValue(output.LaunchSpec.ID), nil
}

// Read returns an existing LaunchSpec by ID.
func (x *awsOceanLaunchSpecService) Read(ctx context.Context, specID string) (LaunchSpec, error) {
	input := &awsoc.ReadLaunchSpecInput{
		LaunchSpecID: fi.String(specID),
	}

	output, err := x.svc.ReadLaunchSpec(ctx, input)
	if err != nil {
		return nil, err
	}

	return &awsOceanLaunchSpec{output.LaunchSpec}, nil
}

// Update updates an existing LaunchSpec.
func (x *awsOceanLaunchSpecService) Update(ctx context.Context, spec LaunchSpec) error {
	input := &awsoc.UpdateLaunchSpecInput{
		LaunchSpec: spec.Obj().(*awsoc.LaunchSpec),
	}

	_, err := x.svc.UpdateLaunchSpec(ctx, input)
	return err

}

// Delete deletes an existing LaunchSpec by ID.
func (x *awsOceanLaunchSpecService) Delete(ctx context.Context, specID string) error {
	input := &awsoc.DeleteLaunchSpecInput{
		LaunchSpecID: fi.String(specID),
	}

	_, err := x.svc.DeleteLaunchSpec(ctx, input)
	return err
}

type awsElastigroupInstanceGroup struct {
	obj *awseg.Group
}

// Id returns the ID of the InstanceGroup.
func (x *awsElastigroupInstanceGroup) Id() string {
	return fi.StringValue(x.obj.ID)
}

// Type returns the type of the InstanceGroup.
func (x *awsElastigroupInstanceGroup) Type() InstanceGroupType {
	return InstanceGroupElastigroup
}

// Name returns the name of the InstanceGroup.
func (x *awsElastigroupInstanceGroup) Name() string {
	return fi.StringValue(x.obj.Name)
}

// MinSize returns the minimum size of the InstanceGroup.
func (x *awsElastigroupInstanceGroup) MinSize() int {
	return fi.IntValue(x.obj.Capacity.Minimum)
}

// MaxSize returns the maximum size of the InstanceGroup.
func (x *awsElastigroupInstanceGroup) MaxSize() int {
	return fi.IntValue(x.obj.Capacity.Maximum)
}

// CreatedAt returns the timestamp when the InstanceGroup has been created.
func (x *awsElastigroupInstanceGroup) CreatedAt() time.Time {
	return spotinst.TimeValue(x.obj.CreatedAt)
}

// UpdatedAt returns the timestamp when the InstanceGroup has been updated.
func (x *awsElastigroupInstanceGroup) UpdatedAt() time.Time {
	return spotinst.TimeValue(x.obj.UpdatedAt)
}

// Obj returns the raw object which is a cloud-specific implementation.
func (x *awsElastigroupInstanceGroup) Obj() interface{} {
	return x.obj
}

type awsElastigroupInstance struct {
	obj *awseg.Instance
}

// Id returns the ID of the instance.
func (x *awsElastigroupInstance) Id() string {
	return fi.StringValue(x.obj.ID)
}

// CreatedAt returns the timestamp when the Instance has been created.
func (x *awsElastigroupInstance) CreatedAt() time.Time {
	return spotinst.TimeValue(x.obj.CreatedAt)
}

// Obj returns the raw object which is a cloud-specific implementation.
func (x *awsElastigroupInstance) Obj() interface{} {
	return x.obj
}

type awsOceanInstanceGroup struct {
	obj *awsoc.Cluster
}

// Id returns the ID of the InstanceGroup.
func (x *awsOceanInstanceGroup) Id() string {
	return fi.StringValue(x.obj.ID)
}

// Type returns the type of the InstanceGroup.
func (x *awsOceanInstanceGroup) Type() InstanceGroupType {
	return InstanceGroupOcean
}

// Name returns the name of the InstanceGroup.
func (x *awsOceanInstanceGroup) Name() string {
	return fi.StringValue(x.obj.Name)
}

// MinSize returns the minimum size of the InstanceGroup.
func (x *awsOceanInstanceGroup) MinSize() int {
	return fi.IntValue(x.obj.Capacity.Minimum)
}

// MaxSize returns the maximum size of the InstanceGroup.
func (x *awsOceanInstanceGroup) MaxSize() int {
	return fi.IntValue(x.obj.Capacity.Maximum)
}

// CreatedAt returns the timestamp when the InstanceGroup has been created.
func (x *awsOceanInstanceGroup) CreatedAt() time.Time {
	return spotinst.TimeValue(x.obj.CreatedAt)
}

// UpdatedAt returns the timestamp when the InstanceGroup has been updated.
func (x *awsOceanInstanceGroup) UpdatedAt() time.Time {
	return spotinst.TimeValue(x.obj.UpdatedAt)
}

// Obj returns the raw object which is a cloud-specific implementation.
func (x *awsOceanInstanceGroup) Obj() interface{} {
	return x.obj
}

type awsOceanInstance struct {
	obj *awsoc.Instance
}

// Id returns the ID of the instance.
func (x *awsOceanInstance) Id() string {
	return fi.StringValue(x.obj.ID)
}

// CreatedAt returns the timestamp when the Instance has been created.
func (x *awsOceanInstance) CreatedAt() time.Time {
	return spotinst.TimeValue(x.obj.CreatedAt)
}

// Obj returns the raw object which is a cloud-specific implementation.
func (x *awsOceanInstance) Obj() interface{} {
	return x.obj
}

type awsOceanLaunchSpec struct {
	obj *awsoc.LaunchSpec
}

// Id returns the ID of the LaunchSpec.
func (x *awsOceanLaunchSpec) Id() string {
	return fi.StringValue(x.obj.ID)
}

// Name returns the name of the LaunchSpec.
func (x *awsOceanLaunchSpec) Name() string {
	return fi.StringValue(x.obj.Name)
}

// OceanId returns the ID of the Ocean instance group.
func (x *awsOceanLaunchSpec) OceanId() string {
	return fi.StringValue(x.obj.OceanID)
}

// CreatedAt returns the timestamp when the LaunchSpec has been created.
func (x *awsOceanLaunchSpec) CreatedAt() time.Time {
	return spotinst.TimeValue(x.obj.CreatedAt)
}

// UpdatedAt returns the timestamp when the LaunchSpec has been updated.
func (x *awsOceanLaunchSpec) UpdatedAt() time.Time {
	return spotinst.TimeValue(x.obj.UpdatedAt)
}

// Obj returns the raw object which is a cloud-specific implementation.
func (x *awsOceanLaunchSpec) Obj() interface{} { return x.obj }
