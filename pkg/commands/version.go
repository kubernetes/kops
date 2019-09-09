/*
Copyright 2019 The Kubernetes Authors.

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

package commands

import (
	"fmt"
	"io"

	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
)

type VersionOptions struct {
	Short bool
}

// RunVersion implements the version command logic
func RunVersion(f *util.Factory, out io.Writer, options *VersionOptions) error {
	var s string
	if options.Short {
		s = kops.Version
	} else {
		s = "Version " + kops.Version
		if kops.GitVersion != "" {
			s += " (git-" + kops.GitVersion + ")"
		}
	}

	_, err := fmt.Fprintf(out, "%s\n", s)
	return err
}
