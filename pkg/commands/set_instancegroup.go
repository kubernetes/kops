/*
Copyright 2020 The Kubernetes Authors.

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

package commands

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
)

type instanceGroupConfigSetter = func(*api.InstanceGroup, string) error

// InstanceGroupKeySetters is a map of keys to config setting functions for
// instance groups.
type InstanceGroupKeySetters map[string]instanceGroupConfigSetter

// PrettyPrintKeysWithCommas prints comma-separated keys.
func (ks *InstanceGroupKeySetters) PrettyPrintKeysWithCommas() string {
	keys := make([]string, 0, len(*ks))
	for k := range *ks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

// PrettyPrintKeysWithNewlines prints newline-separated keys.
func (ks *InstanceGroupKeySetters) PrettyPrintKeysWithNewlines() string {
	keys := make([]string, 0, len(*ks))
	for k := range *ks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "\n ")
}

// RunSetInstancegroup implements the set instancegroup command logic.
func RunSetInstancegroup(f *util.Factory, cmd *cobra.Command, out io.Writer, options *SetOptions) error {
	if !featureflag.SpecOverrideFlag.Enabled() {
		return fmt.Errorf("set instancegroup is currently feature gated; set `export KOPS_FEATURE_FLAGS=SpecOverrideFlag`")
	}

	if options.ClusterName == "" {
		return field.Required(field.NewPath("ClusterName"), "Cluster name is required")
	}
	if options.InstanceGroupName == "" {
		return field.Required(field.NewPath("InstanceGroupName"), "Instance Group name is required")
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(options.ClusterName)
	if err != nil {
		return err
	}

	// All instance groups are needed eventually for validation, so let's grab
	// them all and update the pointer to the one we are setting config for.
	instanceGroups, err := ReadAllInstanceGroups(clientset, cluster)
	if err != nil {
		return err
	}
	var instanceGroupToUpdate *api.InstanceGroup
	for _, instanceGroup := range instanceGroups {
		if instanceGroup.GetName() == options.InstanceGroupName {
			instanceGroupToUpdate = instanceGroup
		}
	}
	if instanceGroupToUpdate == nil {
		return fmt.Errorf("unable to find instance group with name %q", options.InstanceGroupName)
	}

	err = SetInstancegroupFields(options.Fields, instanceGroupToUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to set instance group")
	}

	err = UpdateInstanceGroup(clientset, cluster, instanceGroups, instanceGroupToUpdate)
	if err != nil {
		return err
	}

	return nil
}

// SetInstancegroupFields sets field values in the instance group.
func SetInstancegroupFields(fields []string, instanceGroup *api.InstanceGroup) error {
	validKeyToSetters := ValidInstanceGroupKeysToSetters()

	for _, field := range fields {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("unhandled field: %q", field)
		}

		setter, ok := validKeyToSetters[kv[0]]
		if !ok {
			return fmt.Errorf("unhandled field: %q; valid instancegroup keys are: %s", field, validKeyToSetters.PrettyPrintKeysWithCommas())
		}

		err := setter(instanceGroup, kv[1])
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidInstanceGroupKeysToSetters returns the valid keys and config setting
// logic for instance groups.
func ValidInstanceGroupKeysToSetters() InstanceGroupKeySetters {
	return InstanceGroupKeySetters{
		"spec.image": func(ig *api.InstanceGroup, v string) error {
			ig.Spec.Image = v
			return nil
		},
		"spec.machineType": func(ig *api.InstanceGroup, v string) error {
			ig.Spec.MachineType = v
			return nil
		},
		"spec.minSize": func(ig *api.InstanceGroup, v string) error {
			i64, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return fmt.Errorf("unknown int32 value: %q", v)
			}
			i32 := int32(i64)

			ig.Spec.MinSize = &i32
			return nil
		},
		"spec.maxSize": func(ig *api.InstanceGroup, v string) error {
			i64, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return fmt.Errorf("unknown int32 value: %q", v)
			}
			i32 := int32(i64)

			ig.Spec.MaxSize = &i32
			return nil
		},
		"spec.associatePublicIp": func(ig *api.InstanceGroup, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("unknown boolean value: %q", v)
			}

			ig.Spec.AssociatePublicIP = &b
			return nil
		},
	}
}
