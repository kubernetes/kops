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

package text

import (
	"bytes"
)

// SplitContentToSections splits content of a kops manifest into sections.
func SplitContentToSections(content []byte) [][]byte {

	// replace windows line endings with unix ones
	normalized := bytes.Replace(content, []byte("\r\n"), []byte("\n"), -1)

	return bytes.Split(normalized, []byte("\n---\n"))
}
