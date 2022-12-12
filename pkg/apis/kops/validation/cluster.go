/*
Copyright 2017 The Kubernetes Authors.

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

package validation

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func ValidateClusterUpdate(ctx context.Context, obj *kops.Cluster, status *kops.ClusterStatus, old *kops.Cluster) field.ErrorList {
	allErrs := ValidateCluster(ctx, obj, false)

	// Validate etcd cluster changes
	{
		newClusters := make(map[string]kops.EtcdClusterSpec)
		for _, etcdCluster := range obj.Spec.EtcdClusters {
			newClusters[etcdCluster.Name] = etcdCluster
		}
		oldClusters := make(map[string]kops.EtcdClusterSpec)
		for _, etcdCluster := range old.Spec.EtcdClusters {
			oldClusters[etcdCluster.Name] = etcdCluster
		}

		for k, newCluster := range newClusters {
			fp := field.NewPath("spec", "etcdClusters").Key(k)

			if oldCluster, ok := oldClusters[k]; ok {
				allErrs = append(allErrs, validateEtcdClusterUpdate(fp, newCluster, status, oldCluster)...)
			}
		}
		for k := range oldClusters {
			if _, ok := newClusters[k]; !ok {
				fp := field.NewPath("spec", "etcdClusters").Key(k)
				allErrs = append(allErrs, field.Forbidden(fp, "EtcdClusters cannot be removed"))
			}
		}
	}

	allErrs = append(allErrs, validateClusterCloudLabels(obj, field.NewPath("spec", "cloudLabels"))...)

	return allErrs
}

func validateEtcdClusterUpdate(fp *field.Path, obj kops.EtcdClusterSpec, status *kops.ClusterStatus, old kops.EtcdClusterSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj.Name != old.Name {
		allErrs = append(allErrs, field.Forbidden(fp.Child("name"), "name cannot be changed"))
	}

	var etcdClusterStatus *kops.EtcdClusterStatus
	if status != nil {
		for i := range status.EtcdClusters {
			etcdCluster := &status.EtcdClusters[i]
			if etcdCluster.Name == obj.Name {
				etcdClusterStatus = etcdCluster
			}
		}
	}

	// If the etcd cluster has been created (i.e. if we have status) then we can't support some changes
	if etcdClusterStatus != nil {
		newMembers := make(map[string]kops.EtcdMemberSpec)
		for _, member := range obj.Members {
			newMembers[member.Name] = member
		}
		oldMembers := make(map[string]kops.EtcdMemberSpec)
		for _, member := range old.Members {
			oldMembers[member.Name] = member
		}

		for k, newMember := range newMembers {
			fp := fp.Child("etcdMembers").Key(k)

			if oldMember, ok := oldMembers[k]; ok {
				allErrs = append(allErrs, validateEtcdMemberUpdate(fp, newMember, oldMember)...)
			}
		}
	}

	return allErrs
}

func validateEtcdMemberUpdate(fp *field.Path, obj kops.EtcdMemberSpec, old kops.EtcdMemberSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj.Name != old.Name {
		allErrs = append(allErrs, field.Forbidden(fp.Child("name"), "name cannot be changed"))
	}

	if fi.ValueOf(obj.InstanceGroup) != fi.ValueOf(old.InstanceGroup) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("instanceGroup"), "instanceGroup cannot be changed"))
	}

	return allErrs
}

func validateClusterCloudLabels(cluster *kops.Cluster, fldPath *field.Path) (allErrs field.ErrorList) {
	labels := cluster.Spec.CloudLabels
	return validateCloudLabels(labels, fldPath)
}

func validateCloudLabels(labels map[string]string, fldPath *field.Path) (allErrs field.ErrorList) {
	if labels == nil {
		return allErrs
	}

	reservedKeys := []string{
		"Name",
		"KubernetesCluster",
	}

	for _, reservedKey := range reservedKeys {
		_, hasKey := labels[reservedKey]
		if hasKey {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child(reservedKey), fmt.Sprintf("%q is a reserved label and cannot be used as a custom label", reservedKey)))
		}
	}
	reservedPrefixes := []string{
		"kubernetes.io/cluster/",
		"k8s.io/role/",
		"kops.k8s.io/",
	}

	for _, reservedPrefix := range reservedPrefixes {
		for label := range labels {
			if strings.HasPrefix(label, reservedPrefix) {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child(label), fmt.Sprintf("%q is a reserved label prefix and cannot be used as a custom label", reservedPrefix)))
			}
		}
	}

	return allErrs
}
