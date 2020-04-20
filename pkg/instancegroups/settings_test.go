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

package instancegroups

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kops/pkg/apis/kops"
)

func TestSettings(t *testing.T) {
	for _, tc := range []struct {
		name            string
		defaultValue    interface{}
		nonDefaultValue interface{}
	}{
		{
			name:            "MaxUnavailable",
			defaultValue:    intstr.FromInt(1),
			nonDefaultValue: intstr.FromInt(2),
		},
		{
			name:            "MaxSurge",
			defaultValue:    intstr.FromInt(0),
			nonDefaultValue: intstr.FromInt(2),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defaultCluster := &kops.RollingUpdate{}
			setFieldValue(defaultCluster, tc.name, tc.defaultValue)

			nonDefaultCluster := &kops.RollingUpdate{}
			setFieldValue(nonDefaultCluster, tc.name, tc.nonDefaultValue)

			defaultGroup := &kops.RollingUpdate{}
			setFieldValue(defaultGroup, tc.name, tc.defaultValue)

			nonDefaultGroup := &kops.RollingUpdate{}
			setFieldValue(nonDefaultGroup, tc.name, tc.nonDefaultValue)

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, nil, "nil nil")
			assertResolvesValue(t, tc.name, tc.defaultValue, &kops.RollingUpdate{}, nil, "{nil} nil")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, nil, "{default} nil")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, nil, "{nonDefault} nil")

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, &kops.RollingUpdate{}, "nil {nil}")
			assertResolvesValue(t, tc.name, tc.defaultValue, &kops.RollingUpdate{}, &kops.RollingUpdate{}, "{nil} {nil}")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, &kops.RollingUpdate{}, "{default} {nil}")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, &kops.RollingUpdate{}, "{nonDefault} {nil}")

			assertResolvesValue(t, tc.name, tc.defaultValue, nil, defaultGroup, "nil {default}")
			assertResolvesValue(t, tc.name, tc.defaultValue, &kops.RollingUpdate{}, defaultGroup, "{nil} {default}")
			assertResolvesValue(t, tc.name, tc.defaultValue, defaultCluster, defaultGroup, "{default} {default}")
			assertResolvesValue(t, tc.name, tc.defaultValue, nonDefaultCluster, defaultGroup, "{nonDefault} {default}")

			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nil, nonDefaultGroup, "nil {nonDefault}")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, &kops.RollingUpdate{}, nonDefaultGroup, "{nil} {nonDefault}")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, defaultCluster, nonDefaultGroup, "{default} {nonDefault}")
			assertResolvesValue(t, tc.name, tc.nonDefaultValue, nonDefaultCluster, nonDefaultGroup, "{nonDefault} {nonDefault}")
		})
	}
}

func setFieldValue(aStruct interface{}, fieldName string, fieldValue interface{}) {
	field := reflect.ValueOf(aStruct).Elem().FieldByName(fieldName)
	value := reflect.New(field.Type().Elem())
	value.Elem().Set(reflect.ValueOf(fieldValue))
	field.Set(value)
}

func assertResolvesValue(t *testing.T, name string, expected interface{}, rollingUpdateDefault *kops.RollingUpdate, rollingUpdate *kops.RollingUpdate, msg interface{}) bool {
	cluster := kops.Cluster{
		Spec: kops.ClusterSpec{
			RollingUpdate: rollingUpdateDefault,
		},
	}
	instanceGroup := kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			RollingUpdate: rollingUpdate,
		},
	}
	rollingUpdateDefaultCopy := rollingUpdateDefault.DeepCopy()
	rollingUpdateCopy := rollingUpdate.DeepCopy()

	resolved := resolveSettings(&cluster, &instanceGroup, 1)
	value := reflect.ValueOf(resolved).FieldByName(name)

	assert.Equal(t, rollingUpdateDefault, cluster.Spec.RollingUpdate, "cluster not modified")
	assert.True(t, reflect.DeepEqual(rollingUpdateDefault, rollingUpdateDefaultCopy), "RollingUpdate not modified")
	assert.Equal(t, rollingUpdate, instanceGroup.Spec.RollingUpdate, "instancegroup not modified")
	assert.True(t, reflect.DeepEqual(rollingUpdate, rollingUpdateCopy), "RollingUpdate not modified")

	return assert.NotNil(t, value.Interface(), msg) &&
		assert.Equal(t, expected, value.Elem().Interface(), msg)
}

func TestMaxUnavailable(t *testing.T) {
	for _, tc := range []struct {
		numInstances int
		value        string
		expected     int32
	}{
		{
			numInstances: 1,
			value:        "0",
			expected:     0,
		},
		{
			numInstances: 1,
			value:        "0%",
			expected:     1,
		},
		{
			numInstances: 10,
			value:        "39%",
			expected:     3,
		},
		{
			numInstances: 10,
			value:        "100%",
			expected:     10,
		},
	} {
		t.Run(fmt.Sprintf("%s %d", tc.value, tc.numInstances), func(t *testing.T) {
			value := intstr.Parse(tc.value)
			rollingUpdate := kops.RollingUpdate{
				MaxUnavailable: &value,
			}
			instanceGroup := kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					RollingUpdate: &rollingUpdate,
				},
			}
			resolved := resolveSettings(&kops.Cluster{}, &instanceGroup, tc.numInstances)
			assert.Equal(t, intstr.Int, resolved.MaxUnavailable.Type)
			assert.Equal(t, tc.expected, resolved.MaxUnavailable.IntVal)
		})
	}
}

func TestMaxSurge(t *testing.T) {
	for _, tc := range []struct {
		numInstances int
		value        string
		expected     int32
	}{
		{
			numInstances: 1,
			value:        "0",
			expected:     0,
		},
		{
			numInstances: 1,
			value:        "0%",
			expected:     0,
		},
		{
			numInstances: 10,
			value:        "31%",
			expected:     4,
		},
		{
			numInstances: 10,
			value:        "100%",
			expected:     10,
		},
	} {
		t.Run(fmt.Sprintf("%s %d", tc.value, tc.numInstances), func(t *testing.T) {
			value := intstr.Parse(tc.value)
			rollingUpdate := kops.RollingUpdate{
				MaxSurge: &value,
			}
			instanceGroup := kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					RollingUpdate: &rollingUpdate,
				},
			}
			resolved := resolveSettings(&kops.Cluster{}, &instanceGroup, tc.numInstances)
			assert.Equal(t, intstr.Int, resolved.MaxSurge.Type)
			assert.Equal(t, tc.expected, resolved.MaxSurge.IntVal)
			if tc.expected == 0 {
				assert.Equal(t, int32(1), resolved.MaxUnavailable.IntVal, "MaxUnavailable default")
			} else {
				assert.Equal(t, int32(0), resolved.MaxUnavailable.IntVal, "MaxUnavailable default")
			}
		})
	}
}

func TestAWSDefault(t *testing.T) {
	resolved := resolveSettings(&kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: "aws",
		},
	}, &kops.InstanceGroup{}, 1000)
	assert.Equal(t, intstr.Int, resolved.MaxSurge.Type)
	assert.Equal(t, int32(1), resolved.MaxSurge.IntVal)
	assert.Equal(t, intstr.Int, resolved.MaxUnavailable.Type)
	assert.Equal(t, int32(0), resolved.MaxUnavailable.IntVal)
}
