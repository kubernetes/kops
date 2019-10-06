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
	// GenericID is the user id we use for non-privileged containers, where we don't need extra permissions
	// Used by e.g. dns-controller, kops-controller
	GenericID = 10001

	// AWSAuthenticatorID is the user-id for the aws-iam-authenticator (built externally)
	AWSAuthenticatorID = 10000

	// AWSAuthenticatorName is the name for the aws-iam-authenticator user
	AWSAuthenticatorName = "aws-iam-authenticator"

	// KopsControllerID is the user id for kops-controller, which needs some extra permissions e.g. to write local logs
	// This should match the user in cmd/kops-controller/BUILD.bazel
	KopsControllerID = 10011

	// KopsControllerName is the username for the kops-controller user
	KopsControllerName = "kops-controller"
)
