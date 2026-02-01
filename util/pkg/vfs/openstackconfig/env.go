/*
Copyright 2026 The Kubernetes Authors.

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

package openstackconfig

// KOPS_OS_TLS_INSECURE_SKIP_VERIFY is used to configure skipping TLS verification for OpenStack clients
// Ideally there would be a well-known OpenStack environment variable for this purpose,
// but there isn't one at present.
// Instead we create a KOPS_-specific variable.
const EnvKeyOpenstackTLSInsecureSkipVerify = "KOPS_OS_TLS_INSECURE_SKIP_VERIFY"
