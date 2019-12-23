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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func ValidateClusterUpdate(obj *kops.Cluster, status *kops.ClusterStatus, old *kops.Cluster) field.ErrorList {
	allErrs := field.ErrorList{}

	if err := ValidateCluster(obj, false); err != nil {
		allErrs = append(allErrs, err)
	}

	// Validate etcd cluster changes
	{
		newClusters := make(map[string]*kops.EtcdClusterSpec)
		for _, etcdCluster := range obj.Spec.EtcdClusters {
			newClusters[etcdCluster.Name] = etcdCluster
		}
		oldClusters := make(map[string]*kops.EtcdClusterSpec)
		for _, etcdCluster := range old.Spec.EtcdClusters {
			oldClusters[etcdCluster.Name] = etcdCluster
		}

		for k, newCluster := range newClusters {
			fp := field.NewPath("spec", "etcdClusters").Key(k)

			oldCluster := oldClusters[k]
			if oldCluster != nil {
				allErrs = append(allErrs, validateEtcdClusterUpdate(fp, newCluster, status, oldCluster)...)
			}
		}
		for k := range oldClusters {
			newCluster := newClusters[k]
			if newCluster == nil {
				fp := field.NewPath("spec", "etcdClusters").Key(k)
				allErrs = append(allErrs, field.Forbidden(fp, "EtcdClusters cannot be removed"))
			}
		}
	}

	return allErrs
}

func validateEtcdClusterUpdate(fp *field.Path, obj *kops.EtcdClusterSpec, status *kops.ClusterStatus, old *kops.EtcdClusterSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj.Name != old.Name {
		allErrs = append(allErrs, field.Forbidden(fp.Child("Name"), "Name cannot be changed"))
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
		newMembers := make(map[string]*kops.EtcdMemberSpec)
		for _, member := range obj.Members {
			newMembers[member.Name] = member
		}
		oldMembers := make(map[string]*kops.EtcdMemberSpec)
		for _, member := range old.Members {
			oldMembers[member.Name] = member
		}

		for k, newMember := range newMembers {
			fp := fp.Child("Members").Key(k)

			oldMember := oldMembers[k]
			if oldMember != nil {
				allErrs = append(allErrs, validateEtcdMemberUpdate(fp, newMember, etcdClusterStatus, oldMember)...)
			}
		}
		for k := range oldMembers {
			newCluster := newMembers[k]
			if newCluster == nil {
				fp := fp.Child("Members").Key(k)
				allErrs = append(allErrs, field.Forbidden(fp, "EtcdCluster members cannot be removed"))
			}
		}
	}

	return allErrs
}

func validateEtcdMemberUpdate(fp *field.Path, obj *kops.EtcdMemberSpec, status *kops.EtcdClusterStatus, old *kops.EtcdMemberSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	if obj.Name != old.Name {
		allErrs = append(allErrs, field.Forbidden(fp.Child("Name"), "Name cannot be changed"))
	}

	if fi.StringValue(obj.InstanceGroup) != fi.StringValue(old.InstanceGroup) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("InstanceGroup"), "InstanceGroup cannot be changed"))
	}

	if fi.StringValue(obj.VolumeType) != fi.StringValue(old.VolumeType) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("VolumeType"), "VolumeType cannot be changed"))
	}

	if fi.Int32Value(obj.VolumeIops) != fi.Int32Value(old.VolumeIops) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("VolumeIops"), "VolumeIops cannot be changed"))
	}

	if fi.Int32Value(obj.VolumeSize) != fi.Int32Value(old.VolumeSize) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("VolumeSize"), "VolumeSize cannot be changed"))
	}

	if fi.StringValue(obj.KmsKeyId) != fi.StringValue(old.KmsKeyId) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("KmsKeyId"), "KmsKeyId cannot be changed"))
	}

	if fi.BoolValue(obj.EncryptedVolume) != fi.BoolValue(old.EncryptedVolume) {
		allErrs = append(allErrs, field.Forbidden(fp.Child("EncryptedVolume"), "EncryptedVolume cannot be changed"))
	}

	return allErrs
}
