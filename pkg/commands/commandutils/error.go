/*
Copyright 2021 The Kubernetes Authors.

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

package commandutils

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CompletionError a helper function to logs and return an error from a Cobra completion function.
func CompletionError(prefix string, err error) ([]string, cobra.ShellCompDirective) {
	if err == nil {
		cobra.CompError(prefix)
	} else {
		cobra.CompError(fmt.Sprintf("%s: %v", prefix, err))
	}
	return nil, cobra.ShellCompDirectiveError | cobra.ShellCompDirectiveNoFileComp
}
