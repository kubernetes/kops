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

// Subject represents an IAM identity, to which permissions are granted.
// It is implemented by NodeRole objects and per-ServiceAccount objects.
type Subject interface {
	// BuildAWSPolicy builds the AWS permissions for the given subject.
	BuildAWSPolicy(*PolicyBuilder) (*Policy, error)

	// ServiceAccount returns the kubernetes service account used by pods with this specified role.
	// For node roles, it returns an empty NamespacedName and false.
	ServiceAccount() (types.NamespacedName, bool)
}

// NodeRoleMaster represents the role of control-plane nodes, and implements Subject.
type NodeRoleMaster struct {
}

// ServiceAccount implements Subject.
func (_ *NodeRoleMaster) ServiceAccount() (types.NamespacedName, bool) {
	return types.NamespacedName{}, false
}

// NodeRoleNode represents the role of normal ("worker") nodes, and implements Subject.
type NodeRoleNode struct {
}

// ServiceAccount implements Subject.
func (_ *NodeRoleNode) ServiceAccount() (types.NamespacedName, bool) {
	return types.NamespacedName{}, false
}

// NodeRoleNode represents the role of bastion nodes, and implements Subject.
type NodeRoleBastion struct {
}

// ServiceAccount implements Subject.
func (_ *NodeRoleBastion) ServiceAccount() (types.NamespacedName, bool) {
	return types.NamespacedName{}, false
}

// BuildNodeRoleSubject returns a Subject implementation for the specified InstanceGroupRole.
func BuildNodeRoleSubject(igRole kops.InstanceGroupRole) (Subject, error) {
	switch igRole {
	case kops.InstanceGroupRoleMaster:
		return &NodeRoleMaster{}, nil

	case kops.InstanceGroupRoleNode:
		return &NodeRoleNode{}, nil

	case kops.InstanceGroupRoleBastion:
		return &NodeRoleBastion{}, nil

	default:
		return nil, fmt.Errorf("unknown instancegroup role %q", igRole)
	}
}

// ServiceAccountIssuer determines the issuer in the ServiceAccount JWTs
func ServiceAccountIssuer(clusterName string, clusterSpec *kops.ClusterSpec) (string, error) {
	if featureflag.PublicJWKS.Enabled() {
		return "https://api." + clusterName, nil
	}

	return "", fmt.Errorf("ServiceAcccountIssuer not (currently) supported without PublicJWKS")
}

// AddServiceAccountRole adds the appropriate mounts / env vars to enable a pod to use a service-account role
func AddServiceAccountRole(context *IAMModelContext, podSpec *corev1.PodSpec, container *corev1.Container, serviceAccountRole Subject) error {
	cloudProvider := kops.CloudProviderID(context.Cluster.Spec.CloudProvider)

	switch cloudProvider {
	case kops.CloudProviderAWS:
		return addServiceAccountRoleForAWS(context, podSpec, container, serviceAccountRole)
	default:
		return fmt.Errorf("ServiceAccount-level IAM is not yet supported on cloud %T", cloudProvider)
	}
}

func addServiceAccountRoleForAWS(context *IAMModelContext, podSpec *corev1.PodSpec, container *corev1.Container, serviceAccountRole Subject) error {
	roleName, err := context.IAMNameForServiceAccountRole(serviceAccountRole)
	if err != nil {
		return err
	}

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
