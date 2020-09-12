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

package azure

import (
	"k8s.io/kops/upup/pkg/fi"
)

// AzureAPITarget is a target whose purpose is to provide access AzureCloud.
type AzureAPITarget struct {
	Cloud AzureCloud
}

var _ fi.Target = &AzureAPITarget{}

// NewAzureAPITarget returns a new AzureAPITarget.
func NewAzureAPITarget(cloud AzureCloud) *AzureAPITarget {
	return &AzureAPITarget{
		Cloud: cloud,
	}
}

// Finish is called by a lifecycle drive to finish the lifecycle of
// the target.
func (t *AzureAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

// ProcessDeletions returns true if we should delete resources.
func (t *AzureAPITarget) ProcessDeletions() bool {
	return true
}
