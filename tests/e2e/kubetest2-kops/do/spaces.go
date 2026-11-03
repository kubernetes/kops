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

package do

const (
	// StagingBucket is a preexisting, publicly readable Spaces bucket that holds the build
	// artifacts of the e2e jobs.
	StagingBucket = "kops-e2e-staging"

	// SpacesEndpoint is the S3 API endpoint of the Spaces region of StagingBucket.
	SpacesEndpoint = "https://sfo3.digitaloceanspaces.com"
)
