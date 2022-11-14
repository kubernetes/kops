/*
Copyright 2021 The Kubernetes Authors.

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

package kops

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarmPoolSpec_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		spec     *WarmPoolSpec
		expected bool
	}{
		{
			name:     "nil",
			spec:     nil,
			expected: false,
		},
		{
			name:     "empty",
			spec:     &WarmPoolSpec{},
			expected: true,
		},
		{
			name: "1",
			spec: &WarmPoolSpec{
				MaxSize: int64ptr(1),
			},
			expected: true,
		},
		{
			name: "0",
			spec: &WarmPoolSpec{
				MaxSize: int64ptr(0),
			},
			expected: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if actual := tc.spec.IsEnabled(); actual != tc.expected {
				t.Errorf("IsEnabled() = %v, want %v", actual, tc.expected)
			}
		})
	}
}

func int64ptr(v int64) *int64 {
	return &v
}

func TestWarmPoolSpec_ResolveDefaults(t *testing.T) {
	for _, tc := range []struct {
		name            string
		defaultValue    interface{}
		nonDefaultValue interface{}
	}{
		{
			name:            "MinSize",
			defaultValue:    int64(0),
			nonDefaultValue: int64(1),
		},
		{
			name:            "MaxSize",
			defaultValue:    nil,
			nonDefaultValue: int64(1),
		},
		{
			name:            "EnableLifecycleHook",
			defaultValue:    false,
			nonDefaultValue: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defaultCluster := &WarmPoolSpec{}
			setFieldValue(defaultCluster, tc.name, tc.defaultValue)

			nonDefaultCluster := &WarmPoolSpec{}
			setFieldValue(nonDefaultCluster, tc.name, tc.nonDefaultValue)

			defaultGroup := &WarmPoolSpec{}
			setFieldValue(defaultGroup, tc.name, tc.defaultValue)

			nonDefaultGroup := &WarmPoolSpec{}
			setFieldValue(nonDefaultGroup, tc.name, tc.nonDefaultValue)

			expectedDefaultValue := tc.defaultValue
			if expectedDefaultValue == nil {
				expectedDefaultValue = reflect.Zero(reflect.ValueOf(*defaultGroup).FieldByName(tc.name).Type().Elem()).Interface()
			}

			assertResolvesValue(t, tc.name, expectedDefaultValue, nil, nil, InstanceGroupRoleNode, "nil nil node")
			assertResolvesValue(t, tc.name, tc.defaultValue, &WarmPoolSpec{}, nil, InstanceGroupRoleNode, "{nil} nil node")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, nil, InstanceGroupRoleNode, "{default} nil node")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, nil, InstanceGroupRoleNode, "{nonDefault} nil node")

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, &WarmPoolSpec{}, InstanceGroupRoleNode, "nil {nil} node")
			assertResolvesValue(t, tc.name, tc.defaultValue, &WarmPoolSpec{}, &WarmPoolSpec{}, InstanceGroupRoleNode, "{nil} {nil} node")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, &WarmPoolSpec{}, InstanceGroupRoleNode, "{default} {nil} node")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, &WarmPoolSpec{}, InstanceGroupRoleNode, "{nonDefault} {nil} node")

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, defaultGroup, InstanceGroupRoleNode, "nil {default} node")
			assertResolvesValue(t, tc.name, tc.defaultValue, &WarmPoolSpec{}, defaultGroup, InstanceGroupRoleNode, "{nil} {default} node")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, defaultGroup, InstanceGroupRoleNode, "{default} {default} node")
			if reflect.ValueOf(*defaultGroup).FieldByName(tc.name).Type().Kind() == reflect.Ptr && tc.defaultValue != nil {
				assertResolvesValue(t, tc.name, tc.defaultValue, nonDefaultCluster, defaultGroup, InstanceGroupRoleNode, "{nonDefault} {default} node")
			} else {
				assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, defaultGroup, InstanceGroupRoleNode, "{nonDefault} {default} node")
			}

			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nil, nonDefaultGroup, InstanceGroupRoleNode, "nil {nonDefault} node")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, &WarmPoolSpec{}, nonDefaultGroup, InstanceGroupRoleNode, "{nil} {nonDefault} node")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, defaultCluster, nonDefaultGroup, InstanceGroupRoleNode, "{default} {nonDefault} node")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, nonDefaultGroup, InstanceGroupRoleNode, "{nonDefault} {nonDefault} node")

			assertResolvesValue(t, tc.name, expectedDefaultValue, nil, nil, InstanceGroupRoleControlPlane, "nil nil master")
			assertResolvesValue(t, tc.name, expectedDefaultValue, &WarmPoolSpec{}, nil, InstanceGroupRoleControlPlane, "{nil} nil master")
			assertResolvesValue(t, tc.name, expectedDefaultValue, defaultCluster, nil, InstanceGroupRoleControlPlane, "{default} nil master")
			assertResolvesValue(t, tc.name, expectedDefaultValue, nonDefaultCluster, nil, InstanceGroupRoleControlPlane, "{nonDefault} nil master")

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, &WarmPoolSpec{}, InstanceGroupRoleControlPlane, "nil {nil} master")
			assertResolvesValue(t, tc.name, tc.defaultValue, &WarmPoolSpec{}, &WarmPoolSpec{}, InstanceGroupRoleControlPlane, "{nil} {nil} master")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, &WarmPoolSpec{}, InstanceGroupRoleControlPlane, "{default} {nil} master")
			assertResolvesValue(t, tc.name, tc.defaultValue, nonDefaultCluster, &WarmPoolSpec{}, InstanceGroupRoleControlPlane, "{nonDefault} {nil} master")

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, defaultGroup, InstanceGroupRoleControlPlane, "nil {default} master")
			assertResolvesValue(t, tc.name, tc.defaultValue, &WarmPoolSpec{}, defaultGroup, InstanceGroupRoleControlPlane, "{nil} {default} master")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, defaultGroup, InstanceGroupRoleControlPlane, "{default} {default} master")
			assertResolvesValue(t, tc.name, tc.defaultValue, nonDefaultCluster, defaultGroup, InstanceGroupRoleControlPlane, "{nonDefault} {default} master")

			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nil, nonDefaultGroup, InstanceGroupRoleControlPlane, "nil {nonDefault} master")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, &WarmPoolSpec{}, nonDefaultGroup, InstanceGroupRoleControlPlane, "{nil} {nonDefault} master")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, defaultCluster, nonDefaultGroup, InstanceGroupRoleControlPlane, "{default} {nonDefault} master")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, nonDefaultGroup, InstanceGroupRoleControlPlane, "{nonDefault} {nonDefault} master")
		})
	}
}

func setFieldValue(aStruct interface{}, fieldName string, fieldValue interface{}) {
	field := reflect.ValueOf(aStruct).Elem().FieldByName(fieldName)
	fieldType := field.Type()
	if fieldType.Kind() == reflect.Ptr {
		if fieldValue != nil {
			value := reflect.New(fieldType.Elem())
			value.Elem().Set(reflect.ValueOf(fieldValue))
			field.Set(value)
		}
	} else {
		field.Set(reflect.ValueOf(fieldValue))
	}
}

func assertResolvesValue(t *testing.T, name string, expected interface{}, warmPoolSpecDefault *WarmPoolSpec, ig *WarmPoolSpec, role InstanceGroupRole, msg interface{}) bool {
	cluster := Cluster{
		Spec: ClusterSpec{
			WarmPool: warmPoolSpecDefault,
		},
	}
	instanceGroup := InstanceGroup{
		Spec: InstanceGroupSpec{
			Role:     role,
			WarmPool: ig,
		},
	}
	warmPoolSpecDefaultCopy := warmPoolSpecDefault.DeepCopy()
	warmPoolSpecCopy := ig.DeepCopy()

	resolved := cluster.Spec.WarmPool.ResolveDefaults(&instanceGroup)
	value := reflect.ValueOf(*resolved).FieldByName(name)

	assert.Equal(t, warmPoolSpecDefault, cluster.Spec.WarmPool, "cluster not modified")
	assert.True(t, reflect.DeepEqual(warmPoolSpecDefault, warmPoolSpecDefaultCopy), "WarmPoolSpec not modified")
	assert.Equal(t, ig, instanceGroup.Spec.WarmPool, "instancegroup not modified")
	assert.True(t, reflect.DeepEqual(ig, warmPoolSpecCopy), "WarmPoolSpec not modified")

	if expected == nil {
		return assert.Nil(t, value.Interface(), msg)
	} else if value.Type().Kind() == reflect.Ptr {
		return assert.NotNil(t, value.Interface(), msg) &&
			assert.Equal(t, expected, value.Elem().Interface(), msg)
	} else {
		return assert.Equal(t, expected, value.Interface(), msg)
	}
}
