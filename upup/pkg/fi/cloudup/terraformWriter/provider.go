/*
Copyright 2023 The Kubernetes Authors.

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

package terraformWriter

// TerraformProvider is a provider definition for a terraform file written to cloud storage (S3, GCS, etc)
type TerraformProvider struct {
	// Name is the name of the terraform provider
	Name string
	// Arguments are additional settings used in the provider definition
	Arguments map[string]string
}
