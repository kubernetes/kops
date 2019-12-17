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

package systemd

import (
	"strings"
)

// UnitExtensions is a slice containing all valid systemd unit file extensions.
// See https://www.freedesktop.org/software/systemd/man/systemd.unit.html
var UnitExtensions = []string{
	".automount",
	".device",
	".mount",
	".path",
	".scope",
	".service",
	".slice",
	".socket",
	".swap",
	".target",
	".timer",
}

// UnitFileExtensionValid checks whether the provided filename ends with a valid
// systemd unit file extension.
func UnitFileExtensionValid(name string) bool {
	for _, ext := range UnitExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
