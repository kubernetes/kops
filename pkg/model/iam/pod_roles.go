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
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/util/pkg/vfs"
)

// PodRole holds a pod-specific IAM policy
type PodRole string

const (
	// PodRoleEmpty indicates we are not using a PodRole
	PodRoleEmpty PodRole = ""
	// PodRoleDNSController is the role used by the dns-controller pods
	PodRoleDNSController PodRole = "dns-controller"
)

// PodOrNodeRole is used to specify generation of permissions for either a pod or a node
type PodOrNodeRole struct {
	// NodeRole is non-empty if we are generating permissions for a node
	NodeRole kops.InstanceGroupRole
	// PodRole is non-empty if we are generating permissions for a pod
	PodRole PodRole
}

// ServiceAccountForPodRole returns the kubernetes service account used by pods with the specified role
func ServiceAccountForPodRole(podRole PodRole) types.NamespacedName {
	// We assume (for now) that the namespace is kube-system and the service account name matches the pod role
	namespace := "kube-system"
	name := string(podRole)

	return types.NamespacedName{Namespace: namespace, Name: name}
}

// ServiceAccountIssuer determines the issuer in the ServiceAccount JWTs
func ServiceAccountIssuer(clusterName string, clusterSpec *kops.ClusterSpec) (string, error) {
	if clusterSpec.Discovery == nil || clusterSpec.Discovery.Base == "" {
		return "", fmt.Errorf("must specify discovery.base with UsePodIAM")
	}

	if clusterName == "" {
		return "", fmt.Errorf("cluster name not specified")
	}

	base, err := vfs.Context.BuildVfsPath(clusterSpec.Discovery.Base)
	if err != nil {
		return "", fmt.Errorf("cannot parse VFS path %q: %v", clusterSpec.Discovery.Base, err)
	}

	p := base.Join("identity", clusterName)

	var publicURL string
	switch p := p.(type) {
	case *vfs.S3Path:
		publicURL = "https://" + p.Bucket() + ".s3.amazonaws.com/" + p.Key()
	default:
		return "", fmt.Errorf("unhandled scheme in %q for computing serviceAccountIssuer", p)
	}

	return publicURL, nil
}

// AddPodRole adds the appropriate mounts / env vars to enable a pod to use a pod role
func AddPodRole(context *IAMModelContext, podSpec *corev1.PodSpec, container *corev1.Container, podRole PodRole) error {
	cloudProvider := kops.CloudProviderID(context.Cluster.Spec.CloudProvider)

	switch cloudProvider {
	case kops.CloudProviderAWS:
		return addPodRoleForAWS(context, podSpec, container, podRole)
	default:
		return fmt.Errorf("Pod-level IAM is not yet supported on cloud %T", cloudProvider)
	}
}

func addPodRoleForAWS(context *IAMModelContext, podSpec *corev1.PodSpec, container *corev1.Container, podRole PodRole) error {
	roleName := context.IAMNameForPodRole(podRole)

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
