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

package cloudup

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// Target is the type of target we are operating against.
type Target string

const (
	// TargetDirect means we will apply the changes directly to the cloud.
	TargetDirect Target = "direct"
	// TargetDryRun means we will not apply the changes but will print what would have been done.
	TargetDryRun Target = "dryrun"
	// TargetTerraform means we will generate terraform code.
	TargetTerraform Target = "terraform"
)

// Target can be used as a flag value.
var _ pflag.Value = (*Target)(nil)

func (t *Target) String() string {
	return string(*t)
}

func (t *Target) Set(value string) error {
	switch strings.ToLower(value) {
	case string(TargetDirect), string(TargetDryRun), string(TargetTerraform):
		*t = Target(value)
		return nil
	default:
		return fmt.Errorf("invalid target: %q", value)
	}
}

func (t *Target) Type() string {
	return "target"
}
