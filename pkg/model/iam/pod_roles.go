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

package iam

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/wellknownusers"
)

// ServiceAccountRole holds a pod-specific IAM policy
type ServiceAccountRole string

const (
	// ServiceAccountRoleEmpty indicates we are not using a ServiceAccountRole
	ServiceAccountRoleEmpty ServiceAccountRole = ""
	// ServiceAccountRoleDNSController is the role used by the dns-controller pods
	ServiceAccountRoleDNSController ServiceAccountRole = "dns-controller"
)

// NodeOrServiceAccountRole is a union of our IAM subjects.
type NodeOrServiceAccountRole struct {
	// NodeRole is non-empty if we are generating permissions for a node.
	NodeRole kops.InstanceGroupRole
	// ServiceAccountRole is non-empty if we are generating permissions for a pod.
	ServiceAccountRole ServiceAccountRole
}

// IsServiceAccountRole is true if this is a service-account-scoped role
func (r *NodeOrServiceAccountRole) IsServiceAccountRole() bool {
	return r.ServiceAccountRole != ServiceAccountRoleEmpty
}

// IsNodeRole is true if this is a node-scoped role
func (r *NodeOrServiceAccountRole) IsNodeRole() bool {
	return r.NodeRole != ""
}

// ServiceAccountForServiceAccountRole returns the kubernetes service account used by pods with the specified role
func ServiceAccountForServiceAccountRole(serviceAccountRole ServiceAccountRole) types.NamespacedName {
	// We assume (for now) that the namespace is kube-system and the service account name matches the service-account role
	namespace := "kube-system"
	name := string(serviceAccountRole)

	return types.NamespacedName{Namespace: namespace, Name: name}
}

// ServiceAccountIssuer determines the issuer in the ServiceAccount JWTs
func ServiceAccountIssuer(clusterName string, clusterSpec *kops.ClusterSpec) (string, error) {
	if featureflag.PublicJWKS.Enabled() {
		return "https://api." + clusterName, nil
	}

	return "", fmt.Errorf("ServiceAcccountIssuer not (currently) supported without PublicJWKS")
}

// AddServiceAccountRole adds the appropriate mounts / env vars to enable a pod to use a service-account role
func AddServiceAccountRole(context *IAMModelContext, podSpec *corev1.PodSpec, container *corev1.Container, serviceAccountRole ServiceAccountRole) error {
	cloudProvider := kops.CloudProviderID(context.Cluster.Spec.CloudProvider)

	switch cloudProvider {
	case kops.CloudProviderAWS:
		return addServiceAccountRoleForAWS(context, podSpec, container, serviceAccountRole)
	default:
		return fmt.Errorf("ServiceAccount-level IAM is not yet supported on cloud %T", cloudProvider)
	}
}

func addServiceAccountRoleForAWS(context *IAMModelContext, podSpec *corev1.PodSpec, container *corev1.Container, serviceAccountRole ServiceAccountRole) error {
	roleName := context.IAMNameForServiceAccountRole(serviceAccountRole)

	awsRoleARN := "arn:aws:iam::" + context.AWSAccountID + ":role/" + roleName
	tokenDir := "/var/run/secrets/amazonaws.com/"
	tokenName := "token"

	volume := corev1.Volume{
		Name: "token-amazonaws-com",
	}

	mode := int32(0o644)
	expiration := int64(86400)
	volume.Projected = &corev1.ProjectedVolumeSource{
		DefaultMode: &mode,
		Sources: []corev1.VolumeProjection{
			{
				ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
					Audience:          "amazonaws.com",
					ExpirationSeconds: &expiration,
					Path:              tokenName,
				},
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, volume)

	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		MountPath: tokenDir,
		Name:      volume.Name,
		ReadOnly:  true,
	})

	container.Env = append(container.Env, corev1.EnvVar{
		Name:  "AWS_ROLE_ARN",
		Value: awsRoleARN,
	})

	container.Env = append(container.Env, corev1.EnvVar{
		Name:  "AWS_WEB_IDENTITY_TOKEN_FILE",
		Value: tokenDir + tokenName,
	})

	// Set securityContext.fsGroup to enable file to be read
	// background: https://github.com/kubernetes/enhancements/pull/1598
	if podSpec.SecurityContext == nil {
		podSpec.SecurityContext = &corev1.PodSecurityContext{}
	}
	if podSpec.SecurityContext.FSGroup == nil {
		fsGroup := int64(wellknownusers.Generic)
		podSpec.SecurityContext.FSGroup = &fsGroup
	}

	return nil
}
