/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mock encapsulates mocks for testing GCE provider functionality.
// These methods are used to override the mock objects' methods in order to
// intercept the standard processing and to add custom logic for test purposes.
//
//	// Example usage:
//
// cloud := cloud.NewMockGCE()
// cloud.MockTargetPools.AddInstanceHook = mock.AddInstanceHook
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	cloud "github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/filter"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	ga "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

var (
	// InUseError is a shared variable with error code StatusBadRequest for error verification.
	InUseError = &googleapi.Error{Code: http.StatusBadRequest, Message: "It's being used by god."}
	// InternalServerError is shared variable with error code StatusInternalServerError for error verification.
	InternalServerError = &googleapi.Error{Code: http.StatusInternalServerError}
	// UnauthorizedErr wraps a Google API error with code StatusForbidden.
	UnauthorizedErr = &googleapi.Error{Code: http.StatusForbidden}
)

// gceObject is an abstraction of all GCE API object in go client
type gceObject interface {
	MarshalJSON() ([]byte, error)
}

// AttachDiskHook mocks attaching a disk to an instance
func AttachDiskHook(ctx context.Context, key *meta.Key, req *ga.AttachedDisk, m *cloud.MockInstances, options ...cloud.Option) error {
	instance, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	instance.Disks = append(instance.Disks, req)
	return nil
}

// Verify AttachDiskHook implements cloud.MockInstances.AttachDiskHook.
var _ = cloud.MockInstances{
	AttachDiskHook: AttachDiskHook,
}

// DetachDiskHook mocks detaching a disk from an instance
func DetachDiskHook(ctx context.Context, key *meta.Key, diskName string, m *cloud.MockInstances, options ...cloud.Option) error {
	instance, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	for i, disk := range instance.Disks {
		if disk.DeviceName == diskName {
			instance.Disks = append(instance.Disks[:i], instance.Disks[i+1:]...)
			return nil
		}
	}
	return &googleapi.Error{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("Disk: %s was not found in Instance %s", diskName, key.String()),
	}
}

// Verify DetachDiskHook implements cloud.MockInstances.DetachDiskHook.
var _ = cloud.MockInstances{
	DetachDiskHook: DetachDiskHook,
}

// AddInstanceHook mocks adding a Instance to MockTargetPools
func AddInstanceHook(ctx context.Context, key *meta.Key, req *ga.TargetPoolsAddInstanceRequest, m *cloud.MockTargetPools, options ...cloud.Option) error {
	pool, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in TargetPools", key.String()),
		}
	}

	for _, instance := range req.Instances {
		pool.Instances = append(pool.Instances, instance.Instance)
	}

	return nil
}

// Verify AddInstanceHook implements cloud.MockTargetPools.AddInstanceHook.
var _ = cloud.MockTargetPools{
	AddInstanceHook: AddInstanceHook,
}

// RemoveInstanceHook mocks removing a Instance from MockTargetPools
func RemoveInstanceHook(ctx context.Context, key *meta.Key, req *ga.TargetPoolsRemoveInstanceRequest, m *cloud.MockTargetPools, options ...cloud.Option) error {
	pool, err := m.Get(ctx, key)
	if err != nil {
		return &googleapi.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("Key: %s was not found in TargetPools", key.String()),
		}
	}

	for _, instanceToRemove := range req.Instances {
		for i, instance := range pool.Instances {
			if instanceToRemove.Instance == instance {
				// Delete instance from pool.Instances without preserving order
				pool.Instances[i] = pool.Instances[len(pool.Instances)-1]
				pool.Instances = pool.Instances[:len(pool.Instances)-1]
				break
			}
		}
	}

	return nil
}

// Verify RemoveInstanceHook implements cloud.MockTargetPools.RemoveInstanceHook.
var _ = cloud.MockTargetPools{
	RemoveInstanceHook: RemoveInstanceHook,
}

func convertAndInsertAlphaForwardingRule(key *meta.Key, obj gceObject, mRules map[meta.Key]*cloud.MockForwardingRulesObj, version meta.Version, projectID string) (bool, error) {
	if !key.Valid() {
		return true, fmt.Errorf("invalid GCE key (%+v)", key)
	}

	if _, ok := mRules[*key]; ok {
		err := &googleapi.Error{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("MockForwardingRule %v exists", key),
		}
		return true, err
	}

	enc, err := obj.MarshalJSON()
	if err != nil {
		return true, err
	}
	var fwdRule alpha.ForwardingRule
	if err := json.Unmarshal(enc, &fwdRule); err != nil {
		return true, err
	}
	// Set the default values for the Alpha fields.
	if fwdRule.NetworkTier == "" {
		fwdRule.NetworkTier = cloud.NetworkTierDefault.ToGCEValue()
	}

	fwdRule.Name = key.Name
	if fwdRule.SelfLink == "" {
		fwdRule.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, version, projectID, "forwardingRules", key)
	}

	mRules[*key] = &cloud.MockForwardingRulesObj{Obj: fwdRule}
	return true, nil
}

// InsertFwdRuleHook mocks inserting a ForwardingRule. ForwardingRules are
// expected to default to Premium tier if no NetworkTier is specified.
func InsertFwdRuleHook(ctx context.Context, key *meta.Key, obj *ga.ForwardingRule, m *cloud.MockForwardingRules, options ...cloud.Option) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionGA, "forwardingRules")
	return convertAndInsertAlphaForwardingRule(key, obj, m.Objects, meta.VersionGA, projectID)
}

// Verify InsertFwdRuleHook implements cloud.MockForwardingRules.InsertHook.
var _ = cloud.MockForwardingRules{
	InsertHook: InsertFwdRuleHook,
}

// InsertBetaFwdRuleHook mocks inserting a BetaForwardingRule.
func InsertBetaFwdRuleHook(ctx context.Context, key *meta.Key, obj *beta.ForwardingRule, m *cloud.MockBetaForwardingRules, options ...cloud.Option) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionBeta, "forwardingRules")
	return convertAndInsertAlphaForwardingRule(key, obj, m.Objects, meta.VersionBeta, projectID)
}

// Verify InsertBetaFwdRuleHook implements cloud.MockBetaForwardingRules.InsertHook.
var _ = cloud.MockBetaForwardingRules{
	InsertHook: InsertBetaFwdRuleHook,
}

// InsertAlphaFwdRuleHook mocks inserting an AlphaForwardingRule.
func InsertAlphaFwdRuleHook(ctx context.Context, key *meta.Key, obj *alpha.ForwardingRule, m *cloud.MockAlphaForwardingRules, options ...cloud.Option) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionAlpha, "forwardingRules")
	return convertAndInsertAlphaForwardingRule(key, obj, m.Objects, meta.VersionAlpha, projectID)
}

// Verify InsertAlphaFwdRuleHook implements cloud.MockAlphaForwardingRules.InsertHook.
var _ = cloud.MockAlphaForwardingRules{
	InsertHook: InsertAlphaFwdRuleHook,
}

// AddressAttributes maps from Address key to a map of Instances
type AddressAttributes struct {
	IPCounter int // Used to assign Addresses with no IP a unique IP address
}

func convertAndInsertAlphaAddress(key *meta.Key, obj gceObject, mAddrs map[meta.Key]*cloud.MockAddressesObj, version meta.Version, projectID string, addressAttrs AddressAttributes) (bool, error) {
	if !key.Valid() {
		return true, fmt.Errorf("invalid GCE key (%+v)", key)
	}

	if _, ok := mAddrs[*key]; ok {
		err := &googleapi.Error{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("MockAddresses %v exists", key),
		}
		return true, err
	}

	enc, err := obj.MarshalJSON()
	if err != nil {
		return true, err
	}
	var addr alpha.Address
	if err := json.Unmarshal(enc, &addr); err != nil {
		return true, err
	}

	// Set default address type if not present.
	if addr.AddressType == "" {
		addr.AddressType = string(cloud.SchemeExternal)
	}

	var existingAddresses []*ga.Address
	for _, obj := range mAddrs {
		existingAddresses = append(existingAddresses, obj.ToGA())
	}

	for _, existingAddr := range existingAddresses {
		if addr.Address == existingAddr.Address {
			msg := fmt.Sprintf("MockAddresses IP %v in use", addr.Address)

			// When the IP is already in use, this call returns a StatusBadRequest
			// if the address is an external address, and StatusConflict if an
			// internal address. This is to be consistent with actual GCE API.
			errorCode := http.StatusConflict
			if addr.AddressType == string(cloud.SchemeExternal) {
				errorCode = http.StatusBadRequest
			}

			return true, &googleapi.Error{Code: errorCode, Message: msg}
		}
	}

	// Set default values used in tests
	addr.Name = key.Name
	if addr.SelfLink == "" {
		addr.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, version, projectID, "addresses", key)
	}

	if addr.Address == "" {
		if addr.IpVersion == "IPV6" {
			addr.Address = fmt.Sprintf("1111:2222:3333:4444:5555:%d:0:0", addressAttrs.IPCounter)
			addr.PrefixLength = 96
		} else {
			addr.Address = fmt.Sprintf("1.2.3.%d", addressAttrs.IPCounter)
		}
		addressAttrs.IPCounter++
	}

	// Set the default values for the Alpha fields.
	if addr.NetworkTier == "" {
		addr.NetworkTier = cloud.NetworkTierDefault.ToGCEValue()
	}

	mAddrs[*key] = &cloud.MockAddressesObj{Obj: addr}
	return true, nil
}

// InsertAddressHook mocks inserting an Address.
func InsertAddressHook(ctx context.Context, key *meta.Key, obj *ga.Address, m *cloud.MockAddresses, options ...cloud.Option) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionGA, "addresses")
	return convertAndInsertAlphaAddress(key, obj, m.Objects, meta.VersionGA, projectID, m.X.(AddressAttributes))
}

// Verify InsertAddressHook implements cloud.MockAddresses.InsertHook.
var _ = cloud.MockAddresses{
	InsertHook: InsertAddressHook,
}

// InsertBetaAddressHook mocks inserting a BetaAddress.
func InsertBetaAddressHook(ctx context.Context, key *meta.Key, obj *beta.Address, m *cloud.MockBetaAddresses, options ...cloud.Option) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionBeta, "addresses")
	return convertAndInsertAlphaAddress(key, obj, m.Objects, meta.VersionBeta, projectID, m.X.(AddressAttributes))
}

// Verify InsertBetaAddressHook implements MockBetaAddresses.InsertHook.
var _ = cloud.MockBetaAddresses{
	InsertHook: InsertBetaAddressHook,
}

// InsertAlphaAddressHook mocks inserting an Address. Addresses are expected to
// default to Premium tier if no NetworkTier is specified.
func InsertAlphaAddressHook(ctx context.Context, key *meta.Key, obj *alpha.Address, m *cloud.MockAlphaAddresses, options ...cloud.Option) (bool, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	projectID := m.ProjectRouter.ProjectID(ctx, meta.VersionBeta, "addresses")
	return convertAndInsertAlphaAddress(key, obj, m.Objects, meta.VersionAlpha, projectID, m.X.(AddressAttributes))
}

// Verify InsertAlphaAddressHook implements MockAlphaAddresses.InsertHook.
var _ = cloud.MockAlphaAddresses{
	InsertHook: InsertAlphaAddressHook,
}

// InstanceGroupAttributes maps from InstanceGroup key to a map of Instances
type InstanceGroupAttributes struct {
	InstanceMap map[meta.Key]map[string]*ga.InstanceWithNamedPorts
	Lock        *sync.Mutex
}

// AddInstances adds a list of Instances passed by InstanceReference
func (igAttrs *InstanceGroupAttributes) AddInstances(key *meta.Key, instanceRefs []*ga.InstanceReference) error {
	igAttrs.Lock.Lock()
	defer igAttrs.Lock.Unlock()

	instancesWithNamedPorts, ok := igAttrs.InstanceMap[*key]
	if !ok {
		instancesWithNamedPorts = make(map[string]*ga.InstanceWithNamedPorts)
	}

	for _, instance := range instanceRefs {
		iWithPort := &ga.InstanceWithNamedPorts{
			Instance: instance.Instance,
		}

		instancesWithNamedPorts[instance.Instance] = iWithPort
	}

	igAttrs.InstanceMap[*key] = instancesWithNamedPorts
	return nil
}

// RemoveInstances removes a list of Instances passed by InstanceReference
func (igAttrs *InstanceGroupAttributes) RemoveInstances(key *meta.Key, instanceRefs []*ga.InstanceReference) error {
	igAttrs.Lock.Lock()
	defer igAttrs.Lock.Unlock()

	instancesWithNamedPorts, ok := igAttrs.InstanceMap[*key]
	if !ok {
		instancesWithNamedPorts = make(map[string]*ga.InstanceWithNamedPorts)
	}

	for _, instanceToRemove := range instanceRefs {
		if _, ok := instancesWithNamedPorts[instanceToRemove.Instance]; ok {
			delete(instancesWithNamedPorts, instanceToRemove.Instance)
		} else {
			return &googleapi.Error{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("%s is not a member of %s", instanceToRemove.Instance, key.String()),
			}
		}
	}

	igAttrs.InstanceMap[*key] = instancesWithNamedPorts
	return nil
}

// List gets a list of InstanceWithNamedPorts
func (igAttrs *InstanceGroupAttributes) List(key *meta.Key) []*ga.InstanceWithNamedPorts {
	igAttrs.Lock.Lock()
	defer igAttrs.Lock.Unlock()

	instancesWithNamedPorts, ok := igAttrs.InstanceMap[*key]
	if !ok {
		instancesWithNamedPorts = make(map[string]*ga.InstanceWithNamedPorts)
	}

	var instanceList []*ga.InstanceWithNamedPorts
	for _, val := range instancesWithNamedPorts {
		instanceList = append(instanceList, val)
	}

	return instanceList
}

// AddInstancesHook mocks adding instances from an InstanceGroup
func AddInstancesHook(ctx context.Context, key *meta.Key, req *ga.InstanceGroupsAddInstancesRequest, m *cloud.MockInstanceGroups, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	var attrs InstanceGroupAttributes
	attrs = m.X.(InstanceGroupAttributes)
	attrs.AddInstances(key, req.Instances)
	m.X = attrs
	return nil
}

// Verify AddInstancesHook implements MockInstanceGroups.AddInstancesHook.
var _ = cloud.MockInstanceGroups{
	AddInstancesHook: AddInstancesHook,
}

// ListInstancesHook mocks listing instances from an InstanceGroup
func ListInstancesHook(ctx context.Context, key *meta.Key, req *ga.InstanceGroupsListInstancesRequest, filter *filter.F, m *cloud.MockInstanceGroups, options ...cloud.Option) ([]*ga.InstanceWithNamedPorts, error) {
	_, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var attrs InstanceGroupAttributes
	attrs = m.X.(InstanceGroupAttributes)
	instances := attrs.List(key)

	return instances, nil
}

// Verify ListInstancesHook implements MockInstanceGroups.ListInstancesHook.
var _ = cloud.MockInstanceGroups{
	ListInstancesHook: ListInstancesHook,
}

// RemoveInstancesHook mocks removing instances from an InstanceGroup
func RemoveInstancesHook(ctx context.Context, key *meta.Key, req *ga.InstanceGroupsRemoveInstancesRequest, m *cloud.MockInstanceGroups, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	var attrs InstanceGroupAttributes
	attrs = m.X.(InstanceGroupAttributes)
	attrs.RemoveInstances(key, req.Instances)
	m.X = attrs
	return nil
}

// Verify RemoveInstancesHook implements MockInstanceGroups.RemoveInstancesHook.
var _ = cloud.MockInstanceGroups{
	RemoveInstancesHook: RemoveInstancesHook,
}

// UpdateFirewallHook defines the hook for updating a Firewall. It replaces the
// object with the same key in the mock with the updated object.
func UpdateFirewallHook(ctx context.Context, key *meta.Key, obj *ga.Firewall, m *cloud.MockFirewalls, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "firewalls")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "firewalls", key)

	m.Objects[*key] = &cloud.MockFirewallsObj{Obj: obj}
	return nil
}

// Verify UpdateFirewallHook implements MockFirewalls.UpdateHook.
var _ = cloud.MockFirewalls{
	UpdateHook: UpdateFirewallHook,
}

// UpdateAlphaFirewallHook defines the hook for updating an alpha Firewall. It replaces the
// object with the same key in the mock with the updated object.
func UpdateAlphaFirewallHook(ctx context.Context, key *meta.Key, obj *alpha.Firewall, m *cloud.MockAlphaFirewalls, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "firewalls")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "firewalls", key)

	m.Objects[*key] = &cloud.MockFirewallsObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaFirewallHook implements MockAlphaFirewalls.UpdateHook.
var _ = cloud.MockAlphaFirewalls{
	UpdateHook: UpdateAlphaFirewallHook,
}

// UpdateBetaFirewallHook defines the hook for updating a beta Firewall. It replaces the
// object with the same key in the mock with the updated object.
func UpdateBetaFirewallHook(ctx context.Context, key *meta.Key, obj *beta.Firewall, m *cloud.MockBetaFirewalls, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "firewalls")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionBeta, projectID, "firewalls", key)

	m.Objects[*key] = &cloud.MockFirewallsObj{Obj: obj}
	return nil
}

// Verify UpdateBetaFirewallHook implements MockBetaFirewalls.UpdateHook.
var _ = cloud.MockBetaFirewalls{
	UpdateHook: UpdateBetaFirewallHook,
}

// UpdateHealthCheckHook defines the hook for updating a HealthCheck. It
// replaces the object with the same key in the mock with the updated object.
func UpdateHealthCheckHook(ctx context.Context, key *meta.Key, obj *ga.HealthCheck, m *cloud.MockHealthChecks, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "healthChecks")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockHealthChecksObj{Obj: obj}
	return nil
}

// Verify UpdateHealthCheckHook implements MockHealthChecks.UpdateHook.
var _ = cloud.MockHealthChecks{
	UpdateHook: UpdateHealthCheckHook,
}

// UpdateAlphaHealthCheckHook defines the hook for updating an alpha HealthCheck.
// It replaces the object with the same key in the mock with the updated object.
func UpdateAlphaHealthCheckHook(ctx context.Context, key *meta.Key, obj *alpha.HealthCheck, m *cloud.MockAlphaHealthChecks, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "healthChecks")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockHealthChecksObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaHealthCheckHook implements MockAlphaHealthChecks.UpdateHook.
var _ = cloud.MockAlphaHealthChecks{
	UpdateHook: UpdateAlphaHealthCheckHook,
}

// UpdateAlphaRegionHealthCheckHook defines the hook for updating an alpha HealthCheck.
// It replaces the object with the same key in the mock with the updated object.
func UpdateAlphaRegionHealthCheckHook(ctx context.Context, key *meta.Key, obj *alpha.HealthCheck, m *cloud.MockAlphaRegionHealthChecks, options ...cloud.Option) error {
	if _, err := m.Get(ctx, key); err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "healthChecks")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockRegionHealthChecksObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaRegionHealthCheckHook implements MockAlphaRegionHealthChecks.UpdateHook.
var _ = cloud.MockAlphaRegionHealthChecks{
	UpdateHook: UpdateAlphaRegionHealthCheckHook,
}

// UpdateBetaHealthCheckHook defines the hook for updating a HealthCheck. It
// replaces the object with the same key in the mock with the updated object.
func UpdateBetaHealthCheckHook(ctx context.Context, key *meta.Key, obj *beta.HealthCheck, m *cloud.MockBetaHealthChecks, options ...cloud.Option) error {
	if _, err := m.Get(ctx, key); err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "healthChecks")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionBeta, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockHealthChecksObj{Obj: obj}
	return nil
}

// Verify UpdateBetaHealthCheckHook implements MockBetaHealthChecks.UpdateHook.
var _ = cloud.MockBetaHealthChecks{
	UpdateHook: UpdateBetaHealthCheckHook,
}

// UpdateBetaRegionHealthCheckHook defines the hook for updating a HealthCheck. It
// replaces the object with the same key in the mock with the updated object.
func UpdateBetaRegionHealthCheckHook(ctx context.Context, key *meta.Key, obj *beta.HealthCheck, m *cloud.MockBetaRegionHealthChecks, options ...cloud.Option) error {
	if _, err := m.Get(ctx, key); err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "healthChecks")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionBeta, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockRegionHealthChecksObj{Obj: obj}
	return nil
}

// Verify UpdateBetaRegionHealthCheckHook implements MockBetaRegionHealthChecks.UpdateHook.
var _ = cloud.MockBetaRegionHealthChecks{
	UpdateHook: UpdateBetaRegionHealthCheckHook,
}

// UpdateRegionHealthCheckHook defines the hook for updating a HealthCheck. It
// replaces the object with the same key in the mock with the updated object.
func UpdateRegionHealthCheckHook(ctx context.Context, key *meta.Key, obj *ga.HealthCheck, m *cloud.MockRegionHealthChecks, options ...cloud.Option) error {
	if _, err := m.Get(ctx, key); err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "healthChecks")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "healthChecks", key)

	m.Objects[*key] = &cloud.MockRegionHealthChecksObj{Obj: obj}
	return nil
}

// Verify UpdateRegionHealthCheckHook implements MockRegionHealthChecks.UpdateHook.
var _ = cloud.MockRegionHealthChecks{
	UpdateHook: UpdateRegionHealthCheckHook,
}

// UpdateRegionBackendServiceHook defines the hook for updating a Region
// BackendsService. It replaces the object with the same key in the mock with
// the updated object.
func UpdateRegionBackendServiceHook(ctx context.Context, key *meta.Key, obj *ga.BackendService, m *cloud.MockRegionBackendServices, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "backendServices")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockRegionBackendServicesObj{Obj: obj}
	return nil
}

// Verify UpdateRegionBackendServiceHook implements MockRegionBackendServices.UpdateHook.
var _ = cloud.MockRegionBackendServices{
	UpdateHook: UpdateRegionBackendServiceHook,
}

// UpdateAlphaRegionBackendServiceHook defines the hook for updating a Region
// BackendsService. It replaces the object with the same key in the mock with
// the updated object.
func UpdateAlphaRegionBackendServiceHook(ctx context.Context, key *meta.Key, obj *alpha.BackendService, m *cloud.MockAlphaRegionBackendServices, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "backendServices")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockRegionBackendServicesObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaRegionBackendServiceHook implements MockAlphaRegionBackendServices.UpdateHook.
var _ = cloud.MockAlphaRegionBackendServices{
	UpdateHook: UpdateAlphaRegionBackendServiceHook,
}

// UpdateBetaRegionBackendServiceHook defines the hook for updating a Region
// BackendsService. It replaces the object with the same key in the mock with
// the updated object.
func UpdateBetaRegionBackendServiceHook(ctx context.Context, key *meta.Key, obj *beta.BackendService, m *cloud.MockBetaRegionBackendServices, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "backendServices")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockRegionBackendServicesObj{Obj: obj}
	return nil
}

// Verify UpdateBetaRegionBackendServiceHook implements MockBetaRegionBackendServices.UpdateHook.
var _ = cloud.MockBetaRegionBackendServices{
	UpdateHook: UpdateBetaRegionBackendServiceHook,
}

// UpdateBackendServiceHook defines the hook for updating a BackendService.
// It replaces the object with the same key in the mock with the updated object.
func UpdateBackendServiceHook(ctx context.Context, key *meta.Key, obj *ga.BackendService, m *cloud.MockBackendServices, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "backendServices")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockBackendServicesObj{Obj: obj}
	return nil
}

// Verify UpdateRegionBackendServiceHook implements MockRegionBackendServices.UpdateHook.
var _ = cloud.MockRegionBackendServices{
	UpdateHook: UpdateRegionBackendServiceHook,
}

// UpdateAlphaBackendServiceHook defines the hook for updating an alpha BackendService.
// It replaces the object with the same key in the mock with the updated object.
func UpdateAlphaBackendServiceHook(ctx context.Context, key *meta.Key, obj *alpha.BackendService, m *cloud.MockAlphaBackendServices, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "backendServices")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockBackendServicesObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaBackendServiceHook implements MockAlphaBackendServices.UpdateHook.
var _ = cloud.MockAlphaBackendServices{
	UpdateHook: UpdateAlphaBackendServiceHook,
}

// UpdateBetaBackendServiceHook defines the hook for updating an beta BackendService.
// It replaces the object with the same key in the mock with the updated object.
func UpdateBetaBackendServiceHook(ctx context.Context, key *meta.Key, obj *beta.BackendService, m *cloud.MockBetaBackendServices, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "backendServices")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionBeta, projectID, "backendServices", key)

	m.Objects[*key] = &cloud.MockBackendServicesObj{Obj: obj}
	return nil
}

// Verify UpdateBetaBackendServiceHook implements MockBetaBackendServices.UpdateHook.
var _ = cloud.MockBetaBackendServices{
	UpdateHook: UpdateBetaBackendServiceHook,
}

// UpdateURLMapHook defines the hook for updating a UrlMap.
// It replaces the object with the same key in the mock with the updated object.
func UpdateURLMapHook(ctx context.Context, key *meta.Key, obj *ga.UrlMap, m *cloud.MockUrlMaps, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "urlMaps")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "urlMaps", key)

	m.Objects[*key] = &cloud.MockUrlMapsObj{Obj: obj}
	return nil
}

// Verify UpdateURLMapHook implements MockUrlMaps.UpdateHook.
var _ = cloud.MockUrlMaps{
	UpdateHook: UpdateURLMapHook,
}

// UpdateAlphaURLMapHook defines the hook for updating an alpha UrlMap.
// It replaces the object with the same key in the mock with the updated object.
func UpdateAlphaURLMapHook(ctx context.Context, key *meta.Key, obj *alpha.UrlMap, m *cloud.MockAlphaUrlMaps, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "urlMaps")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "urlMaps", key)

	m.Objects[*key] = &cloud.MockUrlMapsObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaURLMapHook implements MockAlphaUrlMaps.UpdateHook.
var _ = cloud.MockAlphaUrlMaps{
	UpdateHook: UpdateAlphaURLMapHook,
}

// UpdateBetaURLMapHook defines the hook for updating a beta UrlMap.
// It replaces the object with the same key in the mock with the updated object.
func UpdateBetaURLMapHook(ctx context.Context, key *meta.Key, obj *beta.UrlMap, m *cloud.MockBetaUrlMaps, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "urlMaps")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionBeta, projectID, "urlMaps", key)

	m.Objects[*key] = &cloud.MockUrlMapsObj{Obj: obj}
	return nil
}

// Verify UpdateBetaURLMapHook implements MockBetaUrlMaps.UpdateHook.
var _ = cloud.MockBetaUrlMaps{
	UpdateHook: UpdateBetaURLMapHook,
}

// UpdateAlphaRegionURLMapHook defines the hook for updating an alpha UrlMap.
// It replaces the object with the same key in the mock with the updated object.
func UpdateAlphaRegionURLMapHook(ctx context.Context, key *meta.Key, obj *alpha.UrlMap, m *cloud.MockAlphaRegionUrlMaps, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "alpha", "urlMaps")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionAlpha, projectID, "urlMaps", key)

	m.Objects[*key] = &cloud.MockRegionUrlMapsObj{Obj: obj}
	return nil
}

// Verify UpdateAlphaRegionURLMapHook implements MockAlphaRegionUrlMaps.UpdateHook.
var _ = cloud.MockAlphaRegionUrlMaps{
	UpdateHook: UpdateAlphaRegionURLMapHook,
}

// UpdateBetaRegionURLMapHook defines the hook for updating an alpha UrlMap.
// It replaces the object with the same key in the mock with the updated object.
func UpdateBetaRegionURLMapHook(ctx context.Context, key *meta.Key, obj *beta.UrlMap, m *cloud.MockBetaRegionUrlMaps, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "beta", "urlMaps")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionBeta, projectID, "urlMaps", key)

	m.Objects[*key] = &cloud.MockRegionUrlMapsObj{Obj: obj}
	return nil
}

// Verify UpdateBetaRegionURLMapHook implements MockBetaRegionUrlMaps.UpdateHook.
var _ = cloud.MockBetaRegionUrlMaps{
	UpdateHook: UpdateBetaRegionURLMapHook,
}

// UpdateRegionURLMapHook defines the hook for updating a GA Regional UrlMap.
// It replaces the object with the same key in the mock with the updated object.
func UpdateRegionURLMapHook(ctx context.Context, key *meta.Key, obj *ga.UrlMap, m *cloud.MockRegionUrlMaps, options ...cloud.Option) error {
	_, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	obj.Name = key.Name
	projectID := m.ProjectRouter.ProjectID(ctx, "ga", "urlMaps")
	obj.SelfLink = cloud.SelfLinkWithGroup(meta.APIGroupCompute, meta.VersionGA, projectID, "urlMaps", key)

	m.Objects[*key] = &cloud.MockRegionUrlMapsObj{Obj: obj}
	return nil
}

// Verify UpdateRegionURLMapHook implements MockRegionUrlMaps.UpdateHook.
var _ = cloud.MockRegionUrlMaps{
	UpdateHook: UpdateRegionURLMapHook,
}

// SetTargetGlobalForwardingRuleHook defines the hook for setting the target proxy for a GlobalForwardingRule.
func SetTargetGlobalForwardingRuleHook(ctx context.Context, key *meta.Key, obj *ga.TargetReference, m *cloud.MockGlobalForwardingRules, options ...cloud.Option) error {
	fw, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	fw.Target = obj.Target
	return nil
}

// Verify SetTargetGlobalForwardingRuleHook implements MockGlobalForwardingRules.SetTargetHook.
var _ = cloud.MockGlobalForwardingRules{
	SetTargetHook: SetTargetGlobalForwardingRuleHook,
}

// SetTargetForwardingRuleHook defines the hook for setting the target proxy for a ForwardingRule.
func SetTargetForwardingRuleHook(ctx context.Context, key *meta.Key, obj *ga.TargetReference, m *cloud.MockForwardingRules, options ...cloud.Option) error {
	fw, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	fw.Target = obj.Target
	return nil
}

// Verify SetTargetForwardingRuleHook implements MockForwardingRules.SetTargetHook.
var _ = cloud.MockForwardingRules{
	SetTargetHook: SetTargetForwardingRuleHook,
}

// SetTargetAlphaForwardingRuleHook defines the hook for setting the target proxy for an Alpha ForwardingRule.
func SetTargetAlphaForwardingRuleHook(ctx context.Context, key *meta.Key, obj *alpha.TargetReference, m *cloud.MockAlphaForwardingRules, options ...cloud.Option) error {
	fw, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	fw.Target = obj.Target
	return nil
}

// Verify SetTargetAlphaForwardingRuleHook implements MockAlphaForwardingRules.SetTargetHook.
var _ = cloud.MockAlphaForwardingRules{
	SetTargetHook: SetTargetAlphaForwardingRuleHook,
}

// SetTargetBetaForwardingRuleHook defines the hook for setting the target proxy for an Alpha ForwardingRule.
func SetTargetBetaForwardingRuleHook(ctx context.Context, key *meta.Key, obj *beta.TargetReference, m *cloud.MockBetaForwardingRules, options ...cloud.Option) error {
	fw, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	fw.Target = obj.Target
	return nil
}

// Verify SetTargetBetaForwardingRuleHook implements MockBetaForwardingRules.SetTargetHook.
var _ = cloud.MockBetaForwardingRules{
	SetTargetHook: SetTargetBetaForwardingRuleHook,
}

// SetTargetAlphaGlobalForwardingRuleHook defines the hook for setting the target proxy for an alpha GlobalForwardingRule.
func SetTargetAlphaGlobalForwardingRuleHook(ctx context.Context, key *meta.Key, ref *alpha.TargetReference, m *cloud.MockAlphaGlobalForwardingRules, options ...cloud.Option) error {
	fw, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	fw.Target = ref.Target
	return nil
}

// Verify SetTargetAlphaGlobalForwardingRuleHook implements MockAlphaGlobalForwardingRules.SetTargetHook.
var _ = cloud.MockAlphaGlobalForwardingRules{
	SetTargetHook: SetTargetAlphaGlobalForwardingRuleHook,
}

// SetTargetBetaGlobalForwardingRuleHook defines the hook for setting the target proxy for a beta GlobalForwardingRule.
func SetTargetBetaGlobalForwardingRuleHook(ctx context.Context, key *meta.Key, obj *beta.TargetReference, m *cloud.MockBetaGlobalForwardingRules, options ...cloud.Option) error {
	fw, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	fw.Target = obj.Target
	return nil
}

// Verify SetTargetBetaGlobalForwardingRuleHook implements MockBetaGlobalForwardingRules.SetTargetHook.
var _ = cloud.MockBetaGlobalForwardingRules{
	SetTargetHook: SetTargetBetaGlobalForwardingRuleHook,
}

// SetURLMapTargetHTTPProxyHook defines the hook for setting the url map for a TargetHttpProxy.
func SetURLMapTargetHTTPProxyHook(ctx context.Context, key *meta.Key, ref *ga.UrlMapReference, m *cloud.MockTargetHttpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapTargetHTTPProxyHook implements MockTargetHttpProxies.SetUrlMapHook.
var _ = cloud.MockTargetHttpProxies{
	SetUrlMapHook: SetURLMapTargetHTTPProxyHook,
}

// SetURLMapTargetHTTPSProxyHook defines the hook for setting the url map for a TargetHttpsProxy.
func SetURLMapTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *ga.UrlMapReference, m *cloud.MockTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapTargetHTTPSProxyHook implements MockTargetHttpsProxies.SetUrlMapHook.
var _ = cloud.MockTargetHttpsProxies{
	SetUrlMapHook: SetURLMapTargetHTTPSProxyHook,
}

// SetURLMapAlphaRegionTargetHTTPSProxyHook defines the hook for setting the url map for a TargetHttpsProxy.
func SetURLMapAlphaRegionTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *alpha.UrlMapReference, m *cloud.MockAlphaRegionTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapAlphaRegionTargetHTTPSProxyHook implements MockAlphaRegionTargetHttpsProxies.SetUrlMapHook.
var _ = cloud.MockAlphaRegionTargetHttpsProxies{
	SetUrlMapHook: SetURLMapAlphaRegionTargetHTTPSProxyHook,
}

// SetURLMapBetaRegionTargetHTTPSProxyHook defines the hook for setting the url map for a TargetHttpsProxy.
func SetURLMapBetaRegionTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *beta.UrlMapReference, m *cloud.MockBetaRegionTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapBetaRegionTargetHTTPSProxyHook implements MockBetaRegionTargetHttpsProxies.SetUrlMapHook.
var _ = cloud.MockBetaRegionTargetHttpsProxies{
	SetUrlMapHook: SetURLMapBetaRegionTargetHTTPSProxyHook,
}

// SetURLMapRegionTargetHTTPSProxyHook defines the hook for setting the url map for a TargetHttpsProxy.
func SetURLMapRegionTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *ga.UrlMapReference, m *cloud.MockRegionTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapRegionTargetHTTPSProxyHook implements MockRegionTargetHttpsProxies.SetUrlMapHook.
var _ = cloud.MockRegionTargetHttpsProxies{
	SetUrlMapHook: SetURLMapRegionTargetHTTPSProxyHook,
}

// SetURLMapAlphaTargetHTTPProxyHook defines the hook for setting the url map for a TargetHttpProxy.
func SetURLMapAlphaTargetHTTPProxyHook(ctx context.Context, key *meta.Key, ref *alpha.UrlMapReference, m *cloud.MockAlphaTargetHttpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapAlphaTargetHTTPProxyHook implements MockAlphaTargetHttpProxies.SetUrlMapHook.
var _ = cloud.MockAlphaTargetHttpProxies{
	SetUrlMapHook: SetURLMapAlphaTargetHTTPProxyHook,
}

// SetURLMapBetaTargetHTTPProxyHook defines the hook for setting the url map for a beta TargetHttpProxy.
func SetURLMapBetaTargetHTTPProxyHook(ctx context.Context, key *meta.Key, ref *beta.UrlMapReference, m *cloud.MockBetaTargetHttpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapBetaTargetHTTPProxyHook implements MockBetaTargetHttpProxies.SetUrlMapHook.
var _ = cloud.MockBetaTargetHttpProxies{
	SetUrlMapHook: SetURLMapBetaTargetHTTPProxyHook,
}

// SetURLMapBetaTargetHTTPSProxyHook defines the hook for setting the url map for a beta TargetHttpsProxy.
func SetURLMapBetaTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *beta.UrlMapReference, m *cloud.MockBetaTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapBetaTargetHTTPSProxyHook implements MockBetaTargetHttpsProxies.SetUrlMapHook.
var _ = cloud.MockBetaTargetHttpsProxies{
	SetUrlMapHook: SetURLMapBetaTargetHTTPSProxyHook,
}

// SetURLMapAlphaRegionTargetHTTPProxyHook defines the hook for setting the url map for a TargetHttpProxy.
func SetURLMapAlphaRegionTargetHTTPProxyHook(ctx context.Context, key *meta.Key, ref *alpha.UrlMapReference, m *cloud.MockAlphaRegionTargetHttpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapAlphaRegionTargetHTTPProxyHook implements MockAlphaRegionTargetHttpProxies.SetUrlMapHook.
var _ = cloud.MockAlphaRegionTargetHttpProxies{
	SetUrlMapHook: SetURLMapAlphaRegionTargetHTTPProxyHook,
}

// SetURLMapBetaRegionTargetHTTPProxyHook defines the hook for setting the url map for a TargetHttpProxy.
func SetURLMapBetaRegionTargetHTTPProxyHook(ctx context.Context, key *meta.Key, ref *beta.UrlMapReference, m *cloud.MockBetaRegionTargetHttpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapBetaRegionTargetHTTPProxyHook implements MockBetaRegionTargetHttpProxies.SetUrlMapHook.
var _ = cloud.MockBetaRegionTargetHttpProxies{
	SetUrlMapHook: SetURLMapBetaRegionTargetHTTPProxyHook,
}

// SetURLMapRegionTargetHTTPProxyHook defines the hook for setting the url map for a TargetHttpProxy.
func SetURLMapRegionTargetHTTPProxyHook(ctx context.Context, key *meta.Key, ref *ga.UrlMapReference, m *cloud.MockRegionTargetHttpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.UrlMap = ref.UrlMap
	return nil
}

// Verify SetURLMapRegionTargetHTTPProxyHook implements MockRegionTargetHttpProxies.SetUrlMapHook.
var _ = cloud.MockRegionTargetHttpProxies{
	SetUrlMapHook: SetURLMapRegionTargetHTTPProxyHook,
}

// SetBackendServiceAlphaTargetTCPProxyHook defines the hook for setting the backend service for an alpha TargetTcpProxy.
func SetBackendServiceAlphaTargetTCPProxyHook(ctx context.Context, key *meta.Key, ref *alpha.TargetTcpProxiesSetBackendServiceRequest, m *cloud.MockAlphaTargetTcpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.Service = ref.Service
	return nil
}

// Verify SetBackendServiceAlphaTargetTCPProxyHook implements MockAlphaTargetTcpProxies.SetBackendServiceHook.
var _ = cloud.MockAlphaTargetTcpProxies{
	SetBackendServiceHook: SetBackendServiceAlphaTargetTCPProxyHook,
}

// SetBackendServiceBetaTargetTCPProxyHook defines the hook for setting the backend service for a beta TargetTcpProxy.
func SetBackendServiceBetaTargetTCPProxyHook(ctx context.Context, key *meta.Key, ref *beta.TargetTcpProxiesSetBackendServiceRequest, m *cloud.MockBetaTargetTcpProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.Service = ref.Service
	return nil
}

// Verify SetBackendServiceBetaTargetTCPProxyHook implements MockBetaTargetTcpProxies.SetBackendServiceHook.
var _ = cloud.MockBetaTargetTcpProxies{
	SetBackendServiceHook: SetBackendServiceBetaTargetTCPProxyHook,
}

// SetSslCertificateTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslCertificateTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, req *ga.TargetHttpsProxiesSetSslCertificatesRequest, m *cloud.MockTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.SslCertificates = req.SslCertificates
	return nil
}

// Verify SetSslCertificateTargetHTTPSProxyHook implements MockTargetHttpsProxies.SetSslCertificatesHook.
var _ = cloud.MockTargetHttpsProxies{
	SetSslCertificatesHook: SetSslCertificateTargetHTTPSProxyHook,
}

// SetSslCertificateAlphaTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslCertificateAlphaTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, req *alpha.TargetHttpsProxiesSetSslCertificatesRequest, m *cloud.MockAlphaTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	tp.SslCertificates = req.SslCertificates
	return nil
}

// Verify SetSslCertificateAlphaTargetHTTPSProxyHook implements MockAlphaTargetHttpsProxies.SetSslCertificatesHook.
var _ = cloud.MockAlphaTargetHttpsProxies{
	SetSslCertificatesHook: SetSslCertificateAlphaTargetHTTPSProxyHook,
}

// SetSslCertificateBetaTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslCertificateBetaTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, req *beta.TargetHttpsProxiesSetSslCertificatesRequest, m *cloud.MockBetaTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	tp.SslCertificates = req.SslCertificates
	return nil
}

// Verify SetSslCertificateBetaTargetHTTPSProxyHook implements MockBetaTargetHttpsProxies.SetSslCertificatesHook.
var _ = cloud.MockBetaTargetHttpsProxies{
	SetSslCertificatesHook: SetSslCertificateBetaTargetHTTPSProxyHook,
}

// SetSslCertificateAlphaRegionTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslCertificateAlphaRegionTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, req *alpha.RegionTargetHttpsProxiesSetSslCertificatesRequest, m *cloud.MockAlphaRegionTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.SslCertificates = req.SslCertificates
	return nil
}

// Verify SetSslCertificateAlphaRegionTargetHTTPSProxyHook implements MockAlphaRegionTargetHttpsProxies.SetSslCertificatesHook.
var _ = cloud.MockAlphaRegionTargetHttpsProxies{
	SetSslCertificatesHook: SetSslCertificateAlphaRegionTargetHTTPSProxyHook,
}

// SetSslCertificateBetaRegionTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslCertificateBetaRegionTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, req *beta.RegionTargetHttpsProxiesSetSslCertificatesRequest, m *cloud.MockBetaRegionTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.SslCertificates = req.SslCertificates
	return nil
}

// Verify SetSslCertificateBetaRegionTargetHTTPSProxyHook implements MockBetaRegionTargetHttpsProxies.SetSslCertificatesHook.
var _ = cloud.MockBetaRegionTargetHttpsProxies{
	SetSslCertificatesHook: SetSslCertificateBetaRegionTargetHTTPSProxyHook,
}

// SetSslCertificateRegionTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslCertificateRegionTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, req *ga.RegionTargetHttpsProxiesSetSslCertificatesRequest, m *cloud.MockRegionTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.SslCertificates = req.SslCertificates
	return nil
}

// Verify SetSslCertificateRegionTargetHTTPSProxyHook implements MockRegionTargetHttpsProxies.SetSslCertificatesHook.
var _ = cloud.MockRegionTargetHttpsProxies{
	SetSslCertificatesHook: SetSslCertificateRegionTargetHTTPSProxyHook,
}

// SetSslPolicyTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslPolicyTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *ga.SslPolicyReference, m *cloud.MockTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}

	tp.SslPolicy = ref.SslPolicy
	return nil
}

// Verify SetSslPolicyTargetHTTPSProxyHook implements MockTargetHttpsProxies.SetSslPolicyHook.
var _ = cloud.MockTargetHttpsProxies{
	SetSslPolicyHook: SetSslPolicyTargetHTTPSProxyHook,
}

// SetSslPolicyAlphaTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslPolicyAlphaTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *alpha.SslPolicyReference, m *cloud.MockAlphaTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	tp.SslPolicy = ref.SslPolicy
	return nil
}

// Verify SetSslPolicyAlphaTargetHTTPSProxyHook implements MockAlphaTargetHttpsProxies.SetSslPolicyHook.
var _ = cloud.MockAlphaTargetHttpsProxies{
	SetSslPolicyHook: SetSslPolicyAlphaTargetHTTPSProxyHook,
}

// SetSslPolicyBetaTargetHTTPSProxyHook defines the hook for setting ssl certificates on a TargetHttpsProxy.
func SetSslPolicyBetaTargetHTTPSProxyHook(ctx context.Context, key *meta.Key, ref *beta.SslPolicyReference, m *cloud.MockBetaTargetHttpsProxies, options ...cloud.Option) error {
	tp, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	tp.SslPolicy = ref.SslPolicy
	return nil
}

// Verify SetSslPolicyBetaTargetHTTPSProxyHook implements MockBetaTargetHttpsProxies.SetSslPolicyHook.
var _ = cloud.MockBetaTargetHttpsProxies{
	SetSslPolicyHook: SetSslPolicyBetaTargetHTTPSProxyHook,
}

// InsertFirewallsUnauthorizedErrHook mocks firewall insertion. A forbidden error will be thrown as return.
func InsertFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, obj *ga.Firewall, m *cloud.MockFirewalls, options ...cloud.Option) (bool, error) {
	return true, &googleapi.Error{Code: http.StatusForbidden}
}

// Verify InsertFirewallsUnauthorizedErrHook implements MockFirewalls.InsertHook.
var _ = cloud.MockFirewalls{
	InsertHook: InsertFirewallsUnauthorizedErrHook,
}

// UpdateFirewallsUnauthorizedErrHook mocks firewall updating. A forbidden error will be thrown as return.
func UpdateFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, obj *ga.Firewall, m *cloud.MockFirewalls, options ...cloud.Option) error {
	return &googleapi.Error{Code: http.StatusForbidden}
}

// Verify UpdateFirewallsUnauthorizedErrHook implements MockFirewalls.UpdateHook.
var _ = cloud.MockFirewalls{
	UpdateHook: UpdateFirewallsUnauthorizedErrHook,
}

// DeleteFirewallsUnauthorizedErrHook mocks firewall deletion. A forbidden error will be thrown as return.
func DeleteFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, m *cloud.MockFirewalls, options ...cloud.Option) (bool, error) {
	return true, &googleapi.Error{Code: http.StatusForbidden}
}

// Verify DeleteFirewallsUnauthorizedErrHook implements MockFirewalls.DeleteHook.
var _ = cloud.MockFirewalls{
	DeleteHook: DeleteFirewallsUnauthorizedErrHook,
}

// GetFirewallsUnauthorizedErrHook mocks firewall information retrival. A forbidden error will be thrown as return.
func GetFirewallsUnauthorizedErrHook(ctx context.Context, key *meta.Key, m *cloud.MockFirewalls, options ...cloud.Option) (bool, *ga.Firewall, error) {
	return true, nil, &googleapi.Error{Code: http.StatusForbidden}
}

// Verify GetFirewallsUnauthorizedErrHook implements MockFirewalls.GetHook.
var _ = cloud.MockFirewalls{
	GetHook: GetFirewallsUnauthorizedErrHook,
}

// GetTargetPoolInternalErrHook mocks getting target pool. It returns a internal server error.
func GetTargetPoolInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockTargetPools, options ...cloud.Option) (bool, *ga.TargetPool, error) {
	return true, nil, InternalServerError
}

// Verify GetTargetPoolInternalErrHook implements MockTargetPools.GetHook.
var _ = cloud.MockTargetPools{
	GetHook: GetTargetPoolInternalErrHook,
}

// GetForwardingRulesInternalErrHook mocks getting forwarding rules and returns an internal server error.
func GetForwardingRulesInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockForwardingRules, options ...cloud.Option) (bool, *ga.ForwardingRule, error) {
	return true, nil, InternalServerError
}

// Verify GetForwardingRulesInternalErrHook implements MockForwardingRules.GetHook.
var _ = cloud.MockForwardingRules{
	GetHook: GetForwardingRulesInternalErrHook,
}

// GetAddressesInternalErrHook mocks getting network address and returns an internal server error.
func GetAddressesInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockAddresses, options ...cloud.Option) (bool, *ga.Address, error) {
	return true, nil, InternalServerError
}

// Verify GetAddressesInternalErrHook implements MockAddresses.GetHook.
var _ = cloud.MockAddresses{
	GetHook: GetAddressesInternalErrHook,
}

// GetHTTPHealthChecksInternalErrHook mocks getting http health check and returns an internal server error.
func GetHTTPHealthChecksInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHttpHealthChecks, options ...cloud.Option) (bool, *ga.HttpHealthCheck, error) {
	return true, nil, InternalServerError
}

// Verify GetHTTPHealthChecksInternalErrHook implements MockHttpHealthChecks.GetHook.
var _ = cloud.MockHttpHealthChecks{
	GetHook: GetHTTPHealthChecksInternalErrHook,
}

// InsertTargetPoolsInternalErrHook mocks getting target pool and returns an internal server error.
func InsertTargetPoolsInternalErrHook(ctx context.Context, key *meta.Key, obj *ga.TargetPool, m *cloud.MockTargetPools, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify InsertTargetPoolsInternalErrHook implements MockTargetPools.InsertHook.
var _ = cloud.MockTargetPools{
	InsertHook: InsertTargetPoolsInternalErrHook,
}

// InsertForwardingRulesInternalErrHook mocks getting forwarding rule and returns an internal server error.
func InsertForwardingRulesInternalErrHook(ctx context.Context, key *meta.Key, obj *ga.ForwardingRule, m *cloud.MockForwardingRules, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify InsertForwardingRulesInternalErrHook implements MockForwardingRules.InsertHook.
var _ = cloud.MockForwardingRules{
	InsertHook: InsertForwardingRulesInternalErrHook,
}

// DeleteAddressesNotFoundErrHook mocks deleting network address and returns a not found error.
func DeleteAddressesNotFoundErrHook(ctx context.Context, key *meta.Key, m *cloud.MockAddresses, options ...cloud.Option) (bool, error) {
	return true, &googleapi.Error{Code: http.StatusNotFound}
}

// Verify DeleteAddressesNotFoundErrHook implements MockAddresses.DeleteHook.
var _ = cloud.MockAddresses{
	DeleteHook: DeleteAddressesNotFoundErrHook,
}

// DeleteAddressesInternalErrHook mocks deleting address and returns an internal server error.
func DeleteAddressesInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockAddresses, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify DeleteAddressesInternalErrHook implements MockAddresses.DeleteHook.
var _ = cloud.MockAddresses{
	DeleteHook: DeleteAddressesInternalErrHook,
}

// InsertAlphaBackendServiceUnauthorizedErrHook mocks inserting an alpha BackendService and returns a forbidden error.
func InsertAlphaBackendServiceUnauthorizedErrHook(ctx context.Context, key *meta.Key, obj *alpha.BackendService, m *cloud.MockAlphaBackendServices, options ...cloud.Option) (bool, error) {
	return true, UnauthorizedErr
}

// Verify InsertAlphaBackendServiceUnauthorizedErrHook implements MockAlphaBackendServices.InsertHook.
var _ = cloud.MockAlphaBackendServices{
	InsertHook: InsertAlphaBackendServiceUnauthorizedErrHook,
}

// UpdateAlphaBackendServiceUnauthorizedErrHook mocks updating an alpha BackendService and returns a forbidden error.
func UpdateAlphaBackendServiceUnauthorizedErrHook(ctx context.Context, key *meta.Key, obj *alpha.BackendService, m *cloud.MockAlphaBackendServices, options ...cloud.Option) error {
	return UnauthorizedErr
}

// Verify UpdateAlphaBackendServiceUnauthorizedErrHook implements MockAlphaBackendServices.UpdateHook.
var _ = cloud.MockAlphaBackendServices{
	UpdateHook: UpdateAlphaBackendServiceUnauthorizedErrHook,
}

// GetRegionBackendServicesErrHook mocks getting region backend service and returns an internal server error.
func GetRegionBackendServicesErrHook(ctx context.Context, key *meta.Key, m *cloud.MockRegionBackendServices, options ...cloud.Option) (bool, *ga.BackendService, error) {
	return true, nil, InternalServerError
}

// Verify GetRegionBackendServicesErrHook implements MockRegionBackendServices.GetHook.
var _ = cloud.MockRegionBackendServices{
	GetHook: GetRegionBackendServicesErrHook,
}

// UpdateRegionBackendServicesErrHook mocks updating a reegion backend service and returns an internal server error.
func UpdateRegionBackendServicesErrHook(ctx context.Context, key *meta.Key, svc *ga.BackendService, m *cloud.MockRegionBackendServices, options ...cloud.Option) error {
	return InternalServerError
}

// Verify UpdateRegionBackendServicesErrHook implements MockRegionBackendServices.UpdateHook.
var _ = cloud.MockRegionBackendServices{
	UpdateHook: UpdateRegionBackendServicesErrHook,
}

// DeleteRegionBackendServicesErrHook mocks deleting region backend service and returns an internal server error.
func DeleteRegionBackendServicesErrHook(ctx context.Context, key *meta.Key, m *cloud.MockRegionBackendServices, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify DeleteRegionBackendServicesErrHook implements MockRegionBackendServices.DeleteHook.
var _ = cloud.MockRegionBackendServices{
	DeleteHook: DeleteRegionBackendServicesErrHook,
}

// DeleteRegionBackendServicesInUseErrHook mocks deleting region backend service and returns an InUseError.
func DeleteRegionBackendServicesInUseErrHook(ctx context.Context, key *meta.Key, m *cloud.MockRegionBackendServices, options ...cloud.Option) (bool, error) {
	return true, InUseError
}

// Verify DeleteRegionBackendServicesInUseErrHook implements MockRegionBackendServices.DeleteHook.
var _ = cloud.MockRegionBackendServices{
	DeleteHook: DeleteRegionBackendServicesInUseErrHook,
}

// GetInstanceGroupInternalErrHook mocks getting instance group and returns an internal server error.
func GetInstanceGroupInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockInstanceGroups, options ...cloud.Option) (bool, *ga.InstanceGroup, error) {
	return true, nil, InternalServerError
}

// Verify GetInstanceGroupInternalErrHook implements MockInstanceGroups.GetHook.
var _ = cloud.MockInstanceGroups{
	GetHook: GetInstanceGroupInternalErrHook,
}

// GetHealthChecksInternalErrHook mocks getting health check and returns an internal server erorr.
func GetHealthChecksInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHealthChecks, options ...cloud.Option) (bool, *ga.HealthCheck, error) {
	return true, nil, InternalServerError
}

// Verify GetHealthChecksInternalErrHook implements MockHealthChecks.GetHook.
var _ = cloud.MockHealthChecks{
	GetHook: GetHealthChecksInternalErrHook,
}

// DeleteHealthChecksInternalErrHook mocks deleting health check and returns an internal server error.
func DeleteHealthChecksInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHealthChecks, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify DeleteHealthChecksInternalErrHook implements MockHealthChecks.DeleteHook.
var _ = cloud.MockHealthChecks{
	DeleteHook: DeleteHealthChecksInternalErrHook,
}

// DeleteHealthChecksInuseErrHook mocks deleting health check and returns an in use error.
func DeleteHealthChecksInuseErrHook(ctx context.Context, key *meta.Key, m *cloud.MockHealthChecks, options ...cloud.Option) (bool, error) {
	return true, InUseError
}

// Verify DeleteHealthChecksInuseErrHook implements MockHealthChecks.DeleteHook.
var _ = cloud.MockHealthChecks{
	DeleteHook: DeleteHealthChecksInuseErrHook,
}

// DeleteForwardingRuleErrHook mocks deleting forwarding rule and returns an internal server error.
func DeleteForwardingRuleErrHook(ctx context.Context, key *meta.Key, m *cloud.MockForwardingRules, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify DeleteForwardingRuleErrHook implements MockForwardingRules.DeleteHook.
var _ = cloud.MockForwardingRules{
	DeleteHook: DeleteForwardingRuleErrHook,
}

// ListZonesInternalErrHook mocks listing zone and returns an internal server error.
func ListZonesInternalErrHook(ctx context.Context, fl *filter.F, m *cloud.MockZones, options ...cloud.Option) (bool, []*ga.Zone, error) {
	return true, []*ga.Zone{}, InternalServerError
}

// Verify ListZonesInternalErrHook implements MockZones.ListHook.
var _ = cloud.MockZones{
	ListHook: ListZonesInternalErrHook,
}

// DeleteInstanceGroupInternalErrHook mocks deleting instance group and returns an internal server error.
func DeleteInstanceGroupInternalErrHook(ctx context.Context, key *meta.Key, m *cloud.MockInstanceGroups, options ...cloud.Option) (bool, error) {
	return true, InternalServerError
}

// Verify DeleteInstanceGroupInternalErrHook implements MockInstanceGroups.DeleteHook.
var _ = cloud.MockInstanceGroups{
	DeleteHook: DeleteInstanceGroupInternalErrHook,
}
