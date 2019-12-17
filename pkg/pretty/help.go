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

package pretty

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
)

// Bash markdown-quotes a bash command for insertion into help text.
func Bash(s string) string {
	return fmt.Sprintf("`%s`", s)
}

// LongDesc is used for formatting help text for a commands Long Description.
// It de-dents it and trims it.
func LongDesc(s string) string {
	s = heredoc.Doc(s)
	s = strings.TrimSpace(s)
	return s
}
