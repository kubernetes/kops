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
	}{
		{
			kopsStateStore: "s3://abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk",
			clusterName: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcde." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk." +
				"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.com",
		},
	}

	output1 := cacheFilePath(inputs[0].kopsStateStore, inputs[0].clusterName)
	_, file := path.Split(output1)

	if len(file) > 71 {
		t.Errorf("cacheFilePath() got %v, too long(%v)", output1, len(file))
	}
}
