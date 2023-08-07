/*
Copyright 2023 The Kubernetes Authors.

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

package helpers

import (
	"path"
	"testing"
)

func Test_cacheFilePath(t *testing.T) {
	inputs := []struct {
		kopsStateStore string
		clusterName    string
		wantFilename   string
	}{
		{
			kopsStateStore: "s3://abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk",
			clusterName: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcde." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.com",
			wantFilename: "abcdefghijklmnopqrstuvwxyzabcdef_2gdSmQxAE3nMLFrXXuEuCBoVaajWndsK8OMnUV",
		},
	}

	for _, input := range inputs {
		output := cacheFilePath(input.kopsStateStore, input.clusterName)
		_, file := path.Split(output)

		// Explicitly guard against overly long filenames - this is why we do truncation
		if len(file) > 71 {
			t.Errorf("cacheFilePath() got %v, too long(%v)", output, len(file))
		}

		if file != input.wantFilename {
			t.Errorf("cacheFilePath(%q, %q) got %v, want %v", input.kopsStateStore, input.clusterName, file, input.wantFilename)
		}
	}
}
