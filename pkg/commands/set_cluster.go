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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
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

		// For now we have hard-code the values we want to support; we'll get test coverage and then do this properly...
		switch kv[0] {
		case "spec.kubelet.authorizationMode":
			cluster.Spec.Kubelet.AuthorizationMode = kv[1]
		case "spec.kubelet.authenticationTokenWebhook":
			v, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf("unknown boolean value: %q", kv[1])
			}
			cluster.Spec.Kubelet.AuthenticationTokenWebhook = &v
		case "cluster.spec.nodePortAccess":
			cluster.Spec.NodePortAccess = append(cluster.Spec.NodePortAccess, kv[1])
		case "spec.kubernetesVersion":
			cluster.Spec.KubernetesVersion = kv[1]
		case "spec.masterPublicName":
			cluster.Spec.MasterPublicName = kv[1]
		case "spec.kubeDNS.provider":
			if cluster.Spec.KubeDNS == nil {
				cluster.Spec.KubeDNS = &api.KubeDNSConfig{}
			}
			cluster.Spec.KubeDNS.Provider = kv[1]
		case "cluster.spec.etcdClusters[*].enableEtcdTLS":
			v, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf("unknown boolean value: %q", kv[1])
			}
			for _, c := range cluster.Spec.EtcdClusters {
				c.EnableEtcdTLS = v
			}
		case "cluster.spec.etcdClusters[*].enableTLSAuth":
			v, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf("unknown boolean value: %q", kv[1])
			}
			for _, c := range cluster.Spec.EtcdClusters {
				c.EnableTLSAuth = v
			}
		case "cluster.spec.etcdClusters[*].version":
			for _, c := range cluster.Spec.EtcdClusters {
				c.Version = kv[1]
			}
		case "cluster.spec.etcdClusters[*].provider":
			p, err := toEtcdProviderType(kv[1])
			if err != nil {
				return err
			}
			for _, etcd := range cluster.Spec.EtcdClusters {
				etcd.Provider = p
			}
		case "cluster.spec.etcdClusters[*].manager.image":
			for _, etcd := range cluster.Spec.EtcdClusters {
				if etcd.Manager == nil {
					etcd.Manager = &api.EtcdManagerSpec{}
				}
				etcd.Manager.Image = kv[1]
			}
		default:
			return fmt.Errorf("unhandled field: %q", field)
		}
	}
	return nil
}

func toEtcdProviderType(in string) (api.EtcdProviderType, error) {
	s := strings.ToLower(in)
	switch s {
	case "legacy":
		return api.EtcdProviderTypeLegacy, nil
	case "manager":
		return api.EtcdProviderTypeManager, nil
	default:
		return api.EtcdProviderTypeManager, fmt.Errorf("unknown etcd provider type %q", in)
	}
}
