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

package wellknownusers

// We define some user ids that we use for non-root containers.
// We base at 10000 because some distros (COS) have pre-defined users around 1000

const (
	// Generic is the user id we use for non-privileged containers, where we don't need extra permissions
	// Used by e.g. dns-controller
	Generic = 10001

	// AWSAuthenticator is the user-id for the aws-iam-authenticator (built externally)
	AWSAuthenticator = 10000

	// KopsControllerID is the user id for kops-controller, which needs some extra permissions e.g. to write local logs
	// This should match the kops-controller DaemonSet's runAsUser
	KopsControllerID = 10011

	// KopsControllerName is the username for the kops-controller user
	KopsControllerName = "kops-controller"

	// 10012 was the user id of the kube-apiserver-healthcheck sidecar (removed in kOps 1.38).

	// KopsChannelsID is the user id for the kops-channels static pod.
	// The user owns the kubeconfig mounted into the container.
	// This should match the static pod's runAsUser.
	KopsChannelsID = 10013

	// KopsChannelsName is the username for the kops-channels user.
	KopsChannelsName = "kops-channels"
)
