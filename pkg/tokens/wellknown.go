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

package tokens

// GetKubernetesAuthTokens_Deprecated returns a list of all the API auth tokens we create.
// Use of these tokens is deprecated for > 1.6 and should be dropped at the appropriate time
func GetKubernetesAuthTokens_Deprecated() []string {
	return []string{
		"kubelet", "kube-proxy", "system:scheduler", "system:controller_manager",
		"system:logging", "system:monitoring", "system:dns", "kube", "admin",
	}
}
