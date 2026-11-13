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

package azure

import (
	"fmt"
	"path"
	"strings"
)

// StagingContainer is a preexisting container with anonymous blob read access that holds the
// build artifacts of the e2e jobs.
const StagingContainer = "azureblob://stkopsstatestore/kops-e2e-staging"

// SplitBlobURL splits an azureblob://<account>/<container>/<prefix> location into the storage
// account, the container and the blob prefix, without leading or trailing slashes.
func SplitBlobURL(location string) (account, container, prefix string, err error) {
	rest, ok := strings.CutPrefix(location, "azureblob://")
	if !ok {
		return "", "", "", fmt.Errorf("%q is not an azureblob:// location", location)
	}
	account, rest, _ = strings.Cut(rest, "/")
	container, prefix, _ = strings.Cut(rest, "/")
	if account == "" || container == "" {
		return "", "", "", fmt.Errorf("%q is not of the form azureblob://<account>/<container>/<prefix>", location)
	}

	return account, container, strings.Trim(prefix, "/"), nil
}

// PublicObjectURL returns the https:// URL that serves urlPath from a container with anonymous
// read access. Unlike S3, Azure keeps a literal "+" in a URL path.
func PublicObjectURL(account, container, urlPath string) string {
	return "https://" + account + ".blob.core.windows.net" + path.Join("/", container, urlPath)
}
