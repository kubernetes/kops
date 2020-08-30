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

package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/util/pkg/reflectutils"
)

type SetClusterOptions struct {
	Fields      []string
	ClusterName string
}

// RunSetCluster implements the set cluster command logic
func RunSetCluster(ctx context.Context, f *util.Factory, cmd *cobra.Command, out io.Writer, options *SetClusterOptions) error {
	if !featureflag.SpecOverrideFlag.Enabled() {
		return fmt.Errorf("set cluster command is current feature gated; set `export KOPS_FEATURE_FLAGS=SpecOverrideFlag`")
	}

	if options.ClusterName == "" {
		return field.Required(field.NewPath("clusterName"), "Cluster name is required")
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	instanceGroups, err := ReadAllInstanceGroups(ctx, clientset, cluster)
	if err != nil {
		return err
	}

	if err := SetClusterFields(options.Fields, cluster, instanceGroups); err != nil {
		return err
	}

	if err := UpdateCluster(ctx, clientset, cluster, instanceGroups); err != nil {
		return err
	}

	return nil
}

// SetClusterFields sets field values in the cluster
func SetClusterFields(fields []string, cluster *api.Cluster, instanceGroups []*api.InstanceGroup) error {
	for _, field := range fields {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("unhandled field: %q", field)
		}

		key := kv[0]
		key = strings.TrimPrefix(key, "cluster.")

		if err := reflectutils.SetString(cluster, key, kv[1]); err != nil {
			return fmt.Errorf("failed to set %s=%s: %v", kv[0], kv[1], err)
		}
	}
	return nil
}
