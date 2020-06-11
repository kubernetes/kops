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

package etcdmanager

import (
	corev1 "k8s.io/api/core/v1"
)

func addPodIAM(pod *corev1.Pod) error {
	//+export AWS_ROLE_ARN=arn:aws:iam::745864996096:role/k8s-fed-cloud
	//+export AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/tokens/cloud-token

	if len(pod.Spec.Containers) != 1 {
		return fmt.Errorf("expected exactly one etcd-manager container, got %d", len(pod.Spec.Containers))
	}

	container := pod.Spec.Containers[0]

	awsRoleARN := "arn:aws:iam::" + awsAccountID + ":role/k8s-fed-cloud"
	tokenPath := "/var/run/secrets/aws.amazon.com/serviceaccount"

	container.Env = append(container.Env, corev1.EnvVar{
		Name:  "AWS_ROLE_ARN",
		Value: awsRoleARN,
	})

	container.Env = append(container.Env, corev1.EnvVar{
		Name:  "AWS_WEB_IDENTITY_TOKEN_FILE",
		Value: tokenPath,
	})

	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		MountPath: tokenPath,
		Name: "token-aws-amazon-com",
		ReadOnly: true,
	})

	volume := &container.Volume{
		Name: "token-aws-amazon-com",
		Projected: corev1.ProjectedVolumeSource{
			DefaultMode: 0o644,
			Sources: []corev1.VolumeProjection{
				ServiceAccountToken: corev1.ServiceAccountTokenProjection{
					Audience: "aws.amazon.com",
					ExpirationSeconds: 86400,
					Path: "token",
				}
			}
		}
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
}
