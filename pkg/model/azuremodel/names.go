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

package azuremodel

import (
	"regexp"

	"k8s.io/kops/pkg/truncate"
)

// Azure user-assigned identity names must start with a letter or number, contain only alphanumerics, hyphens, and underscores,
// and are limited to 128 characters. See:
// https://learn.microsoft.com/en-us/entra/identity/managed-identities-azure-resources/managed-identity-best-practice-recommendations
const maxUserAssignedManagedIdentityNameLength = 128

var userAssignedManagedIdentityNameInvalidCharacters = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

func sanitizeUserAssignedManagedIdentityName(name string) string {
	sanitized := userAssignedManagedIdentityNameInvalidCharacters.ReplaceAllString(name, "-")
	return truncate.TruncateString(sanitized, truncate.TruncateStringOptions{
		MaxLength: maxUserAssignedManagedIdentityNameLength,
	})
}

func (c *AzureModelContext) NameForUserAssignedManagedIdentityControlPlane() string {
	return sanitizeUserAssignedManagedIdentityName(c.ClusterName())
}
